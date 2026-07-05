package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

type LoggingMiddleware struct {
	logger *logrus.Logger
}

func NewLoggingMiddleware(logger *logrus.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.RequestURI,
			"remote": r.RemoteAddr,
		}).Info("incoming request")

		next.ServeHTTP(w, r)
	})
}

func WriteJSONError(w http.ResponseWriter, statusCode int, errMsg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   errMsg,
		Code:    code,
		Message: http.StatusText(statusCode),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Fprintf(w, `{"error":"failed to encode error response"}`)
	}
}

func WriteJSONErrorWithMessage(w http.ResponseWriter, statusCode int, errMsg, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   errMsg,
		Message: message,
		Code:    code,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Fprintf(w, `{"error":"failed to encode error response"}`)
	}
}
