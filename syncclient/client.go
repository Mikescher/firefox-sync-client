package syncclient

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"fmt"
	"github.com/joomcode/errorx"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"git.blackforestbytes.com/BlackForestBytes/goext/timeext"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type FxAClient struct {
	authURL string
	client  http.Client
}

func NewFxAClient(ctx *cli.FFSContext, serverurl string) *FxAClient {
	c := http.Client{
		Timeout: ctx.Opt.RequestTimeout,
	}

	if ctx.Opt.RequestX509Ignore {

		// Standard values from http.DefaultTransport ( except t.TLSClientConfig )

		d := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}

		t := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           d.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		}

		c.Transport = t
	}

	return &FxAClient{
		authURL: serverurl,
		client:  c,
	}
}

func (f FxAClient) Login(ctx *cli.FFSContext, email string, password string) (LoginSession, SessionVerification, error) {
	resp, stretchpwd, err := f.makeLoginRequest(ctx, email, password, stretchPassword(email, password), false)
	if err != nil {
		return LoginSession{}, "", err
	}

	kft, err := hex.DecodeString(resp.KeyFetchToken)
	if err != nil {
		return LoginSession{}, "", errorx.Decorate(err, "failed to read KeyFetchToken: "+resp.KeyFetchToken)
	}

	st, err := hex.DecodeString(resp.SessionToken)
	if err != nil {
		return LoginSession{}, "", errorx.Decorate(err, "failed to read SessionToken: "+resp.SessionToken)
	}

	ctx.PrintVerboseKV("UserID", resp.UserID)
	ctx.PrintVerboseKV("SessionToken", st)
	ctx.PrintVerboseKV("KeyFetchToken", kft)

	if !resp.Verified {
		if resp.VerificationMethod == "totp-2fa" {
			ctx.PrintVerbose("OTP verification will be required in the next step (totp-2fa)")

			return LoginSession{
				Mail:            email,
				StretchPassword: stretchpwd,
				UserId:          resp.UserID,
				SessionToken:    st,
				KeyFetchToken:   kft,
			}, VerificationTOTP2FA, nil
		}

		if resp.VerificationMethod == "email" {
			return LoginSession{}, "", fferr.DirectOutput.New("You must verify the login attempt (per e-mail) before continuing")
		}

		if resp.VerificationMethod == "email-otp" {
			ctx.PrintVerbose("OTP verification will be required in the next step (email-otp)")

			// is this the same as 2fa ?, can we simply use the same code and input the "otp" code from the mail in /session/verifiy/totp  ??
			return LoginSession{}, "", fferr.DirectOutput.New("You must verify the login attempt (e.g. per e-mail) before continuing")
		}

		if resp.VerificationMethod == "email-2fa" {
			ctx.PrintVerbose("2FA verification will be required in the next step (email-2fa)")

			return LoginSession{
				Mail:            email,
				StretchPassword: stretchpwd,
				UserId:          resp.UserID,
				SessionToken:    st,
				KeyFetchToken:   kft,
			}, VerificationMail2FA, nil
		}

		if resp.VerificationMethod == "email-captcha" {
			ctx.PrintVerbose("Captcha verification will be required in the next step (email-captcha)")

			// is this the same as 2fa ?, can we simply use the same code and input the "captcha" code from the mail in /session/verifiy/totp  ??
			return LoginSession{}, "", fferr.DirectOutput.New("Your account was issued a captcha, please solve the captcha mail to your e-mail address first")
		}

		return LoginSession{}, "", errorx.InternalError.New(fmt.Sprintf("The requested verification method '%s' is unknown", resp.VerificationMethod))
	}

	return LoginSession{
		Mail:            email,
		StretchPassword: stretchpwd,
		UserId:          resp.UserID,
		SessionToken:    st,
		KeyFetchToken:   kft,
	}, VerificationNone, nil
}

