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
	"io"
	"strings"

	"github.com/joomcode/errorx"
	"github.com/zenazn/pkcs7pad"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
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
	// If data is padded (with PKCS#7) we add bytes until len(data) is a multiple of blocksize
	// the added bytes equal the amount of added bytes
	// see https://commons.wikimedia.org/wiki/File:Padding_en.png

	// Trivial case - no data, no padding
	if len(data) == 0 {
		return data
	}

	// data-length is not a multiple of blocksize
	// This means teh data is not padded (should not really be possible?)
	// Anyway - just return the plain data - we will probably fail later...
	if len(data)%blocksize != 0 {
		ctx.PrintVerbose("Failed to determine padding (invalid data len), return raw data")
		return data
	}

	pkcsPadLen := int(data[len(data)-1])
	lastBlock := data[len(data)-blocksize:]

	if pkcsPadLen == blocksize {
		//PKCS#7 padded payload - whole last block is padding
		padOkay := true
		for i := 0; i < pkcsPadLen; i++ {
			if data[len(data)-1-i] != byte(pkcsPadLen) {
				padOkay = false
			}
		}
		if padOkay {
			// The whole last block is padding - simply remove it
			return data[:len(data)-pkcsPadLen]
		} else {
			// invalid padding?
			// the last byte should determine the amount of padding
			// and the padding should then fill the last {x} bytes with {x}
			ctx.PrintVerbose("Failed to determine padding, return raw data")
			ctx.PrintVerbose("Last data-block: " + hex.EncodeToString(lastBlock))
			return data
		}
	} else if pkcsPadLen < blocksize {
		//PKCS#7 padded payload - last block is partially padded
		padOkay := true
		for i := 0; i < pkcsPadLen; i++ {
			if data[len(data)-1-i] != byte(pkcsPadLen) {
				padOkay = false
			}
		}
		if padOkay {
			// Remove the last {pkcsPadLen} bytes - they are padding
			return data[:len(data)-pkcsPadLen]
		} else {
			// invalid padding?
			// the last byte should determine the amount of padding
			// and the padding should then fill the last {x} bytes with {x}
			ctx.PrintVerbose("Failed to determine padding, return raw data")
			ctx.PrintVerbose("Last data-block: " + hex.EncodeToString(lastBlock))
			return data
		}
	} else {
		// invalid padding?
		// the last byte should determine the amount of padding - and that should never be more than the blocksize

		ctx.PrintVerbose("Failed to determine padding, return raw data")
		ctx.PrintVerbose("Last data-block: " + hex.EncodeToString(lastBlock))
		return data
	}
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
