package bcapi

import (
	"testing"

	"github.com/ivis-yoshida/gogs/internal/conf"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestResDatasetInfo_正常系(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	URL := conf.BcApiServer.ServerURL + API_URL_CREATE_DATASET_TOKEN

	resBpody := ResNotCreateDatasetToken{}
	resBpody.DatasetList[0].DatasetLocation = "location1"
	resBpody.DatasetList[0].InputAddress = "input1"
	resBpody.DatasetList[0].SrcAddress = "src1"
	resBpody.DatasetList[0].OutputAddress = "output1"
	listBody := []ResNotCreateDatasetToken{resBpody}

	resder, _ := httpmock.NewJsonResponder(200, listBody)

	httpmock.RegisterResponder("POST", URL, resder)

	contentMap := map[string]UploadDatasetInfo{}
	contentMap["dataset/test.txt"] = UploadDatasetInfo{
		InputAddress:  "fkjshfidjlfd",
		SrcAddress:    "kdfsjflkldsjgld",
		OutputAddress: "jnkfjdkflksjfdl",
	}

	res, err := CreateDatasetToken("user01", contentMap)

	assert.NoError(t, err)
	assert.Equal(t, resBpody.DatasetList[0].DatasetLocation, res.DatasetList[0].DatasetLocation)
	assert.Equal(t, resBpody.DatasetList[0].InputAddress, res.DatasetList[0].InputAddress)
	assert.Equal(t, resBpody.DatasetList[0].SrcAddress, res.DatasetList[0].SrcAddress)
	assert.Equal(t, resBpody.DatasetList[0].OutputAddress, res.DatasetList[0].OutputAddress)

}
