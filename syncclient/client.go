package syncclient

import (
	"bytes"
	"crypto/dsa"
	"crypto/rand"
	"crypto/sha256"
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
	"net/http"
	"strconv"
	"strings"
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

func (f FxAClient) Login(ctx *cli.FFSContext, email string, password string) (LoginSession, error) {
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
		if len(string(respBodyRaw)) > 1 {
			return LoginSession{}, errorx.InternalError.New(fmt.Sprintf("call to /login returned statuscode %v\nBody:\n%v", rawResp.StatusCode, string(respBodyRaw)))
		} else {
			return LoginSession{}, errorx.InternalError.New(fmt.Sprintf("call to /login returned statuscode %v", rawResp.StatusCode))
		}
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
		StretchPassword: stretchpwd,
		UserId:          resp.UserID,
		SessionToken:    st,
		KeyFetchToken:   kft,
	}, nil
}

func (f FxAClient) RegisterDevice(ctx *cli.FFSContext, session LoginSession, deviceName string) error {

	ctx.PrintVerbose("Register device-name '" + deviceName + "'")

	body := registerDeviceRequestSchema{
		Name: deviceName,
		Type: "cli",
	}

	_, _, err := f.requestWithHawkToken(ctx, "POST", "/account/device", body, session.SessionToken, "sessionToken")
	if err != nil {
		return errorx.Decorate(err, "Failed to register device")
	}

	ctx.PrintVerbose("Device registered as '" + deviceName + "'")

	return nil
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

func (f FxAClient) AssertBrowserID(ctx *cli.FFSContext, session KeyedSession) (BrowserIdSession, error) {
	ctx.PrintVerbose("Create & Sign Certificate")

	duration := time.Second * consts.DefaultBIDAssertionDuration

	params := dsa.Parameters{}
	err := dsa.GenerateParameters(&params, rand.Reader, dsa.L1024N160)
	if err != nil {
		return BrowserIdSession{}, errorx.Decorate(err, "Failed to generate DSA params")
	}

	var privateKey dsa.PrivateKey
	privateKey.PublicKey.Parameters = params

	err = dsa.GenerateKey(&privateKey, rand.Reader)
	if err != nil {
		return BrowserIdSession{}, errorx.Decorate(err, "Failed to generate DSA key-pair")
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
		return BrowserIdSession{}, errorx.Decorate(err, "Failed to sign cert")
	}

	var resp signCertResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return BrowserIdSession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
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
		return BrowserIdSession{}, errorx.Decorate(err, "failed to generate JWT")
	}

	ctx.PrintVerboseKV("Assertion:JWT", assertion)

	bid := resp.Certificate + "~" + assertion

	ctx.PrintVerboseKV("BID-Assertion", bid)

	return session.Extend(bid, t0, duration), nil
}

