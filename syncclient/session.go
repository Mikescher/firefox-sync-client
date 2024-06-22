package syncclient

import (
	"encoding/hex"
	"encoding/json"
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
	"gogs.mikescher.com/BlackForestBytes/goext/timeext"
	"os"
	"path/filepath"
	"time"
)

type LoginSession struct {
	StretchPassword []byte
	UserId          string
	SessionToken    []byte
	KeyFetchToken   []byte
}

type SessionVerification string //@enum:type

const (
	VerificationNone    SessionVerification = "NONE"
	VerificationTOTP2FA SessionVerification = "TOTP_2FA"
	VerificationMail2FA SessionVerification = "MAIL_2FA"
)

type KeyedSession struct {
	UserId       string
	SessionToken []byte
	KeyA         []byte
	KeyB         []byte
}

type FxABrowserID struct {
	BrowserID    string
	CertTime     time.Time
	CertDuration time.Duration
}

type OAuthSession struct {
	UserId       string
	SessionToken []byte
	KeyA         []byte
	KeyB         []byte
	AccessToken  string
	RefreshToken string
	KeyID        string
	CertTime     time.Time
	CertDuration time.Duration
}

type HawkCredentials struct {
	HawkID            string
	HawkKey           string
	APIEndpoint       string
	HawkHashAlgorithm string
}

type HawkSession struct {
	UserId            string
	SessionToken      []byte
	KeyA              []byte
	KeyB              []byte
	AccessToken       string
	RefreshToken      string
	KeyID             string
	Timeout           time.Time
	HawkID            string
	HawkKey           string
	APIEndpoint       string
	HawkHashAlgorithm string
}

type CryptoSession struct {
	UserId            string
	SessionToken      []byte
	KeyA              []byte
	KeyB              []byte
	AccessToken       string
	RefreshToken      string
	KeyID             string
	Timeout           time.Time
	HawkID            string
	HawkKey           string
	APIEndpoint       string
	HawkHashAlgorithm string
	CryptoKeys        map[string]KeyBundle
}

type FFSyncSession struct {
	SessionToken      []byte
	KeyA              []byte
	KeyB              []byte
	AccessToken       string
	RefreshToken      string
	UserId            string
	APIEndpoint       string
	HawkID            string
	HawkKey           string
	HawkHashAlgorithm string
	Timeout           time.Time
	BulkKeys          map[string]KeyBundle
}

type sessionHawkJson struct {
	APIEndpoint   string              `json:"apiEndpoint"`
	ID            string              `json:"id"`
	Key           string              `json:"key"`
	HashAlgorithm string              `json:"algorithm"`
	BulkKeys      map[string][]string `json:"bulkKeys"`
}

type sessionJson struct {
	SessionToken string          `json:"sessionToken"`
	KeyA         string          `json:"keyA"`
	KeyB         string          `json:"keyB"`
	AccessToken  string          `json:"accessToken"`
	RefreshToken string          `json:"refreshToken"`
	UserId       string          `json:"userID"`
	Hawk         sessionHawkJson `json:"hawk"`
	Timeout      int64           `json:"timeout"`
}

func (s LoginSession) Extend(ka []byte, kb []byte) KeyedSession {
	return KeyedSession{
		UserId:       s.UserId,
		SessionToken: s.SessionToken,
		KeyA:         ka,
		KeyB:         kb,
	}
}

func (e KeyedSession) Extend(accToken string, refreshToken string, keyID string, t0 time.Time, dur time.Duration) OAuthSession {
	return OAuthSession{
		UserId:       e.UserId,
		SessionToken: e.SessionToken,
		KeyA:         e.KeyA,
		KeyB:         e.KeyB,
		AccessToken:  accToken,
		RefreshToken: refreshToken,
		KeyID:        keyID,
		CertTime:     t0,
		CertDuration: dur,
	}
}

func (e OAuthSession) Extend(cred HawkCredentials, hawkTimeout time.Time) HawkSession {
	return HawkSession{
		UserId:            e.UserId,
		SessionToken:      e.SessionToken,
		KeyA:              e.KeyA,
		KeyB:              e.KeyB,
		AccessToken:       e.AccessToken,
		RefreshToken:      e.RefreshToken,
		KeyID:             e.KeyID,
		Timeout:           timeext.Min(e.CertTime.Add(e.CertDuration), hawkTimeout),
		HawkID:            cred.HawkID,
		HawkKey:           cred.HawkKey,
		APIEndpoint:       cred.APIEndpoint,
		HawkHashAlgorithm: cred.HawkHashAlgorithm,
	}
}

func (e HawkSession) Extend(keys map[string]KeyBundle) CryptoSession {
	return CryptoSession{
		UserId:            e.UserId,
		SessionToken:      e.SessionToken,
		KeyA:              e.KeyA,
		KeyB:              e.KeyB,
		AccessToken:       e.AccessToken,
		RefreshToken:      e.RefreshToken,
		KeyID:             e.KeyID,
		Timeout:           e.Timeout,
		HawkID:            e.HawkID,
		HawkKey:           e.HawkKey,
		APIEndpoint:       e.APIEndpoint,
		HawkHashAlgorithm: e.HawkHashAlgorithm,
		CryptoKeys:        keys,
	}
}

func (e HawkSession) ToKeylessSession() FFSyncSession {
	return FFSyncSession{
		SessionToken:      e.SessionToken,
		KeyA:              e.KeyA,
		KeyB:              e.KeyB,
		UserId:            e.UserId,
		APIEndpoint:       e.APIEndpoint,
		HawkID:            e.HawkID,
		HawkKey:           e.HawkKey,
		HawkHashAlgorithm: e.HawkHashAlgorithm,
		Timeout:           e.Timeout,
	}
}

