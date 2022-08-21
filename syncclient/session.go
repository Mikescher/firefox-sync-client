package syncclient

import (
	"encoding/hex"
	"encoding/json"
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
	"os"
	"path/filepath"
	"time"
)

type LoginSession struct {
	URL             string
	Email           string
	StretchPassword []byte
	UserId          string
	SessionToken    []byte
	KeyFetchToken   []byte
}

type FxAKeys struct {
	KeyA []byte
	KeyB []byte
}

type KeyedSession struct {
	LoginSession
	FxAKeys
}

type FxABrowserID struct {
	BrowserID    string
	CertTime     time.Time
	CertDuration time.Duration
}

type BrowserIdSession struct {
	LoginSession
	FxAKeys
	FxABrowserID
}

type HawkCredentials struct {
	HawkID            string
	HawkKey           string
	APIEndpoint       string
	HawkDuration      int64
	HawkHashAlgorithm string
}

type HawkSession struct {
	LoginSession
	FxAKeys
	FxABrowserID
	HawkCredentials
}

type CryptoKeys struct {
	Keys map[string]KeyBundle
}

type CryptoSession struct {
	LoginSession
	FxAKeys
	FxABrowserID
	HawkCredentials
	CryptoKeys
}

type FFSyncSession struct {
	SessionToken      []byte
	KeyB              []byte
	UserId            string
	APIEndpoint       string
	HawkID            string
	HawkKey           string
	HawkHashAlgorithm string
	HawkTimeout       time.Time
	BulkKeys          map[string]KeyBundle
}

type sessionHawkJson struct {
	APIEndpoint   string              `json:"apiEndpoint"`
	ID            string              `json:"id"`
	Key           string              `json:"key"`
	HashAlgorithm string              `json:"algorithm"`
	Timeout       int64               `json:"timeout"`
	BulkKeys      map[string][]string `json:"bulkKeys"`
}

type sessionJson struct {
	SessionToken string          `json:"sessionToken"`
	KeyB         string          `json:"keyB"`
	UserId       string          `json:"userID"`
	Hawk         sessionHawkJson `json:"hawk"`
}

func (s LoginSession) Extend(ka []byte, kb []byte) KeyedSession {
	return KeyedSession{
		LoginSession: s,
		FxAKeys: FxAKeys{
			KeyA: ka,
			KeyB: kb,
		},
	}
}

func (e KeyedSession) Extend(bid string, t0 time.Time, dur time.Duration) BrowserIdSession {
	return BrowserIdSession{
		LoginSession: e.LoginSession,
		FxAKeys:      e.FxAKeys,
		FxABrowserID: FxABrowserID{
			BrowserID:    bid,
			CertTime:     t0,
			CertDuration: dur,
		},
	}
}

func (e BrowserIdSession) Extend(cred HawkCredentials) HawkSession {
	return HawkSession{
		LoginSession:    e.LoginSession,
		FxAKeys:         e.FxAKeys,
		FxABrowserID:    e.FxABrowserID,
		HawkCredentials: cred,
	}
}

func (e HawkSession) Extend(keys CryptoKeys) CryptoSession {
	return CryptoSession{
		LoginSession:    e.LoginSession,
		FxAKeys:         e.FxAKeys,
		FxABrowserID:    e.FxABrowserID,
		HawkCredentials: e.HawkCredentials,
		CryptoKeys:      keys,
	}
}

func (e HawkSession) ToKeylessSession() FFSyncSession {
	return FFSyncSession{
		SessionToken:      e.SessionToken,
		KeyB:              e.KeyB,
		UserId:            e.UserId,
		APIEndpoint:       e.APIEndpoint,
		HawkID:            e.HawkID,
		HawkKey:           e.HawkKey,
		HawkHashAlgorithm: e.HawkHashAlgorithm,
		HawkTimeout:       e.CertTime.Add(e.CertDuration),
	}
}

func (s CryptoSession) Reduce() FFSyncSession {
	return FFSyncSession{
		SessionToken:      s.SessionToken,
		KeyB:              s.KeyB,
		UserId:            s.UserId,
		APIEndpoint:       s.APIEndpoint,
		HawkID:            s.HawkID,
		HawkKey:           s.HawkKey,
		HawkHashAlgorithm: s.HawkHashAlgorithm,
		HawkTimeout:       s.CertTime.Add(s.CertDuration),
		BulkKeys:          s.Keys,
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
		KeyB:         hex.EncodeToString(s.KeyB),
		UserId:       s.UserId,
		Hawk: sessionHawkJson{
			APIEndpoint:   s.APIEndpoint,
			ID:            s.HawkID,
			Key:           s.HawkKey,
			HashAlgorithm: s.HawkHashAlgorithm,
			Timeout:       s.HawkTimeout.UnixMicro(),
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

func LoadSession(ctx *cli.FFSContext, path string) (FFSyncSession, error) {

	dat, err := os.ReadFile(path)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to open configfile")
	}

	var sj sessionJson
	err = json.Unmarshal(dat, &sj)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to unmarshal config-file")
	}

	sessionToken, err := hex.DecodeString(sj.SessionToken)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to unmarshal config-file (SessionToken)")
	}

	keyb, err := hex.DecodeString(sj.KeyB)
	if err != nil {
		return FFSyncSession{}, errorx.Decorate(err, "failed to unmarshal config-file (KeyB)")
	}

	bulkkeys := make(map[string]KeyBundle, len(sj.Hawk.BulkKeys))
	for k, v := range sj.Hawk.BulkKeys {
		kb, err := keyBundleFromB64Array(v)
		if err != nil {
			return FFSyncSession{}, errorx.Decorate(err, "failed to unmarshal config-file (BulkKeys)")
		}
		bulkkeys[k] = kb
	}

	return FFSyncSession{
		SessionToken:      sessionToken,
		KeyB:              keyb,
		UserId:            sj.UserId,
		APIEndpoint:       sj.Hawk.APIEndpoint,
		HawkID:            sj.Hawk.ID,
		HawkKey:           sj.Hawk.Key,
		HawkHashAlgorithm: sj.Hawk.HashAlgorithm,
		HawkTimeout:       time.UnixMicro(sj.Hawk.Timeout),
		BulkKeys:          bulkkeys,
	}, nil
}