func (f FxAClient) makeLoginRequest(ctx *cli.FFSContext, email string, password string, stretchpwd []byte, is120Retry bool) (loginResponseSchema, []byte, error) {
	ctx.PrintVerboseKV("StretchPW", stretchpwd)

	authPW, err := deriveKey(stretchpwd, "authPW", 32)
	if err != nil {
		return loginResponseSchema{}, nil, errorx.Decorate(err, "failed to derive key")
	}

	ctx.PrintVerboseKV("AuthPW", authPW)

	body := loginRequestSchema{
		Email:  email,
		AuthPW: hex.EncodeToString(authPW), //lowercase
		Reason: "login",
	}

	bytesBody, err := json.Marshal(body)
	if err != nil {
		return loginResponseSchema{}, nil, errorx.Decorate(err, "failed to marshal body")
	}

	requestURL := f.authURL + "/account/login?keys=true"

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(bytesBody))
	if err != nil {
		return loginResponseSchema{}, nil, errorx.Decorate(err, "failed to create request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Mobile; Firefox Accounts; rv:1.0) firefox-sync-client/"+consts.FFSCLIENT_VERSION+"golang/1.19")
	req.Header.Add("Accept", "*/*")

	ctx.PrintVerbose("Request session from " + requestURL)

	ctx.PrintVerbose(fmt.Sprintf("Do HTTP Request [%s]::%s", "POST", requestURL))

	rawResp, err := f.doRequestWithRetries(ctx, req, 1)
	if err != nil {
		return loginResponseSchema{}, nil, errorx.Decorate(err, "failed to do request")
	}

	respBodyRaw, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return loginResponseSchema{}, nil, errorx.Decorate(err, "failed to read response-body request")
	}

	if rawResp.StatusCode != 200 {
		var errResp loginErrorResponseSchema
		err = json.Unmarshal(respBodyRaw, &errResp)
		if err != nil {
			return loginResponseSchema{}, nil, errorx.Decorate(err, "failed to unmarshal error:\n"+string(respBodyRaw))
		}

		// If the email used to stretch the password is different from sync server, the server throws a 400 error
		// with message "Incorrect email case". The response json contains the correct email for stretching the password
		if rawResp.StatusCode == 400 && errResp.ErrNo == 120 && !is120Retry {
			ctx.PrintVerbose("Using " + errResp.Email + " for stretch password and retrying login")
			return f.makeLoginRequest(ctx, email, password, stretchPassword(errResp.Email, password), true)
		}

		if len(string(respBodyRaw)) > 1 {
			return loginResponseSchema{}, nil, errorx.InternalError.New(fmt.Sprintf("call to /login returned statuscode %v\nBody:\n%v", rawResp.StatusCode, string(respBodyRaw)))
		} else {
			return loginResponseSchema{}, nil, errorx.InternalError.New(fmt.Sprintf("call to /login returned statuscode %v", rawResp.StatusCode))
		}
	}

	ctx.PrintVerbose(fmt.Sprintf("Request returned statuscode %d", rawResp.StatusCode))
	ctx.PrintVerbose(fmt.Sprintf("Request returned body:\n%v", string(respBodyRaw)))

	var resp loginResponseSchema
	err = json.Unmarshal(respBodyRaw, &resp)
	if err != nil {
		return loginResponseSchema{}, nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(respBodyRaw))
	}
	return resp, stretchpwd, nil
}

func (f FxAClient) VerifyWithOTP(ctx *cli.FFSContext, session LoginSession, otp string) error {
	body := totpVerifyRequestSchema{
		Code:    otp,
		Service: "login",
	}
	binResp, _, err := f.requestWithHawkToken(ctx, "POST", "/session/verify/totp", body, session.SessionToken, "sessionToken")
	if err != nil {
		return errorx.Decorate(err, "Failed to verify session with OTP")
	}

	var resp totpVerifyResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	if !resp.Success {
		return fferr.DirectOutput.New(fmt.Sprintf("OTP '%s' was not accepted by the server", otp))
	}

	ctx.PrintVerbose("Session verified")

	return nil
}

