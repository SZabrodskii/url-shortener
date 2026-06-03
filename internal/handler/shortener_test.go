package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SZabrodskii/url-shortener/internal/config"
	"github.com/SZabrodskii/url-shortener/internal/service"
	"github.com/gin-gonic/gin"
)

const testBaseURL config.BaseURL = "http://localhost:8080"

type fakeStorage struct {
	getFn func(id string) (string, error)
	putFn func(id, originalURL string) error
}

func (f *fakeStorage) Get(id string) (string, error) {
	if f.getFn == nil {
		return "", service.ErrStorageNotFound
	}
	return f.getFn(id)
}

func (f *fakeStorage) Put(id, originalURL string) error {
	if f.putFn == nil {
		return nil
	}
	return f.putFn(id, originalURL)
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read error") }

func newTestHandler(st service.Storage) *Shortener {
	return &Shortener{
		svc:     service.NewShortener(st),
		baseURL: testBaseURL,
	}
}

func newCtx(w *httptest.ResponseRecorder, req *http.Request, params gin.Params) *gin.Context {
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = params
	return ctx
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestShorten(t *testing.T) {
	t.Run("valid url -> 201", func(t *testing.T) {
		var stored string
		st := &fakeStorage{
			putFn: func(id, originalURL string) error {
				stored = originalURL
				return nil
			},
		}
		h := newTestHandler(st)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
		h.Shorten(newCtx(w, req, nil))

		if w.Code != http.StatusCreated {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusCreated)
		}

		prefix := string(testBaseURL) + "/"
		body := w.Body.String()
		if !strings.HasPrefix(body, prefix) {
			t.Fatalf("body %q has no prefix %q", body, prefix)
		}
		if id := strings.TrimPrefix(body, prefix); len(id) != 8 {
			t.Fatalf("id length: got %d (%q), want 8", len(id), id)
		}
		if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
			t.Fatalf("content-type: got %q, want text/plain", ct)
		}
		if stored != "https://example.com" {
			t.Fatalf("stored url: got %q, want %q", stored, "https://example.com")
		}
	})

	t.Run("empty body -> 400", func(t *testing.T) {
		h := newTestHandler(&fakeStorage{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		h.Shorten(newCtx(w, req, nil))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("whitespace body -> 400", func(t *testing.T) {
		h := newTestHandler(&fakeStorage{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("   \n\t "))
		h.Shorten(newCtx(w, req, nil))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("garbage url -> 400", func(t *testing.T) {
		// putFn не задаём — до storage дело не дойдёт, сервис вернёт ErrInvalidURL.
		h := newTestHandler(&fakeStorage{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-a-url"))
		h.Shorten(newCtx(w, req, nil))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("storage put error -> 400", func(t *testing.T) {
		st := &fakeStorage{
			putFn: func(id, originalURL string) error { return errors.New("boom") },
		}
		h := newTestHandler(st)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
		h.Shorten(newCtx(w, req, nil))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("body read error -> 400", func(t *testing.T) {
		h := newTestHandler(&fakeStorage{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", errReader{})
		h.Shorten(newCtx(w, req, nil))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestResolve(t *testing.T) {
	t.Run("known id -> 307 + Location", func(t *testing.T) {
		const original = "https://example.com"
		st := &fakeStorage{
			getFn: func(id string) (string, error) { return original, nil },
		}
		h := newTestHandler(st)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/abc", nil)
		h.Resolve(newCtx(w, req, gin.Params{{Key: "id", Value: "abc"}}))

		if w.Code != http.StatusTemporaryRedirect {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusTemporaryRedirect)
		}
		if loc := w.Header().Get("Location"); loc != original {
			t.Fatalf("location: got %q, want %q", loc, original)
		}
	})

	t.Run("unknown id -> 400", func(t *testing.T) {
		st := &fakeStorage{
			getFn: func(id string) (string, error) { return "", service.ErrStorageNotFound },
		}
		h := newTestHandler(st)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		h.Resolve(newCtx(w, req, gin.Params{{Key: "id", Value: "missing"}}))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("empty id -> 400", func(t *testing.T) {
		h := newTestHandler(&fakeStorage{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.Resolve(newCtx(w, req, gin.Params{{Key: "id", Value: ""}}))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}
