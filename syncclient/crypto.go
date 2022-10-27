package syncclient

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
	"github.com/zenazn/pkcs7pad"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
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

func decryptPayload(ctx *cli.FFSContext, rawciphertext string, rawiv string, rawhmac string, key KeyBundle) ([]byte, error) {
	iv, err := base64.StdEncoding.DecodeString(rawiv)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to b64-decode iv")
	}

	hmacval, err := hex.DecodeString(rawhmac)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to hex-decode hmac")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(rawciphertext)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to b64-decode ciphertext")
	}

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

	plaintext = removePadding(ctx, plaintext, aes.BlockSize)

	return plaintext, nil
}

func removePadding(ctx *cli.FFSContext, data []byte, blocksize int) []byte {
	// this is a weird amount of guesswork?, I can't find any spec how the data must be padded

	if len(data) == 0 {
		return data
	}

	if len(data)%blocksize != 0 {
		// not padded ???
		return data
	}

	pi00 := len(data) % blocksize
	if pi00 == 0 {
		pi00 = blocksize
	}
	pi01 := len(data) - pi00

	if pi01 >= 0 && data[pi01] == '}' {
		// well-formed JSON payload
		return data[:pi01+1]
	}

	if data[len(data)-1] == '}' {
		// well-formed JSON payload without padding ?!?
		return data
	}

	if c := data[len(data)-1]; int(c) <= blocksize {
		//PKCS7 padded payload
		allpad := true
		for i := 0; i < int(c); i++ {
			if data[len(data)-1-i] != c {
				allpad = false
			}
		}
		if allpad {
			return data[:len(data)-int(c)]
		}
	}

	eot := bytes.LastIndexByte(data, '}')
	if eot >= 0 && len(data)-eot-1 <= blocksize {
		// well-formed JSON payload
		return data[:eot+1]
	}

	ctx.PrintVerbose("Failed to determine padding, return raw data")

	// idk, just return data
	return data
}

func encryptPayload(ctx *cli.FFSContext, plaintext string, key KeyBundle) (string, string, string, error) {
	iv := randBytes(16)

	block, err := aes.NewCipher(key.EncryptionKey)
	if err != nil {
		return "", "", "", errorx.Decorate(err, "cannot create aes cipher")
	}

	padplaintext := pkcs7pad.Pad([]byte(plaintext), aes.BlockSize)

	ciphertext := make([]byte, len(padplaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padplaintext)

	rawciphertext := base64.StdEncoding.EncodeToString(ciphertext)

	hmacBuilder := hmac.New(sha256.New, key.HMACKey)
	hmacBuilder.Write([]byte(rawciphertext))
	hmacval := hmacBuilder.Sum(nil)

	rawhmac := hex.EncodeToString(hmacval)

	rawiv := base64.StdEncoding.EncodeToString(iv)

	return rawciphertext, rawiv, rawhmac, nil
}
