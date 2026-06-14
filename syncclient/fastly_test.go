package syncclient

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"ffsyncclient/cli"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testCtx() *cli.FFSContext {
	return &cli.FFSContext{Context: context.Background()}
}

func TestSolvePOW(t *testing.T) {
	for _, suffix := range []string{"a", "zz", "Ab9"} {
		base := "challenge-base-" + suffix
		sum := sha256.Sum256([]byte(base + suffix))
		target := hex.EncodeToString(sum[:])

		got, ok := solvePOW(base, target)
		if !ok {
			t.Fatalf("solvePOW(%q) returned not-found", suffix)
		}
		check := sha256.Sum256([]byte(base + got))
		if hex.EncodeToString(check[:]) != target {
			t.Fatalf("solvePOW returned %q which does not hash to target", got)
		}
	}
}

func TestSolvePOWUppercaseTarget(t *testing.T) {
	base := "foo"
	sum := sha256.Sum256([]byte(base + "qq"))
	target := strings.ToUpper(hex.EncodeToString(sum[:]))

	if _, ok := solvePOW(base, target); !ok {
		t.Fatal("solvePOW should accept upper-case target hex")
	}
}

func TestSolvePOWNotFound(t *testing.T) {
	if got, ok := solvePOW("base", "deadbeef"); ok {
		t.Fatalf("solvePOW should not find a solution for an impossible target, got %q", got)
	}
}

func TestFastlyCookieCache(t *testing.T) {
	c := newFastlyCookieCache()

	if _, ok := c.get("firefox.com"); ok {
		t.Fatal("empty cache should not return a value")
	}

	c.set("firefox.com", "_fs_ch_cp_x=1")
	if v, ok := c.get("firefox.com"); !ok || v != "_fs_ch_cp_x=1" {
		t.Fatalf("expected cached value, got %q (ok=%v)", v, ok)
	}

	// expired entry must not be returned
	c.entries["firefox.com"] = fastlyCookieEntry{value: "stale", expiresAt: time.Now().Add(-time.Minute)}
	if _, ok := c.get("firefox.com"); ok {
		t.Fatal("expired entry should not be returned")
	}
}

func TestSolveFastlyChallenge(t *testing.T) {
	const scriptID = "testscript"
	const powBase = "abc"
	const powSuffix = "zz"
	powHash := sha256.Sum256([]byte(powBase + powSuffix))
	powHashHex := hex.EncodeToString(powHash[:])

	postBackCalls := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><script src="/_fs-ch-`+scriptID+`/script.js"></script></html>`)
	})
	mux.HandleFunc("/_fs-ch-"+scriptID+"/script.js", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `var x = init([1,2,3], "tok-init", {});`)
	})
	mux.HandleFunc("/_fs-ch-"+scriptID+"/pat", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/_fs-ch-"+scriptID+"/fst-post-back", func(w http.ResponseWriter, r *http.Request) {
		postBackCalls++
		var payload struct {
			Token string          `json:"token"`
			Data  json.RawMessage `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("post-back: failed to decode payload: %v", err)
		}

		if postBackCalls == 1 {
			resp := fastlyPostBackResponse{
				Token:  "tok-final",
				Status: "",
				Challenges: []fastlyOuterChallenge{
					{Type: "pow", Data: fastlyPOWChallenge{Base: powBase, Hash: powHashHex, Hmac: "hm", Expires: "ex"}},
					{Type: "clientmetrics"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// final post-back: verify the PoW answer is correct
		var sols []map[string]any
		if err := json.Unmarshal(payload.Data, &sols); err != nil {
			t.Errorf("post-back: failed to decode solutions: %v", err)
		}
		var answer string
		for _, s := range sols {
			if s["ty"] == "pow" {
				answer, _ = s["answer"].(string)
			}
		}
		if answer != powSuffix {
			t.Errorf("expected pow answer %q, got %q", powSuffix, answer)
		}
		http.SetCookie(w, &http.Cookie{Name: "_fs_ch_cp_test", Value: "cookieval", Path: "/"})
		_ = json.NewEncoder(w).Encode(fastlyPostBackResponse{Status: "success"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cookie, err := solveFastlyChallenge(testCtx(), &http.Client{}, server.URL+"/")
	if err != nil {
		t.Fatalf("solveFastlyChallenge failed: %v", err)
	}
	if cookie != "_fs_ch_cp_test=cookieval" {
		t.Fatalf("unexpected cookie: %q", cookie)
	}
	if postBackCalls != 2 {
		t.Fatalf("expected 2 post-back calls, got %d", postBackCalls)
	}
}
