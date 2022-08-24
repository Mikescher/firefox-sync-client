package syncclient

import (
	"bytes"
	"crypto/dsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joomcode/errorx"
	"io"
	"math"
	"net/http"
	"time"
)

type FxAClient struct {
	authURL string
	client  http.Client
}

func NewFxAClient(serverurl string) *FxAClient {
	return &FxAClient{
		authURL: serverurl,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f FxAClient) Login(ctx *cli.FFSContext, email string, password string, serviceName string) (LoginSession, error) {
	stretchpwd := stretchPassword(email, password)

	ctx.PrintVerboseKV("StretchPW", stretchpwd)

	authPW, err := deriveKey(stretchpwd, "authPW", 32)
	if err != nil {
		return LoginSession{}, errorx.Decorate(err, "failed to derive key")
	}

	ctx.PrintVerboseKV("AuthPW", authPW)

	body := loginRequestSchema{
		Email:  email,
		AuthPW: hex.EncodeToString(authPW), //lowercase
		Reason: "login",
	}

	bytesBody, err := json.Marshal(body)
	if err != nil {
		return LoginSession{}, errorx.Decorate(err, "failed to marshal body")
	}

	requestURL := f.authURL + "/account/login?keys=true"

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(bytesBody))
	if err != nil {
		return LoginSession{}, errorx.Decorate(err, "failed to create request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Mobile; Firefox Accounts; rv:1.0) firefox-sync-client/"+consts.FFSCLIENT_VERSION+"golang/1.19")
	req.Header.Add("Accept", "*/*")

	ctx.PrintVerbose("Request session from " + requestURL)

	rawResp, err := f.client.Do(req)
	if err != nil {
		return LoginSession{}, errorx.Decorate(err, "failed to do request")
	}

	respBodyRaw, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return LoginSession{}, errorx.Decorate(err, "failed to read response-body request")
	}

	//TODO statuscode [429, 500, 503] means retry-after

	if rawResp.StatusCode != 200 {
		return LoginSession{}, errorx.InternalError.New(fmt.Sprintf("call to /login returned statuscode %v\n\n%v", rawResp.StatusCode, string(respBodyRaw)))
	}

	ctx.PrintVerbose(fmt.Sprintf("Request returned statuscode %d", rawResp.StatusCode))
	ctx.PrintVerbose(fmt.Sprintf("Request returned body:\n%v", string(respBodyRaw)))

	var resp loginResponseSchema
	err = json.Unmarshal(respBodyRaw, &resp)
	if err != nil {
		return LoginSession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(respBodyRaw))
	}

	if !resp.Verified {
		return LoginSession{}, errorx.InternalError.New("You must verify the login attempt (e.g. per e-mail) before continuing")
	}

	kft, err := hex.DecodeString(resp.KeyFetchToken)
	if err != nil {
		return LoginSession{}, errorx.Decorate(err, "failed to read KeyFetchToken: "+resp.KeyFetchToken)
	}

	st, err := hex.DecodeString(resp.SessionToken)
	if err != nil {
		return LoginSession{}, errorx.Decorate(err, "failed to read SessionToken: "+resp.SessionToken)
	}

	ctx.PrintVerboseKV("UserID", resp.UserID)
	ctx.PrintVerboseKV("SessionToken", st)
	ctx.PrintVerboseKV("KeyFetchToken", kft)

	return LoginSession{
		URL:             f.authURL,
		Email:           email,
		StretchPassword: stretchpwd,
		UserId:          resp.UserID,
		SessionToken:    st,
		KeyFetchToken:   kft,
	}, nil
}

func (f FxAClient) FetchKeys(ctx *cli.FFSContext, session LoginSession) ([]byte, []byte, error) {

	ctx.PrintVerbose("Request keys from " + "/account/keys")

	binResp, hawkBundleKey, err := f.requestWithHawkToken(ctx, "GET", "/account/keys", nil, session.KeyFetchToken, "keyFetchToken")
	if err != nil {
		return nil, nil, errorx.Decorate(err, "Failed to query account keys")
	}

	var resp keysResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	ctx.PrintVerboseKV("Bundle", resp.Bundle)

	bundle, err := hex.DecodeString(resp.Bundle)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to decode Bundle: "+resp.Bundle)
	}

	keys, err := unbundle("account/keys", hawkBundleKey, bundle)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to unbundle")
	}

	ctx.PrintVerboseKV("Keys<unbundled>", keys)

	unwrapKey, err := deriveKey(session.StretchPassword, "unwrapBkey", 32)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to derive-key")
	}

	ctx.PrintVerboseKV("Keys<unwrapped>", unwrapKey)

	kLow := keys[:32]
	kHigh := keys[32:]

	keyA := kLow
	keyB, err := langext.BytesXOR(kHigh, unwrapKey)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to xor key-b")
	}

	return keyA, keyB, nil
}

