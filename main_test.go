package main

import (
	"fmt"
	"html"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResultHTMLSlots(t *testing.T) {
	imageURL := "https://example.test/image.png"
	bbcode := fmt.Sprintf(`[img src=%s]`, imageURL)
	fragment := fmt.Sprintf(resultHTML, html.EscapeString("catbox.moe"), html.EscapeString(bbcode), html.EscapeString(imageURL))

	if !strings.Contains(fragment, `[img src=https://example.test/image.png]`) {
		t.Fatalf("fragment does not include escaped bbcode in textarea: %s", fragment)
	}
	if !strings.Contains(fragment, `href="https://example.test/image.png"`) {
		t.Fatalf("fragment does not include image URL in link href: %s", fragment)
	}
}

func TestUploadImageFallsBackToImgBB(t *testing.T) {
	var calls []string
	a := &app{
		imgbbKey: "test-key",
		client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls = append(calls, req.URL.Host)
			if req.URL.Host == "catbox.moe" {
				return response(500, "catbox failed"), nil
			}
			if req.URL.Host == "api.imgbb.com" {
				return response(200, `{"success":true,"data":{"url":"https://i.ibb.co/test.png"}}`), nil
			}
			t.Fatalf("unexpected host: %s", req.URL.Host)
			return nil, nil
		})},
	}

	gotURL, host, err := a.uploadImage(t.Context(), []byte("fake image"), "test.png")
	if err != nil {
		t.Fatalf("uploadImage returned error: %v", err)
	}
	if gotURL != "https://i.ibb.co/test.png" || host != "imgbb" {
		t.Fatalf("unexpected upload result: %q %q", gotURL, host)
	}
	if strings.Join(calls, ",") != "catbox.moe,api.imgbb.com" {
		t.Fatalf("unexpected fallback order: %#v", calls)
	}
}

func TestUploadImageUsesCatboxFirst(t *testing.T) {
	var calls []string
	a := &app{
		imgbbKey: "test-key",
		client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls = append(calls, req.URL.Host)
			if req.URL.Host == "catbox.moe" {
				return response(200, "https://files.catbox.moe/test.png"), nil
			}
			t.Fatalf("imgbb should not be called when catbox succeeds")
			return nil, nil
		})},
	}

	gotURL, host, err := a.uploadImage(t.Context(), []byte("fake image"), "test.png")
	if err != nil {
		t.Fatalf("uploadImage returned error: %v", err)
	}
	if gotURL != "https://files.catbox.moe/test.png" || host != "catbox.moe" {
		t.Fatalf("unexpected upload result: %q %q", gotURL, host)
	}
	if strings.Join(calls, ",") != "catbox.moe" {
		t.Fatalf("unexpected call order: %#v", calls)
	}
}

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	limiter := newRateLimiter()
	for i := 0; i < rateLimit; i++ {
		if !limiter.allow("192.0.2.1:1234") {
			t.Fatalf("request %d blocked before limit", i)
		}
	}
	if limiter.allow("192.0.2.1:1234") {
		t.Fatal("request after limit was allowed")
	}
}

func TestUploadRejectsNonImageBeforeProviderCall(t *testing.T) {
	var body strings.Builder
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "note.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("not an image")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	a := &app{
		limiter: newRateLimiter(),
		client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("provider should not be called for invalid content")
			return nil, nil
		})},
	}
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()

	a.upload(res, req)

	if !strings.Contains(res.Body.String(), "Csak JPG, PNG, GIF vagy WebP") {
		t.Fatalf("expected invalid image error, got: %s", res.Body.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	securityHeaders(next).ServeHTTP(res, req)

	headers := res.Result().Header
	if headers.Get("Content-Security-Policy") == "" {
		t.Fatal("missing Content-Security-Policy")
	}
	if headers.Get("Referrer-Policy") != "no-referrer" {
		t.Fatalf("unexpected Referrer-Policy: %q", headers.Get("Referrer-Policy"))
	}
	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("unexpected X-Content-Type-Options: %q", headers.Get("X-Content-Type-Options"))
	}
	if headers.Get("X-Frame-Options") != "DENY" {
		t.Fatalf("unexpected X-Frame-Options: %q", headers.Get("X-Frame-Options"))
	}
}

func TestValidUploadedURL(t *testing.T) {
	tests := map[string]bool{
		"https://files.catbox.moe/test.png": true,
		"https://i.ibb.co/test.png":         true,
		"http://files.catbox.moe/test.png":  false,
		"https://":                          false,
		"https://example.test/a b.png":      false,
		"https://example.test/a].png":       false,
	}
	for raw, want := range tests {
		if got := validUploadedURL(raw); got != want {
			t.Fatalf("validUploadedURL(%q) = %v, want %v", raw, got, want)
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
