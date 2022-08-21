package syncclient

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/joomcode/errorx"
	"golang.org/x/crypto/hkdf"
)

type KeyBundle struct {
	EncryptionKey []byte
	HMACKey       []byte
}

func keyBundleFromMasterKey(master []byte, info string) (KeyBundle, error) {
	r := hkdf.New(sha256.New, master, make([]byte, 0), []byte(info))
	keyMaterial := make([]byte, 2*32)

	n, err := r.Read(keyMaterial)
	if err != nil {
		return KeyBundle{}, errorx.Decorate(err, "hkdf failed")
	}
	if n < 2*32 {
		return KeyBundle{}, errorx.InternalError.New("Not enough data in hkdf")
	}

	return KeyBundle{
		EncryptionKey: keyMaterial[:32],
		HMACKey:       keyMaterial[32:],
	}, nil
}

func keyBundleFromB64Array(arr []string) (KeyBundle, error) {
	if len(arr) != 2 {
		return KeyBundle{}, errorx.InternalError.New("keydata must be an array with two values")
	}

	ec, err := base64.StdEncoding.DecodeString(arr[0])
	if err != nil {
		return KeyBundle{}, errorx.Decorate(err, "failed to decode [0]")
	}

	hc, err := base64.StdEncoding.DecodeString(arr[1])
	if err != nil {
		return KeyBundle{}, errorx.Decorate(err, "failed to decode [1]")
	}

	return KeyBundle{
		EncryptionKey: ec,
		HMACKey:       hc,
	}, nil
}
