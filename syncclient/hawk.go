package syncclient

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/joomcode/errorx"
	"net/url"
	"strings"
	"time"
)

func calcHawkTokenAuth(token []byte, tokentype string, requestMethod string, requestURI string, body string) (string, []byte, error) {

	keyMaterial, err := deriveKey(token, tokentype, 3*32)
	if err != nil {
		return "", nil, errorx.Decorate(err, "failed to derive hawkTokenAuth key")
	}

	id := hex.EncodeToString(keyMaterial[:32])
	authKey := keyMaterial[32:64]
	bundleKey := keyMaterial[64:]

	hashStr := "hawk.1.payload\napplication/json\n" + body + "\n"

	rawHash := sha256.Sum256([]byte(hashStr))

	hash := base64.StdEncoding.EncodeToString(rawHash[:])
	if requestMethod == "GET" || body == "" {
		hash = ""
	}

	nonce := base64.StdEncoding.EncodeToString(randBytes(5))
	ts := fmt.Sprintf("%d", time.Now().Unix())

	requrl, err := url.Parse(requestURI)
	if err != nil {
		return "", nil, errorx.Decorate(err, "failed to parse requestURI: "+requestURI)
	}
	uhost := requrl.Host
	uport := "80"
	if requrl.Scheme == "https" {
		uport = "443"
	}
	if strings.Contains(uhost, ":") {
		_v := uhost
		uhost = _v[0:strings.Index(_v, "=")]
		uport = _v[strings.Index(_v, "=")+1:]
	}

	sigbits := make([]string, 0, 10)
	sigbits = append(sigbits, "hawk.1.header")
	sigbits = append(sigbits, ts)
	sigbits = append(sigbits, nonce)
	sigbits = append(sigbits, requestMethod)
	sigbits = append(sigbits, requrl.Path)
	sigbits = append(sigbits, strings.ToLower(uhost))
	sigbits = append(sigbits, strings.ToLower(uport))
	sigbits = append(sigbits, hash)
	sigbits = append(sigbits, "")
	sigbits = append(sigbits, "")

	sigstr := strings.Join(sigbits, "\n")

	hmacBuilder := hmac.New(sha256.New, authKey)
	hmacBuilder.Write([]byte(sigstr))
	mac := base64.StdEncoding.EncodeToString(hmacBuilder.Sum(nil))

	hdr := `Hawk ` +
		`id="` + id + `", ` +
		`mac="` + mac + `", ` +
		`ts="` + ts + `", ` +
		`nonce="` + nonce + `"`

	if hash != "" {
		hdr += `, hash="` + hash + `"`
	}

	return hdr, bundleKey, nil
}
