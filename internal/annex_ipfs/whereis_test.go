package annex_ipfs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockAnnexWhereResponse struct {
	Command   string   `json:"command"`
	Note      string   `json:"note"`
	Success   bool     `json:"success"`
	Untrusted []string `json:"untrusted"`
	Key       string   `json:"key"`
	Whereis   []struct {
		Here        bool     `json:"here"`
		Uuid        string   `json:"uuid"`
		Urls        []string `json:"urls"`
		Description string   `json:"description"`
	} `json:"whereis"`
	File string `json:"file"`
}

func TestGetAnnexContentInfo_正常系(t *testing.T) {
	mockJson := MockAnnexWhereResponse{}
	mockJson.Command = "Whereis"
	mockJson.Note = "whereis_note"
	mockJson.Success = true
	mockJson.Untrusted = append(mockJson.Untrusted, "where_Untrusted")
	mockJson.Key = "where_key"
	mockJson.File = "test.txt"
	mockJson.Whereis = append(mockJson.Whereis, struct {
		Here        bool     "json:\"here\""
		Uuid        string   "json:\"uuid\""
		Urls        []string "json:\"urls\""
		Description string   "json:\"description\""
	}{true, "whreis_uuid", []string{"ipfs:url"}, "ipfs where_Description"})

	s, _ := json.Marshal(mockJson)

	res, err := GetAnnexContentInfo(&s)
	assert.NoError(t, err)
	assert.Equal(t, mockJson.File, res.FileNm)
	assert.Equal(t, "url", res.IpfsCid)
	assert.Equal(t, mockJson.Key, res.Key)
}
