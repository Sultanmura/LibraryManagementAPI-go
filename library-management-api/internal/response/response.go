package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	StatusCode   int         `json:"statusCode"`
	IsSuccess    bool        `json:"isSuccess"`
	Result       interface{} `json:"result"`
	ErrorCode    *string     `json:"errorCode"`
	ErrorMessage *string     `json:"errorMessage"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	resp := APIResponse{
		StatusCode:   status,
		IsSuccess:    true,
		Result:       data,
		ErrorCode:    nil,
		ErrorMessage: nil,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func Error(w http.ResponseWriter, status int, errCode, errMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	resp := APIResponse{
		StatusCode:   status,
		IsSuccess:    false,
		Result:       nil,
		ErrorCode:    &errCode,
		ErrorMessage: &errMsg,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