func (f FxAClient) RegisterDevice(ctx *cli.FFSContext, session LoginSession, deviceName string, deviceType string) error {

	ctx.PrintVerbose("Register device-name '" + deviceName + "'")

	body := registerDeviceRequestSchema{
		Name: deviceName,
		Type: deviceType,
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

func (f FxAClient) AcquireOAuthToken(ctx *cli.FFSContext, session KeyedSession) (OAuthSession, error) {

	ctx.PrintVerbose("Create OAuth Token")

	t0 := time.Now()

	oAuthBody := oauthTokenRequestSchema{
		GrantType:  "fxa-credentials",
		AccessType: "offline",
		ClientID:   consts.OAuthClientID,
		Scope:      consts.OAuthScope,
	}

	binRespOAuth, _, err := f.requestWithHawkToken(ctx, "POST", "/oauth/token", oAuthBody, session.SessionToken, "sessionToken")
	if err != nil {
		return OAuthSession{}, errorx.Decorate(err, "Failed to request oauth-token")
	}

	var respOAuth oauthTokenResponseSchema
	err = json.Unmarshal(binRespOAuth, &respOAuth)
	if err != nil {
		return OAuthSession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binRespOAuth))
	}

	accTokenDuration := timeext.FromSeconds(respOAuth.ExpiresIn)

	ctx.PrintVerboseKV("AccessToken", respOAuth.AccessToken)
	ctx.PrintVerboseKV("RefreshToken", respOAuth.RefreshToken)
	ctx.PrintVerboseKV("Expiration", accTokenDuration)

	ctx.PrintVerbose("Query ScopedKeyData")

	keyDataBody := scopedKeyDataRequestSchema{
		ClientID: consts.OAuthClientID,
		Scope:    consts.OAuthScope,
	}

	binRespScopedKeyData, _, err := f.requestWithHawkToken(ctx, "POST", "/account/scoped-key-data", keyDataBody, session.SessionToken, "sessionToken")
	if err != nil {
		return OAuthSession{}, errorx.Decorate(err, "Failed to request scoped-key-data")
	}

	var respKeyData scopedKeyDataResponseSchema
	err = json.Unmarshal(binRespScopedKeyData, &respKeyData)
	if err != nil {
		return OAuthSession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binRespScopedKeyData))
	}

	data, ok := respKeyData[consts.OAuthScope]
	if !ok {
		return OAuthSession{}, errorx.InternalError.New("scoped-key-data does not contain scope")
	}

	ctx.PrintVerboseKV("KeyRotationTimestamp", data.KeyRotationTimestamp)

	clientStateBin := sha256.Sum256(session.KeyB)
	clientStateB64 := base64.RawURLEncoding.EncodeToString(clientStateBin[0:16])
	keyID := fmt.Sprintf("%d-%s", data.KeyRotationTimestamp, clientStateB64)

	ctx.PrintVerboseKV("ClientState", clientStateB64)
	ctx.PrintVerboseKV("KeyID", keyID)

	return session.Extend(respOAuth.AccessToken, respOAuth.RefreshToken, keyID, t0, accTokenDuration), nil
}

func (f FxAClient) RefreshOAuthToken(ctx *cli.FFSContext, session KeyedSession, refreshToken string) (OAuthSession, error) {

	ctx.PrintVerbose("Create OAuth Token (via refreshToken)")

	ctx.PrintVerboseKV("RefreshToken", refreshToken)

	t0 := time.Now()

	oAuthBody := oauthTokenRequestSchema{
		GrantType:    "fxa-credentials",
		RefreshToken: refreshToken,
		ClientID:     consts.OAuthClientID,
		Scope:        consts.OAuthScope,
	}

	binRespOAuth, _, err := f.requestWithHawkToken(ctx, "POST", "/oauth/token", oAuthBody, session.SessionToken, "sessionToken")
	if err != nil {
		return OAuthSession{}, errorx.Decorate(err, "Failed to request oauth-token")
	}

	var respOAuth oauthTokenResponseSchema
	err = json.Unmarshal(binRespOAuth, &respOAuth)
	if err != nil {
		return OAuthSession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binRespOAuth))
	}

	accTokenDuration := timeext.FromSeconds(respOAuth.ExpiresIn)

	ctx.PrintVerboseKV("AccessToken", respOAuth.AccessToken)
	ctx.PrintVerboseKV("RefreshToken", refreshToken)
	ctx.PrintVerboseKV("Expiration", accTokenDuration)

	ctx.PrintVerbose("Query ScopedKeyData")

	keyDataBody := scopedKeyDataRequestSchema{
		ClientID: consts.OAuthClientID,
		Scope:    consts.OAuthScope,
	}

	binRespScopedKeyData, _, err := f.requestWithHawkToken(ctx, "POST", "/account/scoped-key-data", keyDataBody, session.SessionToken, "sessionToken")
	if err != nil {
		return OAuthSession{}, errorx.Decorate(err, "Failed to request scoped-key-data")
	}

	var respKeyData scopedKeyDataResponseSchema
	err = json.Unmarshal(binRespScopedKeyData, &respKeyData)
	if err != nil {
		return OAuthSession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binRespScopedKeyData))
	}

	data, ok := respKeyData[consts.OAuthScope]
	if !ok {
		return OAuthSession{}, errorx.InternalError.New("scoped-key-data does not contain scope")
	}

	ctx.PrintVerboseKV("KeyRotationTimestamp", data.KeyRotationTimestamp)

	clientStateBin := sha256.Sum256(session.KeyB)
	clientStateB64 := base64.RawURLEncoding.EncodeToString(clientStateBin[0:16])
	keyID := fmt.Sprintf("%d-%s", data.KeyRotationTimestamp, clientStateB64)

	ctx.PrintVerboseKV("ClientState", clientStateB64)
	ctx.PrintVerboseKV("KeyID", keyID)

	return session.Extend(respOAuth.AccessToken, refreshToken, keyID, t0, accTokenDuration), nil
}

