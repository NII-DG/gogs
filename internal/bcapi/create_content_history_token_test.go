package bcapi

import (
	"testing"

	"github.com/ivis-yoshida/gogs/internal/conf"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateContentHistory(t *testing.T) {
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