func (f FxAClient) GetCollectionsInfo(ctx *cli.FFSContext, session FFSyncSession) ([]models.CollectionInfo, error) {
	binResp, err := f.request(ctx, session, "GET", "/info/collections", nil)
	if err != nil {
		return nil, errorx.Decorate(err, "API request failed")
	}

	var resp collectionsInfoResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	result := make([]models.CollectionInfo, 0, len(resp))
	for k, v := range resp {
		sec, dec := math.Modf(v)
		result = append(result, models.CollectionInfo{
			Name:         k,
			LastModified: time.Unix(int64(sec), int64(dec*(1e9))),
		})
	}

	return result, nil
}

func (f FxAClient) GetCollectionsCounts(ctx *cli.FFSContext, session FFSyncSession) ([]models.CollectionCount, error) {
	binResp, err := f.request(ctx, session, "GET", "/info/collection_counts", nil)
	if err != nil {
		return nil, errorx.Decorate(err, "API request failed")
	}

	var resp collectionsCountResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	result := make([]models.CollectionCount, 0, len(resp))
	for k, v := range resp {
		result = append(result, models.CollectionCount{
			Name:  k,
			Count: v,
		})
	}

	return result, nil
}

func (f FxAClient) GetCollectionsUsage(ctx *cli.FFSContext, session FFSyncSession) ([]models.CollectionUsage, error) {
	binResp, err := f.request(ctx, session, "GET", "/info/collection_usage", nil)
	if err != nil {
		return nil, errorx.Decorate(err, "API request failed")
	}

	var resp collectionsUsageResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	result := make([]models.CollectionUsage, 0, len(resp))
	for k, v := range resp {
		result = append(result, models.CollectionUsage{
			Name:  k,
			Usage: int64(v * 1024),
		})
	}

	return result, nil
}

func (f FxAClient) AssertBrowserID(ctx *cli.FFSContext, session KeyedSession) (BrowserIdSession, error) {
	ctx.PrintVerbose("Create & Sign Certificate")

	bid, t0, dur, err := f.getBrowserIDAssertion(ctx, session)
	if err != nil {
		return BrowserIdSession{}, errorx.Decorate(err, "Failed to assert BID")
	}

	ctx.PrintVerboseKV("BID-Assertion", bid)

	return session.Extend(bid, t0, dur), nil
}

func (f FxAClient) HawkAuth(ctx *cli.FFSContext, session BrowserIdSession) (HawkSession, error) {
	ctx.PrintVerbose("Authenticate HAWK")

	sha := sha256.New()
	sha.Write(session.KeyB)
	sessionState := hex.EncodeToString(sha.Sum(nil)[0:16])

	ctx.PrintVerboseKV("Session-State", sessionState)

	cred, err := f.getHawkCredentials(ctx, session.BrowserID, sessionState)
	if err != nil {
		return HawkSession{}, errorx.Decorate(err, "Failed to get hawk credentials")
	}

	return session.Extend(cred), nil
}