func (f FxAClient) HawkAuth(ctx *cli.FFSContext, session BrowserIdSession) (HawkSession, error) {
	ctx.PrintVerbose("Authenticate HAWK")

	sha := sha256.New()
	sha.Write(session.KeyB)
	sessionState := hex.EncodeToString(sha.Sum(nil)[0:16])

	ctx.PrintVerboseKV("Session-State", sessionState)

	auth := "BrowserID " + session.BrowserID

	req, err := http.NewRequestWithContext(ctx, "GET", ctx.Opt.TokenServerURL+"/1.0/sync/1.5", nil)
	if err != nil {
		return HawkSession{}, errorx.Decorate(err, "failed to create request")
	}

	req.Header.Add("Authorization", auth)
	req.Header.Add("X-Client-State", sessionState)

	ctx.PrintVerbose("Query HAWK credentials")

	t0 := time.Now()

	rawResp, err := f.client.Do(req)
	if err != nil {
		return HawkSession{}, errorx.Decorate(err, "failed to do request")
	}

	respBodyRaw, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return HawkSession{}, errorx.Decorate(err, "failed to read response-body request")
	}

	if rawResp.StatusCode != 200 {
		if len(string(respBodyRaw)) > 1 {
			return HawkSession{}, errorx.InternalError.New(fmt.Sprintf("api call returned statuscode %v\nBody:\n%v", rawResp.StatusCode, string(respBodyRaw)))
		} else {
			return HawkSession{}, errorx.InternalError.New(fmt.Sprintf("api call returned statuscode %v", rawResp.StatusCode))
		}
	}

	var resp hawkCredResponseSchema
	err = json.Unmarshal(respBodyRaw, &resp)
	if err != nil {
		return HawkSession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(respBodyRaw))
	}

	hawkTimeOut := t0.Add(time.Second * time.Duration(resp.Duration))

	ctx.PrintVerboseKV("HAWK-ID", resp.ID)
	ctx.PrintVerboseKV("HAWK-Key", resp.Key)
	ctx.PrintVerboseKV("HAWK-UserID", resp.UID)
	ctx.PrintVerboseKV("HAWK-Endpoint", resp.APIEndpoint)
	ctx.PrintVerboseKV("HAWK-Duration", resp.Duration)
	ctx.PrintVerboseKV("HAWK-HashAlgo", resp.HashAlgorithm)
	ctx.PrintVerboseKV("HAWK-FxA-Uid", resp.HashedFxAUID)
	ctx.PrintVerboseKV("HAWK-NodeType", resp.NodeType)
	ctx.PrintVerboseKV("HAWK-Timeout", hawkTimeOut)

	if resp.HashAlgorithm != "sha256" {
		return HawkSession{}, errorx.InternalError.New("HAWK-HashAlgorithm '" + resp.HashAlgorithm + "' is currently not supported")
	}

	cred := HawkCredentials{
		HawkID:            resp.ID,
		HawkKey:           resp.Key,
		APIEndpoint:       resp.APIEndpoint,
		HawkHashAlgorithm: resp.HashAlgorithm,
	}

	return session.Extend(cred, hawkTimeOut), nil
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
	ctx.PrintVerboseKV("record.Modified", langext.UnixFloatSeconds(resp.Modified))
	ctx.PrintVerboseKV("record.Payload", resp.Payload)

	var payload payloadSchema
	err = json.Unmarshal([]byte(resp.Payload), &payload)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to unmarshal payload:\n"+resp.Payload)
	}

	ctx.PrintVerboseKV("payload.IV", payload.IV)
	ctx.PrintVerboseKV("payload.HMAC", payload.HMAC)
	ctx.PrintVerboseKV("payload.Ciphertext", payload.Ciphertext)

	dplBin, err := decryptPayload(payload.Ciphertext, payload.IV, payload.HMAC, syncKeys)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to decrypt payload")
	}

	ctx.PrintVerboseKV("payload<decrypted>", string(dplBin))

	var cryptoKeys cryptoKeysSchema
	err = json.Unmarshal(dplBin, &cryptoKeys)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to unmarshal cryptoKeys:\n"+resp.Payload)
	}

	result := make(map[string]KeyBundle, len(cryptoKeys.Collections)+1)

	result[""], err = keyBundleFromB64Array(cryptoKeys.Default)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to hex-decode cryptokeys.default")
	}

	for k, v := range cryptoKeys.Collections {
		result[k], err = keyBundleFromB64Array(v)
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

func (f FxAClient) AutoRefreshSession(ctx *cli.FFSContext, session FFSyncSession) (FFSyncSession, error) {
	session, changed, err := f.RefreshSession(ctx, session, ctx.Opt.ForceRefreshSession)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to refresh session")
	}

	if changed && ctx.Opt.SaveRefreshedSession {

		ctx.PrintVerbose("Saved new session after auto-update")

		ctx.PrintVerbose("Save session to " + ctx.Opt.SessionFilePath)

		cfp, err := ctx.AbsSessionFilePath()
		if err != nil {
			return FFSyncSession{}, errorx.Decorate(err, "failed to get session file path")
		}

		err = session.Save(cfp)
		if err != nil {
			return FFSyncSession{}, errorx.Decorate(err, "failed to save session")
		}

	}

	return session, nil
}

