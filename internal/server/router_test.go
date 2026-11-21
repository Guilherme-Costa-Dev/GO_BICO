package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRouterServeUploadsSemAutenticacao(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "abc.jpg"), []byte{0xff, 0xd8, 0xff}, 0o644); err != nil {
		t.Fatalf("preparar arquivo: %v", err)
	}
	t.Setenv("UPLOAD_DIR", dir)

	api := New(nil, nil)

	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/uploads/abc.jpg", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /uploads/abc.jpg: esperado 200, veio %d", rec.Code)
	}
}

func TestRouterBloqueiaUploadSemToken(t *testing.T) {
	t.Setenv("UPLOAD_DIR", t.TempDir())

	api := New(nil, nil)

	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/uploads", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /uploads sem token: esperado 401, veio %d", rec.Code)
	}
}
