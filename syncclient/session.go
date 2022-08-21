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

type FxASession struct {
	URL               string
	Email             string
	StretchPassword   []byte
	UserId            string
	SessionToken      []byte
	KeyFetchToken     []byte
	SessionUpdateTime time.Time
}

type FxAKeys struct {
	KeyA           []byte
	KeyB           []byte
	KeysUpdateTime time.Time
}

type FxASessionExt struct {
	FxASession
	FxAKeys
}

type sessionJson struct {
	URL               string `json:"u"`
	Email             string `json:"em"`
	StretchPassword   string `json:"sp"`
	UserId            string `json:"uid"`
	SessionToken      string `json:"st"`
	KeyFetchToken     string `json:"kft"`
	SessionUpdateTime string `json:"sut"`
	KeyA              string `json:"ka"`
	KeyB              string `json:"kb"`
	KeysUpdateTime    string `json:"kut"`
}

func (s FxASession) Extend(ka []byte, kb []byte) FxASessionExt {
	return FxASessionExt{
		FxASession: s,
		FxAKeys: FxAKeys{
			KeyA:           ka,
			KeyB:           kb,
			KeysUpdateTime: time.Now(),
		},
	}
}

func (e FxASessionExt) Save(path string) error {

	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0644)
	if err != nil {
		return errorx.Decorate(err, "failed to mkdir directory "+dir)
	}

	sj := sessionJson{
		URL:               e.URL,
		Email:             e.Email,
		StretchPassword:   hex.EncodeToString(e.StretchPassword),
		UserId:            e.UserId,
		SessionToken:      hex.EncodeToString(e.SessionToken),
		KeyFetchToken:     hex.EncodeToString(e.KeyFetchToken),
		SessionUpdateTime: e.SessionUpdateTime.UTC().Format(time.RFC3339Nano),
		KeyA:              hex.EncodeToString(e.KeyA),
		KeyB:              hex.EncodeToString(e.KeyB),
		KeysUpdateTime:    e.KeysUpdateTime.UTC().Format(time.RFC3339Nano),
	}

	dat, err := json.Marshal(sj)
	if err != nil {
		return errorx.Decorate(err, "failed to marshal json")
	}

	err = os.WriteFile(path, dat, 0600)
	if err != nil {
		return errorx.Decorate(err, "failed to write file")
	}

	return nil
}

func LoadSession(ctx *cli.FFSContext, path string) (FxASessionExt, error) {

	dat, err := os.ReadFile(path)
	if err != nil {
		return FxASessionExt{}, errorx.Decorate(err, "failed to open configfile")
	}

	var sj sessionJson
	err = json.Unmarshal(dat, &sj)
	if err != nil {
		return FxASessionExt{}, errorx.Decorate(err, "failed to unmarshal config-file")
	}

	stretchPW, err := hex.DecodeString(sj.StretchPassword)
	if err != nil {
		return FxASessionExt{}, errorx.Decorate(err, "failed to unmarshal config-file (StretchPassword)")
	}

	sessionUpdateTime, err := time.Parse(time.RFC3339Nano, sj.SessionUpdateTime)
	if err != nil {
		return FxASessionExt{}, errorx.Decorate(err, "failed to unmarshal config-file (SessionUpdateTime)")
	}

	keyA, err := hex.DecodeString(sj.KeyA)
	if err != nil {
		return FxASessionExt{}, errorx.Decorate(err, "failed to unmarshal config-file (KeyA)")
	}

	keyB, err := hex.DecodeString(sj.KeyB)
	if err != nil {
		return FxASessionExt{}, errorx.Decorate(err, "failed to unmarshal config-file (KeyB)")
	}

	keysUpdateTime, err := time.Parse(time.RFC3339Nano, sj.KeysUpdateTime)
	if err != nil {
		return FxASessionExt{}, errorx.Decorate(err, "failed to unmarshal config-file (KeysUpdateTime)")
	}

	st, err := hex.DecodeString(sj.SessionToken)
	if err != nil {
		return FxASessionExt{}, errorx.Decorate(err, "failed to unmarshal config-file (SessionToken)")
	}

	kft, err := hex.DecodeString(sj.KeyFetchToken)
	if err != nil {
		return FxASessionExt{}, errorx.Decorate(err, "failed to unmarshal config-file (KeyFetchToken)")
	}

	session := FxASessionExt{
		FxASession: FxASession{
			URL:               sj.URL,
			Email:             sj.Email,
			StretchPassword:   stretchPW,
			UserId:            sj.UserId,
			SessionToken:      st,
			KeyFetchToken:     kft,
			SessionUpdateTime: sessionUpdateTime,
		},
		FxAKeys: FxAKeys{
			KeyA:           keyA,
			KeyB:           keyB,
			KeysUpdateTime: keysUpdateTime,
		},
	}

	if session.URL != ctx.Opt.ServerURL {
		return FxASessionExt{}, errorx.Decorate(err, "config-file references a different server-url")
	}

	return session, nil

}
