package annex_ipfs

import (
	"encoding/json"
	"regexp"
	"unsafe"

	//logv2 "unknwon.dev/clog/v2"
	"github.com/NII-DG/gogs/internal/jsonfunc"
)

//git annex add --to --jsonの構造体

type AnnexAddResponse struct {
	Command string `json:"command"`
	Note    string `json:"note"`
	Success bool   `json:"success"`
	Key     string `json:"key"`
	File    string `json:"file"`
}

func GetAnnexAddInfo(rawJson *[]byte) ([]AnnexAddResponse, error) {
	annexAddResponseList := []AnnexAddResponse{}
	reg := "\r\n|\n"
	strMsg := *(*string)(unsafe.Pointer(rawJson))            //[]byte to string
	splitByline := regexp.MustCompile(reg).Split(strMsg, -1) //改行分割
	for _, unitData := range splitByline {
		if jsonfunc.IsJSONString(unitData) {
			byteJson := []byte(unitData)
			var data AnnexAddResponse
			if err := json.Unmarshal(byteJson, &data); err != nil {
				return nil, err
			}
			annexAddResponseList = append(annexAddResponseList, data)
		}
	}
	return annexAddResponseList, nil
}