func (f FxAClient) GetCryptoKeys(ctx *cli.FFSContext, session HawkSession) (CryptoSession, error) {
	ctx.PrintVerbose("Get crypto/keys from storage")

	syncKeys, err := keyBundleFromMasterKey(session.KeyB, "identity.mozilla.com/picl/v1/oldsync")
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "Failed to generate syncKeys")
	}

	ctx.PrintVerboseKV("syncKeys.EncryptionKey", syncKeys.EncryptionKey)
	ctx.PrintVerboseKV("syncKeys.HMACKey", syncKeys.HMACKey)

	binResp, err := f.request(ctx, session.ToKeylessSession(), "GET", "/storage/crypto/keys", nil)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "API request failed")
	}

	var resp getRecordSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	ctx.PrintVerboseKV("record.ID", resp.ID)
	ctx.PrintVerboseKV("record.Modified", resp.Modified)
	ctx.PrintVerboseKV("record.Payload", resp.Payload)

	var payload payloadSchema
	err = json.Unmarshal([]byte(resp.Payload), &payload)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to unmarshal payload:\n"+resp.Payload)
	}

	ctx.PrintVerboseKV("payload.IV", payload.IV)
	ctx.PrintVerboseKV("payload.HMAC", payload.HMAC)
	ctx.PrintVerboseKV("payload.Ciphertext", payload.Ciphertext)

	ciphertext, err := base64.StdEncoding.DecodeString(payload.Ciphertext)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to b64-decode payload.ciphertext")
	}

	iv, err := base64.StdEncoding.DecodeString(payload.IV)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to b64-decode payload.iv")
	}

	hmac, err := hex.DecodeString(payload.HMAC)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to hex-decode payload.hmac")
	}

	dplBin, err := decryptPayload(payload.Ciphertext, ciphertext, iv, hmac, syncKeys)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to decrypt payload")
	}

	ctx.PrintVerboseKV("payload<decrypted>", string(dplBin))

	var cryptoKeys cryptoKeysSchema
	err = json.Unmarshal(dplBin, &cryptoKeys)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to unmarshal cryptoKeys:\n"+resp.Payload)
	}

	result := CryptoKeys{Keys: make(map[string]KeyBundle, len(cryptoKeys.Collections)+1)}

	result.Keys[""], err = keyBundleFromB64Array(cryptoKeys.Default)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to hex-decode cryptokeys.default")
	}

	for k, v := range cryptoKeys.Collections {
		result.Keys[k], err = keyBundleFromB64Array(v)
		if err != nil {
			return CryptoSession{}, errorx.Decorate(err, "failed to hex-decode cryptokeys.default")
		}
	}

	ctx.PrintVerboseKV("bulkKeys.0", cryptoKeys.Default[0])
	ctx.PrintVerboseKV("bulkKeys.1", cryptoKeys.Default[1])
	for k, v := range cryptoKeys.Collections {
		ctx.PrintVerboseKV("bulkKeys."+k+".0", v[0])
		ctx.PrintVerboseKV("bulkKeys."+k+".1", v[1])
	}

	return session.Extend(result), nil
}

func (f FxAClient) getBrowserIDAssertion(ctx *cli.FFSContext, session KeyedSession) (string, time.Time, time.Duration, error) {

	duration := time.Second * consts.DefaultBIDAssertionDuration

	params := dsa.Parameters{}
	err := dsa.GenerateParameters(&params, rand.Reader, dsa.L1024N160)
	if err != nil {
		return "", time.Time{}, 0, errorx.Decorate(err, "Failed to generate DSA params")
	}

	var privateKey dsa.PrivateKey
	privateKey.PublicKey.Parameters = params

	err = dsa.GenerateKey(&privateKey, rand.Reader)
	if err != nil {
		return "", time.Time{}, 0, errorx.Decorate(err, "Failed to generate DSA key-pair")
	}

	body := signCertRequestSchema{
		PublicKey: signCertRequestSchemaPKey{
			Algorithm: "DS",
			P:         privateKey.P.Text(16),
			Q:         privateKey.Q.Text(16),
			G:         privateKey.G.Text(16),
			Y:         privateKey.Y.Text(16),
		},
		Duration: duration.Milliseconds(),
	}

	ctx.PrintVerbose("Sign new certificate via " + "/certificate/sign")

	binResp, _, err := f.requestWithHawkToken(ctx, "POST", "/certificate/sign", body, session.SessionToken, "sessionToken")
	if err != nil {
		return "", time.Time{}, 0, errorx.Decorate(err, "Failed to sign cert")
	}

	var resp signCertResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return "", time.Time{}, 0, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	ctx.PrintVerboseKV("Certificate", resp.Certificate)

	t0 := time.Now()
	exp := t0.UnixMilli() + duration.Milliseconds()

	token := jwt.NewWithClaims(&SigningMethodDS128{}, jwt.MapClaims{
		"exp": exp,
		"aud": ctx.Opt.TokenServerURL,
	})

	assertion, err := token.SignedString(&privateKey)
	if err != nil {
		return "", time.Time{}, 0, errorx.Decorate(err, "failed to generate JWT")
	}

	ctx.PrintVerboseKV("Assertion:JWT", assertion)

	return resp.Certificate + "~" + assertion, t0, duration, nil
}

