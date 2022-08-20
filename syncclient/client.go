package syncclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"fmt"
	"github.com/joomcode/errorx"
	"io"
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

func (f FxAClient) Login(ctx *cli.FFSContext, email string, password string) (FxASession, error) {
	stretchpwd := stretchPassword(email, password)

	ctx.PrintVerbose("StretchPW       := " + hex.EncodeToString(stretchpwd))

	authPW, err := deriveKey(stretchpwd, "authPW", 32)
	if err != nil {
		return FxASession{}, errorx.Decorate(err, "failed to derive key")
	}

	ctx.PrintVerbose("AuthPW          := " + hex.EncodeToString(authPW))

	body := loginRequestSchema{
		Email:  email,
		AuthPW: hex.EncodeToString(authPW), //lowercase
		Reason: "login",
	}

	bytesBody, err := json.Marshal(body)
	if err != nil {
		return FxASession{}, errorx.Decorate(err, "failed to marshal body")
	}

	url := f.authURL + "/account/login?keys=true"

	//TODO [unblockCode, verificationMethod] (2FA ?)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bytesBody))
	if err != nil {
		return FxASession{}, errorx.Decorate(err, "failed to create request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "firefox-sync-client/"+consts.FFSCLIENT_VERSION)
	req.Header.Add("Accept", "application/json")

	ctx.PrintVerbose("Request session from " + url)

	rawResp, err := f.client.Do(req)
	if err != nil {
		return FxASession{}, errorx.Decorate(err, "failed to do request")
	}

	respBodyRaw, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return FxASession{}, errorx.Decorate(err, "failed to read response-body request")
	}

	//TODO statuscode [429, 500, 503] means retry-after

	if rawResp.StatusCode != 200 {
		return FxASession{}, errorx.InternalError.New(fmt.Sprintf("call to /login returned statuscode %v\n\n%v", rawResp.StatusCode, string(respBodyRaw)))
	}

	ctx.PrintVerbose(fmt.Sprintf("Request returned statuscode %d", rawResp.StatusCode))
	ctx.PrintVerbose(fmt.Sprintf("Request returned body:\n%v", string(respBodyRaw)))

	var resp loginResponseSchema
	err = json.Unmarshal(respBodyRaw, &resp)
	if err != nil {
		return FxASession{}, errorx.Decorate(err, "failed to unmarshal response:\n"+string(respBodyRaw))
	}

	ctx.PrintVerbose("UserID          := " + resp.UserID)
	ctx.PrintVerbose("SessionToken    := " + resp.SessionToken)
	ctx.PrintVerbose("KeyFetchToken   := " + resp.KeyFetchToken)

	return FxASession{
		Email:             email,
		StretchPassword:   stretchpwd,
		UserId:            resp.UserID,
		SessionToken:      resp.SessionToken,
		KeyFetchToken:     resp.KeyFetchToken,
		SessionUpdateTime: time.Now(),
	}, nil
}

func (f FxAClient) FetchKeys(ctx *cli.FFSContext, session FxASession) ([]byte, []byte, error) {

	kft, err := hex.DecodeString(session.KeyFetchToken)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to decode KeyFetchToken: "+session.KeyFetchToken)
	}

	url := f.authURL + "/account/keys"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to create request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "firefox-sync-client/"+consts.FFSCLIENT_VERSION)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Host", req.URL.Host)

	ctx.PrintVerbose(fmt.Sprintf("Calculate HAWK-Auth-Token"))

	hawkAuth, hawkBundleKey, err := calcHawkTokenAuth(kft, "keyFetchToken", req.Method, req.URL.String(), "")
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to create hawk-auth")
	}

	ctx.PrintVerbose(fmt.Sprintf("HAWK-Auth-Token := %v", hawkAuth))
	ctx.PrintVerbose(fmt.Sprintf("HAWK-Bundle-Key := %v", hex.EncodeToString(hawkBundleKey)))

	req.Header.Add("Authorization", hawkAuth)

	ctx.PrintVerbose("Request keys from " + url)

	rawResp, err := f.client.Do(req)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to do request")
	}

	respBodyRaw, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to read response-body request")
	}

	//TODO statuscode [429, 500, 503] means retry-after

	if rawResp.StatusCode != 200 {
		return nil, nil, errorx.InternalError.New(fmt.Sprintf("call to /keys returned statuscode %v\n\n%v", rawResp.StatusCode, string(respBodyRaw)))
	}

	ctx.PrintVerbose(fmt.Sprintf("Request returned statuscode %d", rawResp.StatusCode))
	ctx.PrintVerbose(fmt.Sprintf("Request returned body:\n%v", string(respBodyRaw)))

	var resp keysResponseSchema
	err = json.Unmarshal(respBodyRaw, &resp)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to unmarshal response:\n"+string(respBodyRaw))
	}

	ctx.PrintVerbose(fmt.Sprintf("Bundle          := %v", resp.Bundle))

	bundle, err := hex.DecodeString(resp.Bundle)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to decode Bundle: "+resp.Bundle)
	}

	keys, err := unbundle("account/keys", hawkBundleKey, bundle)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to unbundle")
	}

	ctx.PrintVerbose(fmt.Sprintf("Keys<unbundled> := %v", hex.EncodeToString(keys)))

	unwrapKey, err := deriveKey(session.StretchPassword, "unwrapBkey", 32)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to derive-key")
	}

	ctx.PrintVerbose(fmt.Sprintf("Keys<unwrapped> := %v", hex.EncodeToString(unwrapKey)))

	kLow := keys[:32]
	kHigh := keys[32:]

	keyA := kLow
	keyB, err := langext.BytesXOR(kHigh, unwrapKey)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "failed to xor key-b")
	}

	return keyA, keyB, nil
}
