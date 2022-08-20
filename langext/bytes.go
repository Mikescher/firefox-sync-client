package langext

import "github.com/joomcode/errorx"

func BytesXOR(a []byte, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, errorx.InternalError.New("length mismatch")
	}

	r := make([]byte, len(a))

	for i := 0; i < len(a); i++ {
		r[i] = a[i] ^ b[i]
	}

	return r, nil
}