func (s CryptoSession) Reduce() FFSyncSession {
	return FFSyncSession{
		SessionToken:      s.SessionToken,
		KeyA:              s.KeyA,
		KeyB:              s.KeyB,
		AccessToken:       s.AccessToken,
		RefreshToken:      s.RefreshToken,
		UserId:            s.UserId,
		APIEndpoint:       s.APIEndpoint,
		HawkID:            s.HawkID,
		HawkKey:           s.HawkKey,
		HawkHashAlgorithm: s.HawkHashAlgorithm,
		Timeout:           s.Timeout,
		BulkKeys:          s.CryptoKeys,
	}
}

func (s FFSyncSession) Save(path string) error {
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0644)
	if err != nil {
		return errorx.Decorate(err, "failed to mkdir directory "+dir)
	}

	sj := sessionJson{
		SessionToken: hex.EncodeToString(s.SessionToken),
		KeyA:         hex.EncodeToString(s.KeyA),
		KeyB:         hex.EncodeToString(s.KeyB),
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		UserId:       s.UserId,
		Timeout:      s.Timeout.UnixMicro(),
		Hawk: sessionHawkJson{
			APIEndpoint:   s.APIEndpoint,
			ID:            s.HawkID,
			Key:           s.HawkKey,
			HashAlgorithm: s.HawkHashAlgorithm,
			BulkKeys:      make(map[string][]string, len(s.BulkKeys)),
		},
	}
	for k, v := range s.BulkKeys {
		sj.Hawk.BulkKeys[k] = []string{hex.EncodeToString(v.EncryptionKey), hex.EncodeToString(v.HMACKey)}
	}

	dat, err := json.MarshalIndent(sj, "", "  ")
	if err != nil {
		return errorx.Decorate(err, "failed to marshal json")
	}

	err = os.WriteFile(path, dat, 0600)
	if err != nil {
		return errorx.Decorate(err, "failed to write file")
	}

	return nil
}

func (s FFSyncSession) Expired() bool {
	return s.Timeout.Before(time.Now().Add(15 * time.Minute))
}

func (s FFSyncSession) ToKeyed() KeyedSession {
	return KeyedSession{
		UserId:       s.UserId,
		SessionToken: s.SessionToken,
		KeyA:         s.KeyA,
		KeyB:         s.KeyB,
	}
}

func LoadSession(ctx *cli.FFSContext, path string) (FFSyncSession, error) {

	dat, err := os.ReadFile(path)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to open sessionfile")
	}

	var sj sessionJson
	err = json.Unmarshal(dat, &sj)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to unmarshal session file")
	}

	sessionToken, err := hex.DecodeString(sj.SessionToken)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to unmarshal session file (SessionToken)")
	}

	keya, err := hex.DecodeString(sj.KeyA)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to unmarshal session file (KeyA)")
	}

	keyb, err := hex.DecodeString(sj.KeyB)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to unmarshal session file (KeyB)")
	}

	bulkkeys := make(map[string]KeyBundle, len(sj.Hawk.BulkKeys))
	for k, v := range sj.Hawk.BulkKeys {

		if len(v) != 2 {
			return FFSyncSession{}, errorx.InternalError.New("failed to decode bulkKeys['" + k + "']: must be an array with two values")
		}

		ec, err := hex.DecodeString(v[0])
		if err != nil {
			return FFSyncSession{}, errorx.Decorate(err, "failed to decode bulkKeys['"+k+"'][0]")
		}

		hc, err := hex.DecodeString(v[1])
		if err != nil {
			return FFSyncSession{}, errorx.Decorate(err, "failed to decode bulkKeys['"+k+"'][1]")
		}

		bulkkeys[k] = KeyBundle{EncryptionKey: ec, HMACKey: hc}
	}

	accessToken := sj.AccessToken
	if accessToken == "" {
		return FFSyncSession{}, errorx.InternalError.New("failed to load session: AccessToken is empty")
	}

	refreshToken := sj.RefreshToken
	if refreshToken == "" {
		return FFSyncSession{}, errorx.InternalError.New("failed to load session: RefreshToken is empty")
	}

	ctx.PrintVerboseKV("SessionToken", sessionToken)
	ctx.PrintVerboseKV("KeyA", keya)
	ctx.PrintVerboseKV("KeyB", keyb)
	ctx.PrintVerboseKV("AccessToken", accessToken)
	ctx.PrintVerboseKV("RefreshToken", refreshToken)
	ctx.PrintVerboseKV("UserId", sj.UserId)
	ctx.PrintVerboseKV("HawkAPIEndpoint", sj.Hawk.APIEndpoint)
	ctx.PrintVerboseKV("HawkID", sj.Hawk.ID)
	ctx.PrintVerboseKV("HawkKey", sj.Hawk.Key)
	ctx.PrintVerboseKV("HawkHashAlgorithm", sj.Hawk.HashAlgorithm)
	ctx.PrintVerboseKV("Timeout", time.UnixMicro(sj.Timeout))
	for k, v := range bulkkeys {
		ctx.PrintVerboseKV("BulkKeys['"+k+"'].HMACKey", v.HMACKey)
		ctx.PrintVerboseKV("BulkKeys['"+k+"'].EncryptionKey", v.EncryptionKey)
	}

	return FFSyncSession{
		SessionToken:      sessionToken,
		KeyA:              keya,
		KeyB:              keyb,
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		UserId:            sj.UserId,
		APIEndpoint:       sj.Hawk.APIEndpoint,
		HawkID:            sj.Hawk.ID,
		HawkKey:           sj.Hawk.Key,
		HawkHashAlgorithm: sj.Hawk.HashAlgorithm,
		Timeout:           time.UnixMicro(sj.Timeout),
		BulkKeys:          bulkkeys,
	}, nil
}