func (f FxAClient) RefreshSession(ctx *cli.FFSContext, session FFSyncSession, force bool) (FFSyncSession, bool, error) {

	if session.Expired() {
		ctx.PrintVerbose("Saved session is expired (valid until " + session.Timeout.In(ctx.Opt.TimeZone).Format(time.RFC3339) + ")")
		ctx.PrintVerbose("Refreshing session (AssertBrowserID + HawkAuth)")
	} else if force {
		ctx.PrintVerbose("Saved session is not expired (valid until " + session.Timeout.In(ctx.Opt.TimeZone).Format(time.RFC3339) + ")")
		ctx.PrintVerbose("Refreshing session by force (AssertBrowserID + HawkAuth)")
	} else {
		ctx.PrintVerbose("Saved session is valid (valid until " + session.Timeout.In(ctx.Opt.TimeZone).Format(time.RFC3339) + ")")
		return session, false, nil
	}

	sessionBID, err := f.AssertBrowserID(ctx, session.ToKeyed())
	if err != nil {
		return FFSyncSession{}, false, errorx.Decorate(err, "failed to assert BID")
	}

	sessionHawk, err := f.HawkAuth(ctx, sessionBID)
	if err != nil {
		return FFSyncSession{}, false, errorx.Decorate(err, "failed to authenticate HAWK")
	}

	sessionCrypto, err := f.GetCryptoKeys(ctx, sessionHawk)
	if err != nil {
		return FFSyncSession{}, false, errorx.Decorate(err, "failed to get crypto/keys")
	}

	sessionSync := sessionCrypto.Reduce()

	return sessionSync, true, nil
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
		result = append(result, models.CollectionInfo{
			Name:         k,
			LastModified: langext.UnixFloatSeconds(v),
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

func (f FxAClient) GetQuota(ctx *cli.FFSContext, session FFSyncSession) (int64, *int64, error) {
	binResp, err := f.request(ctx, session, "GET", "/info/quota", nil)
	if err != nil {
		return 0, nil, errorx.Decorate(err, "API request failed")
	}

	var resp []any
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return 0, nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	if len(resp) != 2 {
		return 0, nil, errorx.InternalError.New("info/quota returned invali data (array.len)")
	}

	ctx.PrintVerboseKV("quota[0]", resp[0])
	ctx.PrintVerboseKV("quota[1]", resp[1])

	var used int64

	switch v := resp[0].(type) {
	case float64:
		used = int64(v * 1024)
	default:
		return 0, nil, errorx.InternalError.New("info/quota returned invali data (array[0].type)")
	}

	var total *int64 = nil
	if resp[1] != nil {
		switch v := resp[1].(type) {
		case float64:
			total = langext.Ptr(int64(v * 1024))
		default:
			return 0, nil, errorx.InternalError.New("info/quota returned invali data (array[1].type)")
		}
	}

	return used, total, nil
}

func (f FxAClient) ListRecords(ctx *cli.FFSContext, session FFSyncSession, collection string, after *time.Time, sort *string, idOnly bool, decode bool, limit *int, offset *int) ([]models.DecodedRecord, error) {
	url := fmt.Sprintf("/storage/%s", collection)

	params := make([]string, 0, 8)

	if after != nil {
		params = append(params, "newer="+strconv.FormatInt(after.Unix(), 10))
	}
	if sort != nil {
		params = append(params, "sort="+*sort)
	}
	if !idOnly {
		params = append(params, "full=true")
	}
	if limit != nil {
		params = append(params, "limit="+strconv.Itoa(*limit))
	}
	if offset != nil {
		params = append(params, "offset="+strconv.Itoa(*offset))
	}

	if len(params) > 0 {
		url = url + "?" + strings.Join(params, "&")
	}

	binResp, err := f.request(ctx, session, "GET", url, nil)
	if err != nil {
		return nil, errorx.Decorate(err, "API request failed")
	}

	if idOnly {
		var resp listRecordsIDsResponseSchema
		err = json.Unmarshal(binResp, &resp)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
		}

		result := make([]models.DecodedRecord, 0, len(binResp))

		for _, v := range resp {
			result = append(result, models.DecodedRecord{ID: v})
		}
		return result, nil
	}

	var resp listRecordsResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	ctx.PrintVerbose(fmt.Sprintf("API Call returned %d records", len(resp)))

	result := make([]models.DecodedRecord, 0, len(binResp))

	for _, v := range resp {
		result = append(result, models.DecodedRecord{
			ID:       v.ID,
			Payload:  v.Payload,
			Modified: langext.UnixFloatSeconds(v.Modified),
		})
	}

	if decode {
		bulkKeys := session.BulkKeys[""]

		if v, ok := session.BulkKeys[collection]; ok {
			ctx.PrintVerbose("Use collection-specific bulk-keys")

			bulkKeys = v
		} else {
			ctx.PrintVerbose("Use global bulk-keys")
		}
		ctx.PrintVerboseKV("EncryptionKey", bulkKeys.EncryptionKey)
		ctx.PrintVerboseKV("HMACKey", bulkKeys.HMACKey)

		for i, v := range result {

			var payload payloadSchema
			err = json.Unmarshal([]byte(v.Payload), &payload)
			if err != nil {
				return nil, errorx.Decorate(err, "failed to unmarshal payload of record <"+v.ID+">:\n"+v.Payload)
			}

			ctx.PrintVerbose("Decrypt payload of " + v.ID)

			dplBin, err := decryptPayload(payload.Ciphertext, payload.IV, payload.HMAC, bulkKeys)
			if err != nil {
				return nil, errorx.Decorate(err, "failed to decrypt payload of record <"+v.ID+">")
			}

			result[i].DecodedData = dplBin
		}
	}

	return result, nil
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
	for k, va := range rawResp.Header {
		for _, v := range va {
			ctx.PrintVerbose(fmt.Sprintf("Request returned Header [%s] := '%s'", k, v))
		}
	}
	ctx.PrintVerbose(fmt.Sprintf("Request returned body:\n%s", string(respBodyRaw)))

	return respBodyRaw, nil
}
