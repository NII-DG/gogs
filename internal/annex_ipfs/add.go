package annex_ipfs

import (
	"encoding/json"
	"regexp"
	"unsafe"
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
	reg := "\r\n|\n"
	strMsg := *(*string)(unsafe.Pointer(rawJson))            //[]byte to string
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
	var data []AnnexAddResponse
	if err := json.Unmarshal(byteJson, &data); err != nil {
		return nil, err
	}
	return data, nil
}
