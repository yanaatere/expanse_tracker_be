package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yanaatere/expense_tracking/auth"
)

const (
	maxUploadSize = 5 << 20 // 5 MB
	uploadsDir    = "uploads/receipts"
)

var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// @Summary Upload receipt image
// @Description Upload a receipt image file (protected). Returns the stored URL.
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param receipt formData file true "Receipt image (JPEG/PNG/WebP, max 5 MB)"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/uploads/receipts [post]
func (h *UploadHandler) UploadReceipt(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		WriteError(w, http.StatusBadRequest, "File too large (max 5 MB)")
		return
	}

	file, header, err := r.FormFile("receipt")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Missing file field 'receipt'")
		return
	}
	defer file.Close()

	// Detect content type from first 512 bytes
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	ext, ok := allowedMimeTypes[contentType]
	if !ok {
		WriteError(w, http.StatusBadRequest, "Only JPEG, PNG, and WebP images are allowed")
		return
	}

	// Seek back to beginning after sniffing
	if seeker, ok2 := file.(io.ReadSeeker); ok2 {
		if _, err2 := seeker.Seek(0, io.SeekStart); err2 != nil {
			WriteError(w, http.StatusInternalServerError, "Failed to process file")
			return
		}
	}

	// Ensure uploads directory exists
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create upload directory")
		return
	}

	// Build a safe, unique filename
	baseName := sanitizeBaseName(header.Filename)
	filename := fmt.Sprintf("%d_%d_%s%s", userID, time.Now().UnixNano(), baseName, ext)
	fullPath := filepath.Join(uploadsDir, filename)

	dst, err := os.Create(fullPath)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	url := fmt.Sprintf("/uploads/receipts/%s", filename)
	WriteSuccess(w, http.StatusOK, map[string]string{"url": url})
}

// sanitizeBaseName strips the extension and keeps only safe characters (max 32).
func sanitizeBaseName(name string) string {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	if result == "" {
		result = "receipt"
	}
	if len(result) > 32 {
		result = result[:32]
	}
	return result
}
