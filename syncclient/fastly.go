package syncclient

// Fastly anti-bot challenge solver.
//
// Based on the algorithm described in github.com/pagpeter/fastly-antibot:
// - GET protected page → extract challenge script ID from HTML
// - GET challenge script → extract token
// - POST PAT check (optional Apple PAT, always fails gracefully)
// - POST fst-post-back to receive PoW + clientmetrics challenges
// - Solve PoW: brute-force a short suffix such that SHA256(base+suffix) == target
// - POST solutions → server sets _fs_ch_cp_* cookie (valid ~1 hour)

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"ffsyncclient/cli"
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

// The challenge is always solved against accounts.firefox.com: it serves the
// Fastly challenge page and yields a _fs_ch_cp_* cookie scoped to .firefox.com,
// which covers api.accounts.firefox.com and the other Firefox subdomains.
const fastlyChallengePageURL = "https://accounts.firefox.com/"

const fastlyUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"

// fastlyMaxPOWSuffixLen bounds the brute-force; live challenges use a 2-char suffix.
const fastlyMaxPOWSuffixLen = 3

var (
	fastlyScriptIDRegex = regexp.MustCompile(`_fs-ch-([^/'"?\s]+)`)
	fastlyTokenRegex    = regexp.MustCompile(`init\(\[[^\]]*\],\s*"([^"]+)"`)
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

// solvePOW finds a suffix (up to fastlyMaxPOWSuffixLen chars) such that
// SHA256(base+suffix) hex-encodes to targetHex. Returns false if none is found.
func solvePOW(base, targetHex string) (string, bool) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	target := strings.ToLower(targetHex)
	for length := 1; length <= fastlyMaxPOWSuffixLen; length++ {
		idx := make([]int, length)
		for {
			suffix := make([]byte, length)
			for k := range idx {
				suffix[k] = charset[idx[k]]
			}
			sum := sha256.Sum256([]byte(base + string(suffix)))
			if hex.EncodeToString(sum[:]) == target {
				return string(suffix), true
			}
			p := length - 1
			for ; p >= 0; p-- {
				idx[p]++
				if idx[p] < len(charset) {
					break
				}
				idx[p] = 0
			}
			if p < 0 {
				break
			}
		}
	}
	return "", false
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

type fastlyPOWSolution struct {
	Ty      string `json:"ty"`
	Base    string `json:"base"`
	Answer  string `json:"answer"`
	Hmac    string `json:"hmac"`
	Expires string `json:"expires"`
}

type fastlyClientMetrics struct {
	Ty                 string `json:"ty"`
	Webdriver          bool   `json:"webdriver"`
	BotDetectionResult struct {
		BotDetected bool `json:"bot_detected"`
		BotKind     any  `json:"bot_kind"`
	} `json:"bot_detection_result"`
	BrowserMetrics struct {
		ClientData string `json:"client_data"`
		ErrorTrace string `json:"error_trace"`
	} `json:"browser_metrics"`
}

func newFastlyClientMetrics() fastlyClientMetrics {
	r := fastlyClientMetrics{
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

func fastlyGetScriptID(ctx *cli.FFSContext, client *http.Client, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
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
	m := fastlyScriptIDRegex.FindSubmatch(body)
	if len(m) < 2 {
		return "", fmt.Errorf("fastly challenge script ID not found in response")
	}
	return string(m[1]), nil
}

func fastlyGetToken(ctx *cli.FFSContext, client *http.Client, domain, scriptID, referer string) (string, error) {
	scriptURL := fmt.Sprintf("%s/_fs-ch-%s/script.js?reload=true", domain, scriptID)
	req, err := http.NewRequestWithContext(ctx, "GET", scriptURL, nil)
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
	m := fastlyTokenRegex.FindSubmatch(body)
	if len(m) < 2 {
		return "", fmt.Errorf("fastly challenge token not found in script")
	}
	return string(m[1]), nil
}

func fastlyDoPAT(ctx *cli.FFSContext, client *http.Client, domain, scriptID, token, referer string) {
	patURL := fmt.Sprintf("%s/_fs-ch-%s/pat?token=%s", domain, scriptID, token)
	req, err := http.NewRequestWithContext(ctx, "POST", patURL, nil)
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

func fastlyPostBack(ctx *cli.FFSContext, client *http.Client, domain, scriptID, referer, token string, data any) (*fastlyPostBackResponse, error) {
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

	req, err := http.NewRequestWithContext(ctx, "POST", postURL, strings.NewReader(payload))
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

func fastlySolveChallenges(ctx *cli.FFSContext, challenges *fastlyPostBackResponse) ([]any, error) {
	results := make([]any, 0, len(challenges.Challenges))
	for _, chl := range challenges.Challenges {
		switch chl.Type {
		case "pow":
			answer, ok := solvePOW(chl.Data.Base, chl.Data.Hash)
			if !ok {
				return nil, fmt.Errorf("failed to solve proof-of-work challenge (base=%q, hash=%q)", chl.Data.Base, chl.Data.Hash)
			}
			ctx.PrintVerbose("Solved Fastly proof-of-work challenge")
			results = append(results, fastlyPOWSolution{
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
	return results, nil
}

// solveFastlyChallenge performs the full Fastly non-interactive PoW challenge
// flow for targetURL and returns the resulting cookie as "name=value". It reuses
// the base client's transport (proxy/TLS settings) but with its own cookie jar.
func solveFastlyChallenge(ctx *cli.FFSContext, base *http.Client, targetURL string) (string, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{
		Transport: base.Transport,
		Timeout:   base.Timeout,
		Jar:       jar,
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	domain := parsedURL.Scheme + "://" + parsedURL.Host

	ctx.PrintVerbose("Fetching Fastly challenge script ID from " + targetURL)
	scriptID, err := fastlyGetScriptID(ctx, client, targetURL)
	if err != nil {
		return "", fmt.Errorf("fastly challenge: %w", err)
	}

	token, err := fastlyGetToken(ctx, client, domain, scriptID, targetURL)
	if err != nil {
		return "", fmt.Errorf("fastly challenge: %w", err)
	}

	fastlyDoPAT(ctx, client, domain, scriptID, token, targetURL)

	initResp, err := fastlyPostBack(ctx, client, domain, scriptID, targetURL, token, []map[string]string{{"ty": "pat", "auth": ""}})
	if err != nil {
		return "", fmt.Errorf("fastly challenge: initial post-back: %w", err)
	}

	token = initResp.Token
	solutions, err := fastlySolveChallenges(ctx, initResp)
	if err != nil {
		return "", fmt.Errorf("fastly challenge: %w", err)
	}

	finalResp, err := fastlyPostBack(ctx, client, domain, scriptID, targetURL, token, solutions)
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
		ctx.PrintVerbose(fmt.Sprintf("No _fs_ch_cp_* cookie found, falling back to cookie %q", cookies[0].Name))
		return cookies[0].Name + "=" + cookies[0].Value, nil
	}
	return "", fmt.Errorf("fastly challenge: no cookie found after successful solve")
}
