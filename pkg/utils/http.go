package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	// Errors happened during request.
	Errors []string `json:"errors,omitempty"`
}

func ResponseBody(obj interface{}) []byte {
	respBody, err := json.Marshal(obj)
	if err != nil {
		return []byte(`{\"errors\":[\"Failed to parse response body\"]}`)
	}
	return respBody
}

func ResponseOKWithBody(rw http.ResponseWriter, obj interface{}) {
	rw.Header().Set("Content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write(ResponseBody(obj))
	if err != nil {
		logrus.Errorf("failed to write response body: %v", err)
	}
}

func ResponseError(rw http.ResponseWriter, statusCode int, err error) {
	ResponseErrorMsg(rw, statusCode, err.Error())
}

func ResponseErrorMsg(rw http.ResponseWriter, statusCode int, errMsg string) {
	rw.WriteHeader(statusCode)
	_, _ = rw.Write(ResponseBody(ErrorResponse{Errors: []string{errMsg}}))
}