func (f FxAClient) HawkAuth(ctx *cli.FFSContext, session OAuthSession) (HawkSession, error) {
	ctx.PrintVerbose("Authenticate HAWK")

	sha := sha256.New()
	sha.Write(session.KeyB)
	sessionState := hex.EncodeToString(sha.Sum(nil)[0:16])

	ctx.PrintVerboseKV("Session-State", sessionState)
	ctx.PrintVerboseKV("AccessToken", session.AccessToken)
	ctx.PrintVerboseKV("KeyID", session.KeyID)

	req, err := http.NewRequestWithContext(ctx, "GET", ctx.Opt.TokenServerURL+"/1.0/sync/1.5", nil)
	if err != nil {
		return HawkSession{}, errorx.Decorate(err, "failed to create request")
	}

	req.Header.Add("Authorization", "Bearer "+session.AccessToken)
	req.Header.Add("X-KeyID", session.KeyID)

	ctx.PrintVerbose("Query HAWK credentials")

	t0 := time.Now()

	rawResp, err := f.doRequestWithRetries(ctx, req, 1)
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

	binResp, err := f.request(ctx, session.ToKeylessSession(), "GET", fmt.Sprintf("/storage/%s/%s", consts.CollectionCrypto, consts.RecordCryptoKeys), nil)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "API request failed")
	}

	var resp recordsResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	ctx.PrintVerboseKV("record.ID", resp.ID)
	ctx.PrintVerboseKV("record.Modified", timeext.UnixFloatSeconds(resp.Modified))
	ctx.PrintVerboseKV("record.Payload", resp.Payload)

	var payload payloadSchema
	err = json.Unmarshal([]byte(resp.Payload), &payload)
	if err != nil {
		return CryptoSession{}, errorx.Decorate(err, "failed to unmarshal payload:\n"+resp.Payload)
	}

	ctx.PrintVerboseKV("payload.IV", payload.IV)
	ctx.PrintVerboseKV("payload.HMAC", payload.HMAC)
	ctx.PrintVerboseKV("payload.Ciphertext", payload.Ciphertext)

	dplBin, err := decryptPayload(ctx, payload.Ciphertext, payload.IV, payload.HMAC, syncKeys)
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

