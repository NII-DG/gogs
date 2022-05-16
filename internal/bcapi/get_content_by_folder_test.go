package bcapi

import (
	"testing"
	"time"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/jsonfunc"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestGetContentByFolder_ok(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	URL := conf.BcApiServer.ServerURL + API_URL_GET_CONTENT_BY_FOLDER

	now := time.Now()
	resBody := jsonfunc.ResContentsInFolder{}
	resBody.ContentsInFolder = append(resBody.ContentsInFolder, struct {
		UserCode        string    "json:\"user_code\""
		ContentLocation string    "json:\"content_location\""
		FullContentHash string    "json:\"full_content_hash\""
		IpfsCid         string    "json:\"ipfs_cid\""
		IsPrivate       bool      "json:\"is_private\""
		AddDateTime     time.Time "json:\"add_date_time\""
	}{
		"usr01",
		"location01",
		"",
		"fjlksjflkdmlkfjd",
		false,
		now,
	})

	resder, _ := httpmock.NewJsonResponder(200, resBody)
	httpmock.RegisterResponder("GET", URL, resder)

	res, err := GetContentByFolder("usr01", "/dataset/test")

	assert.NoError(t, err)
	assert.Equal(t, resBody.ContentsInFolder[0].UserCode, res.ContentsInFolder[0].UserCode)
	assert.Equal(t, resBody.ContentsInFolder[0].ContentLocation, res.ContentsInFolder[0].ContentLocation)
	assert.Equal(t, resBody.ContentsInFolder[0].IpfsCid, res.ContentsInFolder[0].IpfsCid)
}

func TestGetContentByFolder_NG(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	URL := conf.BcApiServer.ServerURL + API_URL_GET_CONTENT_BY_FOLDER

	resder := httpmock.NewStringResponder(400, "mocked")
	httpmock.RegisterResponder("GET", URL, resder)

	_, err := GetContentByFolder("usr01", "/dataset/test")

	assert.Error(t, err)
}
