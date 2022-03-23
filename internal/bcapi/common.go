package bcapi

import (
	"bytes"
	"net/http"

	"github.com/NII-DG/gogs/internal/conf"
)

func createNewRequest(httpMethod, urlPath string, reqBody []byte) (*http.Request, error) {

	fullUrl := conf.BcApiServer.ServerURL + urlPath
	if reqBody != nil {
		return http.NewRequest(httpMethod, fullUrl, bytes.NewReader(reqBody))
	} else {
		return http.NewRequest(httpMethod, fullUrl, nil)
	}
}