func (f FxAClient) RefreshSession(ctx *cli.FFSContext, session FFSyncSession, force bool) (FFSyncSession, bool, error) {

	if session.Expired() {
		ctx.PrintVerbose("Saved session is expired (valid until " + session.Timeout.In(ctx.Opt.TimeZone).Format(time.RFC3339) + ")")
		ctx.PrintVerbose("Refreshing session (OAuth via refreshToken + HawkAuth)")
	} else if force {
		ctx.PrintVerbose("Saved session is not expired (valid until " + session.Timeout.In(ctx.Opt.TimeZone).Format(time.RFC3339) + ")")
		ctx.PrintVerbose("Refreshing session by force (OAuth via refreshToken + HawkAuth)")
	} else {
		ctx.PrintVerbose("Saved session is valid (valid until " + session.Timeout.In(ctx.Opt.TimeZone).Format(time.RFC3339) + ")")
		return session, false, nil
	}

	sessionOAuth, err := f.RefreshOAuthToken(ctx, session.ToKeyed(), session.RefreshToken)
	if err != nil {
		return FFSyncSession{}, false, errorx.Decorate(err, "failed to refresh OAuth")
	}

	sessionHawk, err := f.HawkAuth(ctx, sessionOAuth)
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
			LastModified: timeext.UnixFloatSeconds(v),
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
		return 0, nil, errorx.InternalError.New("info/quota returned invalid data (array.len)")
	}

	ctx.PrintVerboseKV("quota[0]", resp[0])
	ctx.PrintVerboseKV("quota[1]", resp[1])

	var used int64

	switch v := resp[0].(type) {
	case float64:
		used = int64(v * 1024)
	default:
		return 0, nil, errorx.InternalError.New("info/quota returned invalid data (array[0].type)")
	}

	var total *int64 = nil
	if resp[1] != nil {
		switch v := resp[1].(type) {
		case float64:
			total = langext.Ptr(int64(v * 1024))
		default:
			return 0, nil, errorx.InternalError.New("info/quota returned invalid data (array[1].type)")
		}
	}

	return used, total, nil
}

func (f FxAClient) ListRecords(ctx *cli.FFSContext, session FFSyncSession, collection string, after *time.Time, sort *string, idOnly bool, decode bool, limit *int, offset *int) ([]models.Record, error) {
	requrl := fmt.Sprintf("/storage/%s", url.PathEscape(collection))

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
		requrl = requrl + "?" + strings.Join(params, "&")
	}

	binResp, err := f.request(ctx, session, "GET", requrl, nil)
	if err != nil {
		return nil, errorx.Decorate(err, "API request failed")
	}

	if idOnly {
		var resp listRecordsIDsResponseSchema
		err = json.Unmarshal(binResp, &resp)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
		}

		result := make([]models.Record, 0, len(binResp))

		for _, v := range resp {
			result = append(result, models.Record{ID: v})
		}
		return result, nil
	}

	var resp listRecordsResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	ctx.PrintVerbose(fmt.Sprintf("API Call returned %d records", len(resp)))

	result := make([]models.Record, 0, len(binResp))

	for _, v := range resp {
		result = append(result, models.Record{
			ID:           v.ID,
			Payload:      v.Payload,
			SortIndex:    v.SortIndex,
			Modified:     timeext.UnixFloatSeconds(v.Modified),
			ModifiedUnix: v.Modified,
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

			ctx.PrintVerbose("Decrypting payload of " + v.ID)

			dplBin, err := decryptPayload(ctx, payload.Ciphertext, payload.IV, payload.HMAC, bulkKeys)
			if err != nil {
				return nil, errorx.Decorate(err, "failed to decrypt payload of record <"+v.ID+">")
			}

			ctx.PrintVerbose("Decrypted Payload:\n" + string(dplBin))

			result[i].DecodedData = dplBin
		}
	}

	return result, nil
}

