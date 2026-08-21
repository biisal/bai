package providers

import (
	"net/http"
	"testing"

	"github.com/biisal/bai/internal/agent/providers/variant"
	"github.com/openai/openai-go/v3/option"
)

func captureNext(headers *http.Header) option.MiddlewareNext {
	return func(req *http.Request) (*http.Response, error) {
		*headers = req.Header.Clone()
		return &http.Response{StatusCode: http.StatusOK, Header: req.Header}, nil
	}
}

func TestOpenAIMiddleware(t *testing.T) {
	spec := &variant.Spec{
		Name: variant.OpenCode,
		Headers: []variant.Header{
			{Key: "x-test", Value: func() string { return "1" }},
			{Key: "x-dynamic", Value: func() string { return "req-42" }},
		},
		AuthScheme:   "Bearer",
		AuthFallback: "public",
	}

	newReq := func(t *testing.T) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, "https://x", nil)
		if err != nil {
			t.Fatal(err)
		}
		return req
	}

	t.Run("sets spec headers", func(t *testing.T) {
		var got http.Header
		variantMiddleware(spec, "sk-real")(newReq(t), captureNext(&got))
		if v := got.Get("x-test"); v != "1" {
			t.Errorf("x-test = %q, want %q", v, "1")
		}
		if v := got.Get("x-dynamic"); v != "req-42" {
			t.Errorf("x-dynamic = %q, want %q", v, "req-42")
		}
	})

	t.Run("empty api key falls back to public bearer", func(t *testing.T) {
		var got http.Header
		variantMiddleware(spec, "")(newReq(t), captureNext(&got))
		if auth := got.Get("Authorization"); auth != "Bearer public" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer public")
		}
	})

	t.Run("real api key is never overridden", func(t *testing.T) {
		var got http.Header
		req := newReq(t)
		req.Header.Set("Authorization", "Bearer sk-real")
		variantMiddleware(spec, "sk-real")(req, captureNext(&got))
		if auth := got.Get("Authorization"); auth != "Bearer sk-real" {
			t.Errorf("Authorization = %q, want untouched %q", auth, "Bearer sk-real")
		}
	})

	t.Run("spec without fallback leaves auth alone", func(t *testing.T) {
		noAuth := &variant.Spec{Headers: spec.Headers}
		var got http.Header
		variantMiddleware(noAuth, "")(newReq(t), captureNext(&got))
		if auth := got.Get("Authorization"); auth != "" {
			t.Errorf("Authorization = %q, want empty", auth)
		}
	})
}

func TestApplyVariant(t *testing.T) {
	if got := applyVariant(nil, ""); got != nil {
		t.Errorf("applyVariant(nil) = %v, want nil", got)
	}

	spec := &variant.Spec{}
	opts := applyVariant(spec, "")
	if len(opts) != 1 {
		t.Errorf("len(opts) = %d, want 1 middleware", len(opts))
	}
}
