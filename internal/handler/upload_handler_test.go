package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var jpegMinimo = []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}

func novoUploadRequest(t *testing.T, nome string, conteudo []byte) *http.Request {
	t.Helper()

	var corpo bytes.Buffer
	writer := multipart.NewWriter(&corpo)
	parte, err := writer.CreateFormFile("file", nome)
	if err != nil {
		t.Fatalf("criar parte do form: %v", err)
	}
	if _, err := parte.Write(conteudo); err != nil {
		t.Fatalf("escrever conteúdo: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("fechar form: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/uploads", &corpo)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestUploadSalvaImagemNoDisco(t *testing.T) {
	dir := t.TempDir()
	h := NewUploadHandler(dir)

	rec := httptest.NewRecorder()
	h.Upload(rec, novoUploadRequest(t, "foto.jpg", jpegMinimo))

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, veio %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("resposta inválida: %v (%s)", err, rec.Body.String())
	}

	const prefixo = "/uploads/"
	i := strings.LastIndex(resp.URL, prefixo)
	if i < 0 {
		t.Fatalf("url sem prefixo %q: %q", prefixo, resp.URL)
	}
	nome := resp.URL[i+len(prefixo):]
	if nome == "" || strings.ContainsAny(nome, `/\`) || !strings.HasSuffix(nome, ".jpg") {
		t.Fatalf("nome de arquivo inválido na url: %q", nome)
	}

	salvo, err := os.ReadFile(filepath.Join(dir, nome))
	if err != nil {
		t.Fatalf("arquivo não salvo em disco: %v", err)
	}
	if !bytes.Equal(salvo, jpegMinimo) {
		t.Errorf("conteúdo salvo difere do enviado")
	}
}

func TestUploadUsaBaseURLPublica(t *testing.T) {
	t.Setenv("PUBLIC_BASE_URL", "https://bico.exemplo.ts.net/")
	h := NewUploadHandler(t.TempDir())

	rec := httptest.NewRecorder()
	h.Upload(rec, novoUploadRequest(t, "foto.jpg", jpegMinimo))

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, veio %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("resposta inválida: %v", err)
	}
	if !strings.HasPrefix(resp.URL, "https://bico.exemplo.ts.net/uploads/") {
		t.Errorf("url: esperado prefixo de PUBLIC_BASE_URL, veio %q", resp.URL)
	}
}

func TestUploadInfereHTTPSNoFunnel(t *testing.T) {
	t.Setenv("PUBLIC_BASE_URL", "")
	h := NewUploadHandler(t.TempDir())

	req := novoUploadRequest(t, "foto.jpg", jpegMinimo)
	req.Host = "bico.exemplo.ts.net"
	req.Header.Set("X-Forwarded-Proto", "https")

	rec := httptest.NewRecorder()
	h.Upload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, veio %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("resposta inválida: %v", err)
	}
	if !strings.HasPrefix(resp.URL, "https://bico.exemplo.ts.net/uploads/") {
		t.Errorf("url: esperado https, veio %q", resp.URL)
	}
}

func TestUploadRejeitaFormatoNaoSuportado(t *testing.T) {
	h := NewUploadHandler(t.TempDir())

	rec := httptest.NewRecorder()
	h.Upload(rec, novoUploadRequest(t, "arquivo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUploadRejeitaSemArquivo(t *testing.T) {
	h := NewUploadHandler(t.TempDir())

	var corpo bytes.Buffer
	writer := multipart.NewWriter(&corpo)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/uploads", &corpo)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	h.Upload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUploadRejeitaArquivoAcimaDoLimite(t *testing.T) {
	h := NewUploadHandler(t.TempDir())

	grande := make([]byte, maxUploadBytes+1)
	copy(grande, jpegMinimo)

	rec := httptest.NewRecorder()
	h.Upload(rec, novoUploadRequest(t, "grande.jpg", grande))

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("esperado 413, veio %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStaticUploadsServeArquivo(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "abc.jpg"), jpegMinimo, 0o644); err != nil {
		t.Fatalf("preparar arquivo: %v", err)
	}

	h := StaticUploads(dir)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/abc.jpg", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d", rec.Code)
	}
	if rec.Body.Len() != len(jpegMinimo) {
		t.Errorf("tamanho servido: esperado %d, veio %d", len(jpegMinimo), rec.Body.Len())
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("esperado X-Content-Type-Options: nosniff")
	}
}

func TestStaticUploadsNaoListaDiretorio(t *testing.T) {
	h := StaticUploads(t.TempDir())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, veio %d", rec.Code)
	}
}
