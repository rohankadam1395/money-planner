package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/statement"
)

// UploadHandler handles statement file uploads
type UploadHandler struct {
	service *statement.StatementService
	logger  *logrus.Logger
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(service *statement.StatementService, logger *logrus.Logger) *UploadHandler {
	return &UploadHandler{
		service: service,
		logger:  logger,
	}
}

// Upload handles POST /api/statements/upload
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Upload handler called")

	if r.Method != http.MethodPost {
		h.logger.Warn("Invalid method: " + r.Method)
		middleware.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		return
	}

	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "user not authenticated", "UNAUTHORIZED")
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(50 * 1024 * 1024) // 50MB max
	if err != nil {
		middleware.WriteJSONErrorWithMessage(w, http.StatusBadRequest,
			"invalid multipart form", fmt.Sprintf("failed to parse form: %v", err), "INVALID_FORM")
		return
	}

	// Get file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "file field required", "MISSING_FILE")
		return
	}
	defer file.Close()

	// Get bank code from form
	bankCode := r.FormValue("bank_code")
	if bankCode == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "bank_code field required", "MISSING_BANK_CODE")
		return
	}

	// Read file content
	fileContent := make([]byte, handler.Size)
	if n, err := file.Read(fileContent); err != nil || int64(n) != handler.Size {
		middleware.WriteJSONError(w, http.StatusBadRequest, "failed to read file", "FILE_READ_ERROR")
		return
	}

	// Determine file format from filename
	fileFormat := getFileFormat(handler.Filename)
	if fileFormat == "" {
		middleware.WriteJSONErrorWithMessage(w, http.StatusBadRequest,
			"unsupported file format", fmt.Sprintf("file %s has unsupported format", handler.Filename), "UNSUPPORTED_FORMAT")
		return
	}

	// Call service to handle upload
	uploadReq := &statement.UploadRequest{
		FileContent: fileContent,
		FileName:    handler.Filename,
		FileFormat:  fileFormat,
		BankCode:    bankCode,
		UserID:      userID,
	}

	h.logger.WithFields(logrus.Fields{
		"fileName":   handler.Filename,
		"fileFormat": fileFormat,
		"bankCode":   bankCode,
		"userID":     userID,
		"fileSize":   len(fileContent),
	}).Info("Calling service.Upload")

	resp, err := h.service.Upload(uploadReq)
	if err != nil {
		h.logger.WithError(err).Error("Upload failed")
		// Check if it's a duplicate
		if strings.Contains(err.Error(), "duplicate") {
			middleware.WriteJSONError(w, http.StatusConflict, err.Error(), "DUPLICATE_STATEMENT")
			return
		}

		middleware.WriteJSONErrorWithMessage(w, http.StatusInternalServerError,
			"upload failed", err.Error(), "UPLOAD_FAILED")
		return
	}

	h.logger.WithField("statementID", resp.StatementID).Info("Upload successful")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted per spec
	json.NewEncoder(w).Encode(resp)
}

// getFileFormat determines file format from filename
func getFileFormat(filename string) string {
	filename = strings.ToLower(filename)

	if strings.HasSuffix(filename, ".pdf") {
		return "PDF"
	} else if strings.HasSuffix(filename, ".csv") {
		return "CSV"
	} else if strings.HasSuffix(filename, ".xlsx") {
		return "XLSX"
	} else if strings.HasSuffix(filename, ".xls") {
		return "XLSX" // Treat old Excel as XLSX for now
	}

	return ""
}