func (f FxAClient) getHawkCredentials(ctx *cli.FFSContext, bid string, clientState string) (HawkCredentials, error) {
	auth := "BrowserID " + bid

	req, err := http.NewRequestWithContext(ctx, "GET", ctx.Opt.TokenServerURL+"/1.0/sync/1.5", nil)
	if err != nil {
		return HawkCredentials{}, errorx.Decorate(err, "failed to create request")
	}
	req.Header.Add("Authorization", auth)
	req.Header.Add("X-Client-State", clientState)

	ctx.PrintVerbose("Query HAWK credentials")

	rawResp, err := f.client.Do(req)
	if err != nil {
		return HawkCredentials{}, errorx.Decorate(err, "failed to do request")
	}

	respBodyRaw, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return HawkCredentials{}, errorx.Decorate(err, "failed to read response-body request")
	}

	if rawResp.StatusCode != 200 {
		return HawkCredentials{}, errorx.InternalError.New(fmt.Sprintf("api call returned statuscode %v\n\n%v", rawResp.StatusCode, string(respBodyRaw)))
	}

	var resp hawkCredResponseSchema
	err = json.Unmarshal(respBodyRaw, &resp)
	if err != nil {
		return HawkCredentials{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(respBodyRaw))
	}

	ctx.PrintVerboseKV("HAWK-ID", resp.ID)
	ctx.PrintVerboseKV("HAWK-Key", resp.Key)
	ctx.PrintVerboseKV("HAWK-UserID", resp.UID)
	ctx.PrintVerboseKV("HAWK-Endpoint", resp.APIEndpoint)
	ctx.PrintVerboseKV("HAWK-Duration", resp.Duration)
	ctx.PrintVerboseKV("HAWK-HashAlgo", resp.HashAlgorithm)
	ctx.PrintVerboseKV("HAWK-FxA-Uid", resp.HashedFxAUID)
	ctx.PrintVerboseKV("HAWK-NodeType", resp.NodeType)

	return HawkCredentials{
		HawkID:            resp.ID,
		HawkKey:           resp.Key,
		APIEndpoint:       resp.APIEndpoint,
		HawkDuration:      resp.Duration,
		HawkHashAlgorithm: resp.HashAlgorithm,
	}, nil
}

func (f FxAClient) requestWithHawkToken(ctx *cli.FFSContext, method string, relurl string, body any, token []byte, tokenType string) ([]byte, []byte, error) {
	requestURL := f.authURL + relurl

	var outBundleKey []byte

	auth := func(method string, url string, body string, contentType string) (string, error) {
		ctx.PrintVerbose(fmt.Sprintf("Calculate HAWK-Auth-Token (token request)"))
		hawkAuth, hawkBundleKey, err := calcHawkTokenAuth(token, tokenType, method, url, body)
		if err != nil {
			return "", errorx.Decorate(err, "failed to create hawk-auth")
		}
		outBundleKey = hawkBundleKey
		return hawkAuth, nil
	}

	res, err := f.internalRequest(ctx, auth, method, requestURL, body)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "Request failed")
	}

	return res, outBundleKey, nil
}

func (f FxAClient) request(ctx *cli.FFSContext, session FFSyncSession, method string, relurl string, body any) ([]byte, error) {
	requestURL := session.APIEndpoint + relurl

	auth := func(method string, url string, body string, contentType string) (string, error) {
		ctx.PrintVerbose(fmt.Sprintf("Calculate HAWK-Auth-Token (normal request)"))
		hawkAuth, err := calcHawkSessionAuth(session, method, url, body, contentType)
		if err != nil {
			return "", errorx.Decorate(err, "failed to create hawk-auth")
		}
		return hawkAuth, nil
	}

	res, err := f.internalRequest(ctx, auth, method, requestURL, body)
	if err != nil {
		return nil, errorx.Decorate(err, "Request failed")
	}

	return res, nil
}

func (f FxAClient) internalRequest(ctx *cli.FFSContext, auth func(method string, url string, body string, contentType string) (string, error), method string, requestURL string, body any) ([]byte, error) {
	strBody := ""
	var bodyReader io.Reader = nil
	if body != nil {
		bytesBody, err := json.Marshal(body)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to marshal body")
		}
		strBody = string(bytesBody)
		bodyReader = bytes.NewReader(bytesBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to create request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "firefox-sync-client/"+consts.FFSCLIENT_VERSION)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Host", req.URL.Host)

	hawkAuth, err := auth(req.Method, req.URL.String(), strBody, "application/json")
	if err != nil {
		return nil, errorx.Decorate(err, "failed to create auth")
	}

	req.Header.Add("Authorization", hawkAuth)

	ctx.PrintVerboseKV("Authorization", hawkAuth)

	rawResp, err := f.client.Do(req)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to do request")
	}

	respBodyRaw, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to read response-body request")
	}

	//TODO statuscode [429, 500, 503] means retry-after

	if rawResp.StatusCode != 200 {
		if len(string(respBodyRaw)) > 1 {
			return nil, errorx.InternalError.New(fmt.Sprintf("call to %v returned statuscode %v\nBody:\n%v", requestURL, rawResp.StatusCode, string(respBodyRaw)))
		} else {
			return nil, errorx.InternalError.New(fmt.Sprintf("call to %v returned statuscode %v", requestURL, rawResp.StatusCode))
		}
	}

	ctx.PrintVerbose(fmt.Sprintf("Request returned statuscode %d", rawResp.StatusCode))
	ctx.PrintVerbose(fmt.Sprintf("Request returned body:\n%v", string(respBodyRaw)))

	return respBodyRaw, nil
}
