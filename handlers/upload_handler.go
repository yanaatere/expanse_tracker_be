package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/yanaatere/expense_tracking/auth"
)

const (
	maxUploadSize  = 5 << 20 // 5 MB
	signedURLExpiry = time.Hour
)

var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type UploadHandler struct {
	minio  *minio.Client
	bucket string
}

func NewUploadHandler(minioClient *minio.Client, bucket string) *UploadHandler {
	return &UploadHandler{
		minio:  minioClient,
		bucket: bucket,
	}
}

// @Summary Upload receipt image
// @Description Upload a receipt image file (protected). Returns object_name and a 1-hour pre-signed URL.
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

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to read file")
		return
	}
	data := buf.Bytes()

	// Detect content type from first 512 bytes (ignores the Content-Type header).
	contentType := http.DetectContentType(data[:min(512, len(data))])
	ext, ok := allowedMimeTypes[contentType]
	if !ok {
		WriteError(w, http.StatusBadRequest, "Only JPEG, PNG, and WebP images are allowed")
		return
	}

	baseName := sanitizeBaseName(header.Filename)
	objectName := fmt.Sprintf("%d_%d_%s%s", userID, time.Now().UnixNano(), baseName, ext)

	_, err = h.minio.PutObject(
		context.Background(),
		h.bucket,
		objectName,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to upload file")
		return
	}

	signedURL, err := h.minio.PresignedGetObject(context.Background(), h.bucket, objectName, signedURLExpiry, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate signed URL")
		return
	}

	WriteSuccess(w, http.StatusOK, map[string]string{
		"object_name": objectName,
		"url":         signedURL.String(),
	})
}

// @Summary Get signed receipt URL
// @Description Generate a 1-hour pre-signed URL for an owned receipt (protected).
// @Tags Upload
// @Produce json
// @Param objectName path string true "Object name returned from upload"
// @Success 200 {object} object
// @Failure 401 {object} MessageResponse
// @Failure 403 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/uploads/receipts/{objectName} [get]
func (h *UploadHandler) GetSignedReceiptURL(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	objectName := mux.Vars(r)["objectName"]
	if objectName == "" {
		WriteError(w, http.StatusBadRequest, "Missing object name")
		return
	}

	ownerPrefix := strconv.Itoa(int(userID)) + "_"
	if !strings.HasPrefix(objectName, ownerPrefix) {
		WriteError(w, http.StatusForbidden, "Forbidden")
		return
	}

	signedURL, err := h.minio.PresignedGetObject(context.Background(), h.bucket, objectName, signedURLExpiry, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate signed URL")
		return
	}

	WriteSuccess(w, http.StatusOK, map[string]string{"url": signedURL.String()})
}

// @Summary Delete receipt image
// @Description Delete a previously uploaded receipt (protected). Only the owner can delete.
// @Tags Upload
// @Produce json
// @Param objectName path string true "Object name returned from upload"
// @Success 200 {object} object
// @Failure 401 {object} MessageResponse
// @Failure 403 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/uploads/receipts/{objectName} [delete]
func (h *UploadHandler) DeleteReceipt(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	objectName := mux.Vars(r)["objectName"]
	if objectName == "" {
		WriteError(w, http.StatusBadRequest, "Missing object name")
		return
	}

	ownerPrefix := strconv.Itoa(int(userID)) + "_"
	if !strings.HasPrefix(objectName, ownerPrefix) {
		WriteError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err := h.minio.RemoveObject(context.Background(), h.bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete file")
		return
	}

	WriteSuccess(w, http.StatusOK, map[string]string{"deleted": objectName})
}

// sanitizeBaseName strips the extension and keeps only safe characters (max 32).
func sanitizeBaseName(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[:idx]
	}
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