func (f FxAClient) GetRecord(ctx *cli.FFSContext, session FFSyncSession, collection string, recordid string, decode bool) (models.Record, error) {
	binResp, err := f.request(ctx, session, "GET", fmt.Sprintf("/storage/%s/%s", url.PathEscape(collection), url.PathEscape(recordid)), nil)
	if err != nil {
		return models.Record{}, errorx.Decorate(err, "API request failed")
	}

	var resp recordsResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return models.Record{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	record := models.Record{
		ID:           resp.ID,
		RawData:      binResp,
		Modified:     timeext.UnixFloatSeconds(resp.Modified),
		ModifiedUnix: resp.Modified,
		Payload:      resp.Payload,
		SortIndex:    resp.SortIndex,
		TTL:          resp.TTL,
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

		var payload payloadSchema
		err = json.Unmarshal([]byte(record.Payload), &payload)
		if err != nil {
			return models.Record{}, errorx.Decorate(err, "failed to unmarshal payload of record:\n"+record.Payload)
		}

		ctx.PrintVerbose("Decrypting payload")

		dplBin, err := decryptPayload(ctx, payload.Ciphertext, payload.IV, payload.HMAC, bulkKeys)
		if err != nil {
			return models.Record{}, errorx.Decorate(err, "failed to decrypt payload of record")
		}

		record.DecodedData = dplBin
	}

	return record, nil
}

func (f FxAClient) RecordExists(ctx *cli.FFSContext, session FFSyncSession, collection string, recordid string) (bool, error) {
	_, err := f.request(ctx, session, "GET", fmt.Sprintf("/storage/%s/%s", url.PathEscape(collection), url.PathEscape(recordid)), nil)
	if err == nil {
		return true, nil
	}
	if errorx.IsOfType(err, fferr.Request404) {
		return false, nil
	}
	return false, errorx.Decorate(err, "API request failed")
}

func (f FxAClient) SoftDeleteRecord(ctx *cli.FFSContext, session FFSyncSession, collection string, recordid string) error {
	jsonpayload := deletedPayloadData{
		ID:      recordid,
		Deleted: true,
	}

	plainpayload, err := json.Marshal(jsonpayload)
	if err != nil {
		return err
	}

	payload, err := f.EncryptPayload(ctx, session, collection, string(plainpayload))
	if err != nil {
		return err
	}

	bso := recordsRequestSchema{Payload: langext.Ptr(payload)}

	_, err = f.request(ctx, session, "PUT", fmt.Sprintf("/storage/%s/%s", url.PathEscape(collection), url.PathEscape(recordid)), bso)
	if err != nil {
		return errorx.Decorate(err, "API request failed")
	}

	return nil
}

func (f FxAClient) DeleteRecord(ctx *cli.FFSContext, session FFSyncSession, collection string, recordid string) error {
	_, err := f.request(ctx, session, "DELETE", fmt.Sprintf("/storage/%s/%s", url.PathEscape(collection), url.PathEscape(recordid)), nil)
	if err != nil {
		return errorx.Decorate(err, "API request failed")
	}

	return nil
}

func (f FxAClient) DeleteCollection(ctx *cli.FFSContext, session FFSyncSession, collection string) error {
	_, err := f.request(ctx, session, "DELETE", fmt.Sprintf("/storage/%s", url.PathEscape(collection)), nil)
	if err != nil {
		return errorx.Decorate(err, "API request failed")
	}

	return nil
}

func (f FxAClient) DeleteAllData(ctx *cli.FFSContext, session FFSyncSession) error {
	_, err := f.request(ctx, session, "DELETE", "", nil)
	if err != nil {
		return errorx.Decorate(err, "API request failed")
	}

	return nil
}

func (f FxAClient) CheckSession(ctx *cli.FFSContext, session FFSyncSession) (bool, error) {
	binResp, _, err := f.requestWithHawkToken(ctx, "GET", "/session/status", nil, session.SessionToken, "sessionToken")
	if err != nil {
		return false, errorx.Decorate(err, "API request failed")
	}

	var resp sessionStatusResponseSchema
	err = json.Unmarshal(binResp, &resp)
	if err != nil {
		return false, errorx.Decorate(err, "failed to unmarshal response:\n"+string(binResp))
	}

	if resp.State != "verified" {
		return false, nil
	}

	if resp.UserID != session.UserId {
		return false, nil
	}

	return true, nil
}

func (f FxAClient) PutRecord(ctx *cli.FFSContext, session FFSyncSession, collection string, data models.RecordUpdate, forceCreateNew bool, forceUpdateExisting bool) error {

	if forceCreateNew {
		exists, err := f.RecordExists(ctx, session, collection, data.ID)
		if err != nil {
			return errorx.Decorate(err, "failed to check record-exists")
		}
		if exists {
			return fferr.DirectOutput.New("Cannot create record, an record with this ID already exists")
		}
	}

	if forceUpdateExisting {
		exists, err := f.RecordExists(ctx, session, collection, data.ID)
		if err != nil {
			return errorx.Decorate(err, "failed to check record-exists")
		}
		if !exists {
			return fferr.DirectOutput.New("Cannot update record, an record with this ID does not exists")
		}
	}

	bso := recordsRequestSchema{
		ID:        langext.Ptr(data.ID),
		SortIndex: data.SortIndex,
		Payload:   data.Payload,
		TTL:       data.TTL,
	}

	_, err := f.request(ctx, session, "PUT", fmt.Sprintf("/storage/%s/%s", url.PathEscape(collection), url.PathEscape(data.ID)), bso)
	if err != nil {
		return errorx.Decorate(err, "API request failed")
	}

	return nil
}

func (f FxAClient) EncryptPayload(ctx *cli.FFSContext, session FFSyncSession, collection string, rawpayload string) (string, error) {

	bulkKeys := session.BulkKeys[""]

	if v, ok := session.BulkKeys[collection]; ok {
		ctx.PrintVerbose("Use collection-specific bulk-keys")

		bulkKeys = v
	} else {
		ctx.PrintVerbose("Use global bulk-keys")
	}
	ctx.PrintVerboseKV("EncryptionKey", bulkKeys.EncryptionKey)
	ctx.PrintVerboseKV("HMACKey", bulkKeys.HMACKey)

	ctx.PrintVerbose("Encrypting payload")

	ciphertext, iv, hmac, err := encryptPayload(ctx, rawpayload, bulkKeys)
	if err != nil {
		return "", errorx.Decorate(err, "failed to decrypt payload of record")
	}

	ctx.PrintVerboseKV("Ciphertext", ciphertext)
	ctx.PrintVerboseKV("IV", iv)
	ctx.PrintVerboseKV("HMAC", hmac)

	payload := payloadSchema{
		Ciphertext: ciphertext,
		IV:         iv,
		HMAC:       hmac,
	}
	payloadbin, err := json.Marshal(payload)
	if err != nil {
		return "", errorx.Decorate(err, "failed to marshal new payload")
	}

	return string(payloadbin), nil
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

	ctx.PrintVerbose(fmt.Sprintf("Do HTTP Request [%s]::%s", req.Method, requestURL))
	if strBody != "" {
		ctx.PrintVerbose("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
		ctx.PrintVerbose(langext.TryPrettyPrintJson(strBody))
		ctx.PrintVerbose("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	}

	rawResp, err := f.doRequestWithRetries(ctx, req, 1)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to do request")
	}

	respBodyRaw, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to read response-body request")
	}

	ctx.PrintVerbose(fmt.Sprintf("Request returned statuscode %d", rawResp.StatusCode))
	for k, va := range rawResp.Header {
		for _, v := range va {
			ctx.PrintVerbose(fmt.Sprintf("Request returned Header [%s] := '%s'", k, v))
		}
	}
	ctx.PrintVerbose(fmt.Sprintf("Request returned body:\n%s", string(respBodyRaw)))

	if rawResp.StatusCode == 404 {
		if len(string(respBodyRaw)) > 1 {
			return nil, fferr.Request404.New(fmt.Sprintf("call to %v returned statuscode %v\nBody:\n%v", requestURL, rawResp.StatusCode, string(respBodyRaw)))
		} else {
			return nil, fferr.Request404.New(fmt.Sprintf("call to %v returned statuscode %v", requestURL, rawResp.StatusCode))
		}
	}

	if rawResp.StatusCode == 400 {
		if string(respBodyRaw) == "6" {
			return nil, fferr.Request400.New(fmt.Sprintf("call to %v returned statuscode %v (%s: %s)", requestURL, rawResp.StatusCode, string(respBodyRaw), "JSON parse failure, likely due to badly-formed POST data."))
		}
		if string(respBodyRaw) == "8" {
			return nil, fferr.Request400.New(fmt.Sprintf("call to %v returned statuscode %v (%s: %s)", requestURL, rawResp.StatusCode, string(respBodyRaw), "Invalid BSO, likely due to badly-formed POST data."))
		}
		if string(respBodyRaw) == "13" {
			return nil, fferr.Request400.New(fmt.Sprintf("call to %v returned statuscode %v (%s: %s)", requestURL, rawResp.StatusCode, string(respBodyRaw), "Invalid collection, likely invalid chars incollection name."))
		}
		if string(respBodyRaw) == "14" {
			return nil, fferr.Request400.New(fmt.Sprintf("call to %v returned statuscode %v (%s: %s)", requestURL, rawResp.StatusCode, string(respBodyRaw), "User has exceeded their storage quota."))
		}
		if string(respBodyRaw) == "16" {
			return nil, fferr.Request400.New(fmt.Sprintf("call to %v returned statuscode %v (%s: %s)", requestURL, rawResp.StatusCode, string(respBodyRaw), "Client is known to be incompatible with the server."))
		}
		if string(respBodyRaw) == "17" {
			return nil, fferr.Request400.New(fmt.Sprintf("call to %v returned statuscode %v (%s: %s)", requestURL, rawResp.StatusCode, string(respBodyRaw), "Server limit exceeded, likely due to too many items or too large a payload in a POST request."))
		}
		if len(string(respBodyRaw)) > 1 {
			return nil, fferr.Request400.New(fmt.Sprintf("call to %v returned statuscode %v\nBody:\n%v", requestURL, rawResp.StatusCode, string(respBodyRaw)))
		}

		return nil, fferr.Request400.New(fmt.Sprintf("call to %v returned statuscode %v", requestURL, rawResp.StatusCode))
	}

	if rawResp.StatusCode != 200 {
		if len(string(respBodyRaw)) > 1 {
			return nil, errorx.InternalError.New(fmt.Sprintf("call to %v returned statuscode %v\nBody:\n%v", requestURL, rawResp.StatusCode, string(respBodyRaw)))
		} else {
			return nil, errorx.InternalError.New(fmt.Sprintf("call to %v returned statuscode %v", requestURL, rawResp.StatusCode))
		}
	}

	return respBodyRaw, nil
}

func (f FxAClient) doRequestWithRetries(ctx *cli.FFSContext, req *http.Request, try int) (*http.Response, error) {

	ctx.PrintVerbose(fmt.Sprintf("Start HTTP call to %s [[ try %d ]]", req.URL.String(), try))

	resp, err := f.client.Do(req)
	if err != nil {
		ctx.PrintVerbose(fmt.Sprintf("HTTP call returned an error (%s)", err.Error()))

		if try <= ctx.Opt.MaxRequestRetries && strings.HasSuffix(err.Error(), "x509: certificate signed by unknown authority") {
			// not sure why or how this happens
			// but sometimes token.services.mozilla.com returns simply a wrong cert ?!?
			// could never really reproduce it and now we simply retry

			ctx.PrintVerbose(fmt.Sprintf("(x509 error) Retry request after %f sec", ctx.Opt.RequestX509RetryDelay.Seconds()))
			time.Sleep(ctx.Opt.RequestX509RetryDelay)
			return f.doRequestWithRetries(ctx, req, try+1)
		}

		return nil, err
	}

	ctx.PrintVerbose("HTTP call returned Statuscode " + strconv.Itoa(resp.StatusCode))

	if try <= ctx.Opt.MaxRequestRetries && resp.StatusCode == 429 {
		// Client has sent too many requests
		// see https://mozilla.github.io/ecosystem-platform/api#defined-errors

		ctx.PrintVerbose(fmt.Sprintf("(429 | Client has sent too many requests) Retry request after %f sec", ctx.Opt.RequestFloodControlRetryDelay.Seconds()))
		time.Sleep(ctx.Opt.RequestFloodControlRetryDelay)
		return f.doRequestWithRetries(ctx, req, try+1)
	}

	if try <= ctx.Opt.MaxRequestRetries && resp.StatusCode == 500 {
		// Internal Server Error

		ctx.PrintVerbose(fmt.Sprintf("(500 | Internal Server Error) Retry request after %f sec", ctx.Opt.RequestServerErrRetryDelay.Seconds()))
		time.Sleep(ctx.Opt.RequestServerErrRetryDelay)
		return f.doRequestWithRetries(ctx, req, try+1)
	}

	if try <= ctx.Opt.MaxRequestRetries && resp.StatusCode == 502 {
		// Bad Gateway

		ctx.PrintVerbose(fmt.Sprintf("(502 | Bad Gateway) Retry request after %f sec", ctx.Opt.RequestServerErrRetryDelay.Seconds()))
		time.Sleep(ctx.Opt.RequestServerErrRetryDelay)
		return f.doRequestWithRetries(ctx, req, try+1)
	}

	if try <= ctx.Opt.MaxRequestRetries && resp.StatusCode == 503 {
		// Service Unavailable

		ctx.PrintVerbose(fmt.Sprintf("(503 | Service Unavailable) Retry request after %f sec", ctx.Opt.RequestServerErrRetryDelay.Seconds()))
		time.Sleep(ctx.Opt.RequestServerErrRetryDelay)
		return f.doRequestWithRetries(ctx, req, try+1)
	}

	return resp, nil
}
