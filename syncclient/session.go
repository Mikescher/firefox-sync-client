package syncclient

import (
	"encoding/hex"
	"encoding/json"
	"github.com/joomcode/errorx"
	"os"
	"path/filepath"
	"time"
)

type FxASession struct {
	Email             string
	StretchPassword   []byte
	UserId            string
	SessionToken      string
	KeyFetchToken     string
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
		Email:             e.Email,
		StretchPassword:   hex.EncodeToString(e.StretchPassword),
		UserId:            e.UserId,
		SessionToken:      e.SessionToken,
		KeyFetchToken:     e.KeyFetchToken,
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
