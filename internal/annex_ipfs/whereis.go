package annex_ipfs

import (
	"encoding/json"
	"regexp"
	"strings"
	"unsafe"
)

//git annex whereis --jsonの構造体
type AnnexWhereResponse struct {
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

type AnnexWhereResponses []AnnexWhereResponse
type AnnexContentInfo struct {
	Key  string
	Hash string
	File string
}

func (a AnnexWhereResponse) getKey() string {
	return a.Key
}

func (a AnnexWhereResponse) getFile() string {
	return a.File
}

func (a AnnexWhereResponse) getHashValueInIPFS() string {
	whereis := a.Whereis
	var hash string
	for _, w := range whereis {
		if strings.Contains(w.Description, "ipfs") {
			u := w.Urls[0]
			index := strings.LastIndex(u, ":")
			url := &hash
			*url = u[index+1:]
		}
	}
	return hash
}

//key,hash,file_nameの3組を返す
func (a AnnexWhereResponse) getAnnexContentInfo() AnnexContentInfo {
	info := new(AnnexContentInfo)
	info.Key = a.getKey()
	info.Hash = a.getHashValueInIPFS()
	info.File = a.getFile()
	return *info
}

func GetAnnexContentInfoList(msgWhereis *[]byte) ([]AnnexContentInfo, error) {
	reg := "\r\n|\n"
	strMsg := *(*string)(unsafe.Pointer(msgWhereis))         //[]byte to string
	splitByline := regexp.MustCompile(reg).Split(strMsg, -1) //改行分割
	strJson := "["
	for index := 1; index < len(splitByline)-1; index++ {
		if index == len(splitByline)-2 {
			strJson = strJson + splitByline[index]
			strJson = strJson + "]"
		} else {
			strJson = strJson + splitByline[index]
			strJson = strJson + ","
		}
	}
	byteJson := []byte(strJson)
	var data []AnnexWhereResponse
	if err := json.Unmarshal(byteJson, &data); err != nil {
		return nil, err
	}

	contentInfoList := []AnnexContentInfo{}
	for _, d := range data {
		contentInfoList = append(contentInfoList, d.getAnnexContentInfo())
	}
	return contentInfoList, nil
}
