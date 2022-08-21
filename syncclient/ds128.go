package syncclient

import (
	"crypto/dsa"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"github.com/joomcode/errorx"
)

type SigningMethodDS128 struct {
}

func (m *SigningMethodDS128) Alg() string {
	return "DS128"
}

// Verify implements token verification for the SigningMethod.
// For this verify method, key must be an ecdsa.PublicKey struct
func (m *SigningMethodDS128) Verify(signingString, signature string, key interface{}) error {
	panic("Not implemented")
}

// Sign implements token signing for the SigningMethod.
// For this signing method, key must be an ecdsa.PrivateKey struct
func (m *SigningMethodDS128) Sign(signingString string, key interface{}) (string, error) {
	// Get the key
	var dsaKey *dsa.PrivateKey
	switch k := key.(type) {
	case *dsa.PrivateKey:
		dsaKey = k
	default:
		return "", errorx.InternalError.New("ErrInvalidKeyType")
	}

	hasher := sha1.New()
	hasher.Write([]byte(signingString))

	// Sign the string and return r, s
	if r, s, err := dsa.Sign(rand.Reader, dsaKey, hasher.Sum(nil)); err == nil {

		keyBytes := 160 / 8

		// We serialize the outputs (r and s) into big-endian byte arrays
		// padded with zeros on the left to make sure the sizes work out.
		// Output must be 2*keyBytes long.
		out := make([]byte, 2*keyBytes)
		r.FillBytes(out[0:keyBytes]) // r is assigned to the first half of output.
		s.FillBytes(out[keyBytes:])  // s is assigned to the second half of output.

		return base64.RawURLEncoding.EncodeToString(out), nil
	} else {
		return "", err
	}
}
