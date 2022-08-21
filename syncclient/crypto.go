package syncclient

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

func stretchPassword(email string, password string) []byte {
	return pbkdf2.Key([]byte(password), []byte("identity.mozilla.com/picl/v1/quickStretch:"+email), 1000, 32, sha256.New)
}

func deriveKey(secret []byte, namespace string, size int) ([]byte, error) {
	r := hkdf.New(sha256.New, secret, make([]byte, 0), []byte("identity.mozilla.com/picl/v1/"+namespace))
	p := make([]byte, size)
	n, err := r.Read(p)
	if err != nil {
		return nil, errorx.Decorate(err, "hkdf failed")
	}
	if n < size {
		return nil, errorx.InternalError.New("Not enough data in hkdf")
	}
	return p, nil
}

func randBytes(size int) []byte {
	b := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}
	return b
}

func unbundle(namespace string, bundleKey []byte, payload []byte) ([]byte, error) {
	// Split off the last 32 bytes, they're the HMAC.
	ciphertext := payload[:len(payload)-32]
	expectedHMAC := payload[len(payload)-32:]

	// Derive enough key material for HMAC-check and decryption.
	size := 32 + len(ciphertext)
	keyMaterial, err := deriveKey(bundleKey, namespace, size)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to derive key from bundle")
	}

	// Check the HMAC using the derived key.
	hmacKey := keyMaterial[:32]
	okay := verifyHMAC(hmacKey, ciphertext, expectedHMAC)
	if !okay {
		return nil, errorx.InternalError.New("failed to verify hmac")
	}

	// XOR-decrypt the ciphertext using the derived key.
	xorKey := keyMaterial[32:]
	return langext.BytesXOR(xorKey, ciphertext)
}

func verifyHMAC(key []byte, data []byte, insig []byte) bool {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	rsig := h.Sum(nil)

	return bytes.Equal(rsig, insig)
}

func decryptPayload(rawciphertext string, ciphertext []byte, iv []byte, hmacval []byte, key KeyBundle) ([]byte, error) {

	hmacBuilder := hmac.New(sha256.New, key.HMACKey)
	hmacBuilder.Write([]byte(rawciphertext))
	expectedHMAC := hmacBuilder.Sum(nil)

	if !bytes.Equal(hmacval, expectedHMAC) {
		return nil, errorx.InternalError.New("HMAC mismatch")
	}

	block, err := aes.NewCipher(key.EncryptionKey)
	if err != nil {
		return nil, errorx.Decorate(err, "cannot create aes cipher")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext = bytes.Trim(plaintext, " \n\t")

	eot := bytes.LastIndexByte(plaintext, '}')

	plaintext = plaintext[:eot+1]

	return plaintext, nil
}

func PKCS7Unpad(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}
