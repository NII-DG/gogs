package bcapi

import (
	"testing"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateContentHistory_正常系(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	URL := conf.BcApiServer.ServerURL + API_URL_CREATE_CONTENT_HISTORY_TOKEN

	httpmock.RegisterResponder("POST", URL,
		httpmock.NewStringResponder(200, "mocked"),
	)

	contentMap := map[string]string{}
	contentMap["dataset/test.txt"] = "QmIKDHIJIFLF"
	err := CreateContentHistory("user01", contentMap)

	assert.NoError(t, err)

}

func TestCreateContentHistory_異常系(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	URL := conf.BcApiServer.ServerURL + API_URL_CREATE_CONTENT_HISTORY_TOKEN

	httpmock.RegisterResponder("POST", URL,
		httpmock.NewStringResponder(2001, "mocked"),
	)

	contentMap := map[string]string{}
	contentMap["dataset/test.txt"] = "QmIKDHIJIFLF"
	err := CreateContentHistory("user01", contentMap)

	assert.Error(t, err)

}
