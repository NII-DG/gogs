package bcapi

import (
	"testing"

	"github.com/ivis-yoshida/gogs/internal/conf"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGetDatasetInfoByLocation_正常系(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	URL := conf.BcApiServer.ServerURL + API_URL_GET_DATASET_INFO_BY_LOCATION

	resBody := ResDatasetInfo{}
	resBody.DatasetLocation = "dksjfklsjfkdskf"
	resBody.InputAddress = "dkjflkslfdkmlskdfjns"
	resBody.SrcCodeAddress = "nkdkskjdnfkksjfd"
	resBody.OutputAddress = "nfsndfkjnslkdfkjsnfd"

	resder, _ := httpmock.NewJsonResponder(200, resBody)
	httpmock.RegisterResponder("GET", URL, resder)

	res, err := GetDatasetInfoByLocation("user01", "datasetLocation")

	assert.NoError(t, err)
	assert.Equal(t, resBody.DatasetLocation, res.DatasetLocation)
	assert.Equal(t, resBody.InputAddress, res.InputAddress)
	assert.Equal(t, resBody.SrcCodeAddress, res.SrcCodeAddress)
	assert.Equal(t, resBody.OutputAddress, res.OutputAddress)
}

func TestGetDatasetInfoByLocation_異常系(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	URL := conf.BcApiServer.ServerURL + API_URL_GET_DATASET_INFO_BY_LOCATION

	resder := httpmock.NewStringResponder(400, "mocked")
	httpmock.RegisterResponder("GET", URL, resder)

	res, err := GetDatasetInfoByLocation("user01", "datasetLocation")

	assert.Error(t, err)
	assert.Empty(t, res.DatasetLocation)
	assert.Empty(t, res.InputAddress)
	assert.Empty(t, res.SrcCodeAddress)
	assert.Empty(t, res.OutputAddress)
}
