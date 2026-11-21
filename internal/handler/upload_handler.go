package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxUploadBytes  = 8 << 20 // 8 MB
	maxFormOverhead = 1 << 20 // margem para o envelope multipart
	maxSniffBytes   = 512
)

var extPorTipoImagem = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// UploadHandler recebe uma imagem via multipart/form-data, salva em disco e
// devolve a URL pública do arquivo. A leitura é servida por StaticUploads.
type UploadHandler struct {
	dir string
}

func NewUploadHandler(dir string) *UploadHandler {
	if strings.TrimSpace(dir) == "" {
		dir = "./uploads"
	}
	return &UploadHandler{dir: dir}
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+maxFormOverhead)

	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "Arquivo grande demais (máximo 8 MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Formulário multipart inválido", http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Campo 'file' é obrigatório", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buf := make([]byte, maxSniffBytes)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		http.Error(w, "Erro ao ler arquivo", http.StatusInternalServerError)
		return
	}

	ext, ok := extPorTipoImagem[http.DetectContentType(buf[:n])]
	if !ok {
		http.Error(w, "Formato não suportado (use jpg, png, webp ou gif)", http.StatusBadRequest)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Erro ao ler arquivo", http.StatusInternalServerError)
		return
	}

	if err := os.MkdirAll(h.dir, 0o755); err != nil {
		http.Error(w, "Erro ao preparar diretório de uploads", http.StatusInternalServerError)
		return
	}

	nome, err := nomeArquivoAleatorio(ext)
	if err != nil {
		http.Error(w, "Erro ao gerar nome do arquivo", http.StatusInternalServerError)
		return
	}

	destino := filepath.Join(h.dir, nome)
	out, err := os.OpenFile(destino, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		http.Error(w, "Erro ao salvar arquivo", http.StatusInternalServerError)
		return
	}

	copiados, err := io.Copy(out, io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		out.Close()
		os.Remove(destino)
		http.Error(w, "Erro ao salvar arquivo", http.StatusInternalServerError)
		return
	}
	if err := out.Close(); err != nil {
		os.Remove(destino)
		http.Error(w, "Erro ao salvar arquivo", http.StatusInternalServerError)
		return
	}
	if copiados > maxUploadBytes {
		os.Remove(destino)
		http.Error(w, "Arquivo grande demais (máximo 8 MB)", http.StatusRequestEntityTooLarge)
		return
	}

	url := baseURLPublica(r) + "/uploads/" + nome

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

// StaticUploads serve os arquivos de upload. Fica fora do middleware de
// autenticação de propósito: <img src> não envia o header Authorization.
func StaticUploads(dir string) http.Handler {
	if strings.TrimSpace(dir) == "" {
		dir = "./uploads"
	}

	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		fs.ServeHTTP(w, r)
	})
}

func baseURLPublica(r *http.Request) string {
	if base := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/"); base != "" {
		return base
	}

	esquema := "http"
	if r.TLS != nil ||
		strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") ||
		strings.HasSuffix(strings.ToLower(r.Host), ".ts.net") {
		esquema = "https"
	}
	return esquema + "://" + r.Host
}

func nomeArquivoAleatorio(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + ext, nil
}
