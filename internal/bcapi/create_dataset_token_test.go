package bcapi

import (
	"testing"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestResDatasetInfo_正常系(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	URL := conf.BcApiServer.ServerURL + API_URL_CREATE_DATASET_TOKEN

	resBpody := ResNotCreateDatasetToken{}
	resBpody.DatasetList = append(resBpody.DatasetList, struct {
		DatasetLocation string "json:\"dataset_location\""
		InputAddress    string "json:\"input_address\""
		SrcAddress      string "json:\"src_address\""
		OutputAddress   string "json:\"output_address\""
	}{
		"location1", "input1", "src1", "output1",
	})

	resder, _ := httpmock.NewJsonResponder(200, resBpody)

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

func TestResDatasetInfo_異常系(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	URL := conf.BcApiServer.ServerURL + API_URL_CREATE_DATASET_TOKEN

	resBpody := ResNotCreateDatasetToken{}
	resBpody.DatasetList = append(resBpody.DatasetList, struct {
		DatasetLocation string "json:\"dataset_location\""
		InputAddress    string "json:\"input_address\""
		SrcAddress      string "json:\"src_address\""
		OutputAddress   string "json:\"output_address\""
	}{
		"location1", "input1", "src1", "output1",
	})

	resder := httpmock.NewStringResponder(400, "mocked")

	httpmock.RegisterResponder("POST", URL, resder)

	contentMap := map[string]UploadDatasetInfo{}
	contentMap["dataset/test.txt"] = UploadDatasetInfo{
		InputAddress:  "fkjshfidjlfd",
		SrcAddress:    "kdfsjflkldsjgld",
		OutputAddress: "jnkfjdkflksjfdl",
	}

	_, err := CreateDatasetToken("user01", contentMap)

	assert.Error(t, err)
}
