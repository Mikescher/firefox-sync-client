package syncclient

// Fastly anti-bot challenge solver.
//
// Based on the algorithm described in github.com/pagpeter/fastly-antibot:
// - GET protected page → extract challenge script ID from HTML
// - GET challenge script → extract token
// - POST PAT check (optional Apple PAT, always fails gracefully)
// - POST fst-post-back to receive PoW + clientmetrics challenges
// - Solve PoW: brute-force 2-char suffix such that SHA256(base+suffix) == target
// - POST solutions → server sets _fs_ch_cp_* cookie (valid ~1 hour)

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ── cookie cache ────────────────────────────────────────────────────────────

type fastlyCookieEntry struct {
	value     string
	expiresAt time.Time
}

type fastlyCookieCache struct {
	mu      sync.Mutex
	entries map[string]fastlyCookieEntry
}

func newFastlyCookieCache() *fastlyCookieCache {
	return &fastlyCookieCache{entries: make(map[string]fastlyCookieEntry)}
}

func (c *fastlyCookieCache) get(domain string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[domain]
	if !ok || time.Now().After(e.expiresAt) {
		return "", false
	}
	return e.value, true
}

func (c *fastlyCookieCache) set(domain, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[domain] = fastlyCookieEntry{
		value:     value,
		expiresAt: time.Now().Add(50 * time.Minute),
	}
}

// ── PoW solver ───────────────────────────────────────────────────────────────

func solvePOW(base, targetHex string) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	target := strings.ToLower(targetHex)
	for i := 0; i < len(charset); i++ {
		for j := 0; j < len(charset); j++ {
			suffix := string([]byte{charset[i], charset[j]})
			sum := sha256.Sum256([]byte(base + suffix))
			if hex.EncodeToString(sum[:]) == target {
				return suffix
			}
		}
	}
	return ""
}

// ── challenge protocol types ─────────────────────────────────────────────────

type fastlyPOWChallenge struct {
	Base    string `json:"base"`
	Hash    string `json:"hash"`
	Hmac    string `json:"hmac"`
	Expires string `json:"expires"`
}

type fastlyOuterChallenge struct {
	Type string             `json:"ty"`
	Data fastlyPOWChallenge `json:"data"`
}

type fastlyPostBackResponse struct {
	Challenges []fastlyOuterChallenge `json:"ch"`
	Token      string                 `json:"tok"`
	Status     string                 `json:"status"`
}

type fastlyPOWResponse struct {
	Ty      string `json:"ty"`
	Base    string `json:"base"`
	Answer  string `json:"answer"`
	Hmac    string `json:"hmac"`
	Expires string `json:"expires"`
}

type fastlyClientMetricsResponse struct {
	Ty                 string `json:"ty"`
	Webdriver          bool   `json:"webdriver"`
	BotDetectionResult struct {
		BotDetected bool        `json:"bot_detected"`
		BotKind     interface{} `json:"bot_kind"`
	} `json:"bot_detection_result"`
	BrowserMetrics struct {
		ClientData string `json:"client_data"`
		ErrorTrace string `json:"error_trace"`
	} `json:"browser_metrics"`
}

func newFastlyClientMetrics() fastlyClientMetricsResponse {
	r := fastlyClientMetricsResponse{
		Ty:        "clientmetrics",
		Webdriver: false,
	}
	r.BotDetectionResult.BotDetected = false
	r.BotDetectionResult.BotKind = nil
	// Minimal values — confirmed sufficient by live testing (pypi.org, accounts.firefox.com).
	r.BrowserMetrics.ClientData = `{}`
	r.BrowserMetrics.ErrorTrace = `""`
	return r
}

// ── HTTP helpers ─────────────────────────────────────────────────────────────

const fastlyUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"

