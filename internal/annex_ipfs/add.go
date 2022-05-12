package annex_ipfs

import (
	"encoding/json"
	"fmt"
	"regexp"
	"unsafe"

	//logv2 "unknwon.dev/clog/v2"
	"github.com/NII-DG/gogs/internal/jsonfunc"
	"github.com/gogs/git-module"
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

func Add(repoPath string, all bool, files ...string) ([]AnnexAddResponse, error) {
	cmd := git.NewCommand("annex", "add", "--json")
	if all {
		cmd.AddArgs(".")
	}
	msg, err := cmd.AddArgs(files...).RunInDir(repoPath)
	if err == nil {
		reslist, err := GetAnnexAddInfo(&msg)
		if err != nil {
			return nil, fmt.Errorf("[Annex Add Json Error]: %v", err)
		}
		return reslist, nil
	}
	return nil, err
}

func AddByFileNm(repoPath string, fileNm string) (AnnexAddResponse, error) {
	res, err := Add(repoPath, false, fileNm)
	if err != nil {
		return AnnexAddResponse{}, err
	} else if len(res) > 1 {
		return AnnexAddResponse{}, fmt.Errorf("get multiple info by git annex add <file Name>")
	}
	return res[0], err
}
