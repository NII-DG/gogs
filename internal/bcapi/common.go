package bcapi

import (
	"bytes"
	"net/http"
)

var ServerPath string = "http://localhost8080/"

func createNewRequest(httpMethod, urlPath string, reqBody []byte) (*http.Request, error) {

	fullUrl := ServerPath + urlPath
	if reqBody != nil {
		return http.NewRequest(httpMethod, fullUrl, bytes.NewReader(reqBody))
	} else {
		return http.NewRequest(httpMethod, fullUrl, nil)
	}
}