func fastlyGetScriptID(client *http.Client, targetURL string) (string, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", fastlyUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`_fs-ch-([^/'"?\s]+)`)
	m := re.FindSubmatch(body)
	if len(m) < 2 {
		return "", fmt.Errorf("Fastly challenge script ID not found in response")
	}
	return string(m[1]), nil
}

func fastlyGetToken(client *http.Client, domain, scriptID, referer string) (string, error) {
	scriptURL := fmt.Sprintf("%s/_fs-ch-%s/script.js?reload=true", domain, scriptID)
	req, err := http.NewRequest("GET", scriptURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", fastlyUA)
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept", "*/*")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`init\(\[[^\]]*\],\s*"([^"]+)"`)
	m := re.FindSubmatch(body)
	if len(m) < 2 {
		return "", fmt.Errorf("Fastly challenge token not found in script")
	}
	return string(m[1]), nil
}

func fastlyDoPAT(client *http.Client, domain, scriptID, token, referer string) {
	patURL := fmt.Sprintf("%s/_fs-ch-%s/pat?token=%s", domain, scriptID, token)
	req, err := http.NewRequest("POST", patURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", fastlyUA)
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func fastlyPostBack(client *http.Client, domain, scriptID, referer, token string, data any) (*fastlyPostBackResponse, error) {
	postURL := fmt.Sprintf("%s/_fs-ch-%s/fst-post-back", domain, scriptID)

	buf := &strings.Builder{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	dataJSON := strings.TrimSpace(buf.String())

	// Build payload: {"token":"<tok>","data":<data>}
	tokenJSON, _ := json.Marshal(token)
	payload := fmt.Sprintf(`{"token":%s,"data":%s}`, string(tokenJSON), dataJSON)

	req, err := http.NewRequest("POST", postURL, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fastlyUA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", referer)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result fastlyPostBackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func fastlySolveChallenges(challenges *fastlyPostBackResponse) []any {
	results := make([]any, 0, len(challenges.Challenges))
	for _, chl := range challenges.Challenges {
		switch chl.Type {
		case "pow":
			answer := solvePOW(chl.Data.Base, chl.Data.Hash)
			results = append(results, fastlyPOWResponse{
				Ty:      "pow",
				Base:    chl.Data.Base,
				Answer:  answer,
				Hmac:    chl.Data.Hmac,
				Expires: chl.Data.Expires,
			})
		case "clientmetrics":
			results = append(results, newFastlyClientMetrics())
		}
	}
	return results
}

// solveFastlyChallenge performs the full Fastly non-interactive PoW challenge
// flow for targetURL and returns the resulting cookie as "name=value".
func solveFastlyChallenge(targetURL string) (string, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	domain := parsedURL.Scheme + "://" + parsedURL.Host

	scriptID, err := fastlyGetScriptID(client, targetURL)
	if err != nil {
		return "", fmt.Errorf("fastly challenge: %w", err)
	}

	token, err := fastlyGetToken(client, domain, scriptID, targetURL)
	if err != nil {
		return "", fmt.Errorf("fastly challenge: %w", err)
	}

	fastlyDoPAT(client, domain, scriptID, token, targetURL)

	initResp, err := fastlyPostBack(client, domain, scriptID, targetURL, token, []map[string]string{{"ty": "pat", "auth": ""}})
	if err != nil {
		return "", fmt.Errorf("fastly challenge: initial post-back: %w", err)
	}

	token = initResp.Token
	solutions := fastlySolveChallenges(initResp)

	finalResp, err := fastlyPostBack(client, domain, scriptID, targetURL, token, solutions)
	if err != nil {
		return "", fmt.Errorf("fastly challenge: solution post-back: %w", err)
	}
	if finalResp.Status != "success" {
		return "", fmt.Errorf("fastly challenge: not solved (status=%q)", finalResp.Status)
	}

	for _, c := range jar.Cookies(parsedURL) {
		if strings.HasPrefix(c.Name, "_fs_ch_cp_") {
			return c.Name + "=" + c.Value, nil
		}
	}
	// Fallback to any cookie set (older Fastly versions may use different naming)
	if cookies := jar.Cookies(parsedURL); len(cookies) > 0 {
		return cookies[0].Name + "=" + cookies[0].Value, nil
	}
	return "", fmt.Errorf("fastly challenge: no cookie found after successful solve")
}
