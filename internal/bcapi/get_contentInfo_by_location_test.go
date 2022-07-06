package bcapi

import (
	"testing"
	"time"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGetContentInfoByLocation_正常系(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	URL := conf.BcApiServer.ServerURL + API_URL_GET_CONTENT_INFO_BY_LOCATION

	resBody := ResContentInfo{}
	resBody.UserCode = "usr01"
	resBody.FullContentHash = ""
	resBody.IpfsCid = "fnslksjdflkfk"
	resBody.AddDateTime = time.Now()

	resder, _ := httpmock.NewJsonResponder(200, resBody)
	httpmock.RegisterResponder("GET", URL, resder)

	res, err := GetContentInfoByLocation("user01", "filesLocation")

	assert.NoError(t, err)
	assert.Equal(t, resBody.UserCode, res.UserCode)
	assert.Equal(t, resBody.IpfsCid, res.IpfsCid)
}

func TestGetContentInfoByLocation_異常系(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	URL := conf.BcApiServer.ServerURL + API_URL_GET_CONTENT_INFO_BY_LOCATION

	resBody := ResContentInfo{}
	resBody.UserCode = "usr01"
	resBody.FullContentHash = ""
	resBody.IpfsCid = "fnslksjdflkfk"
	resBody.AddDateTime = time.Now()

	resder := httpmock.NewStringResponder(400, "mocked")
	httpmock.RegisterResponder("GET", URL, resder)

	res, err := GetContentInfoByLocation("user01", "filesLocation")

	assert.Error(t, err)
	assert.Empty(t, res.UserCode)
	assert.Empty(t, res.IpfsCid)
}
