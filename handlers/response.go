package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
)

type ApiResponse struct {
	MsgID  string `json:"msgId"`
	Status string `json:"status"`
	Data   any    `json:"data"`
}

func generateMsgID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func WriteSuccess(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ApiResponse{
		MsgID:  generateMsgID(),
		Status: "success",
		Data:   data,
	})
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ApiResponse{
		MsgID:  generateMsgID(),
		Status: "error",
		Data:   map[string]string{"message": message},
	})
}
