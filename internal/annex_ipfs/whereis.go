package annex_ipfs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unsafe"

	"github.com/NII-DG/gogs/internal/jsonfunc"
	"github.com/gogs/git-module"
	log "unknwon.dev/clog/v2"
	//log "unknwon.dev/clog/v2"
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
	Key     string
	IpfsCid string
	FileNm  string
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
	info.IpfsCid = a.getHashValueInIPFS()
	info.FileNm = a.getFile()
	return *info
}

func GetAnnexContentInfo(rawJson *[]byte) (AnnexContentInfo, error) {
	var data AnnexWhereResponse
	if err := json.Unmarshal(*rawJson, &data); err != nil {
		return AnnexContentInfo{}, err
	}

	return data.getAnnexContentInfo(), nil
}

//データセット名に対するAnnxeのメタ情報をMapを取得
func GetAnnexContentInfoListByDatasetNm(rawJson *[]byte, datasetNmList []string) (map[string][]AnnexContentInfo, error) {
	annexContentInfoMap := map[string][]AnnexContentInfo{}
	reg := "\r\n|\n"
	strJson := *(*string)(unsafe.Pointer(rawJson))            //[]byte to string
	splitByline := regexp.MustCompile(reg).Split(strJson, -1) //改行分割
	for _, unitData := range splitByline {
		if jsonfunc.IsJSONString(unitData) {
			byteJson := []byte(unitData)
			var data AnnexWhereResponse
			if err := json.Unmarshal(byteJson, &data); err != nil {
				return nil, err
			}
			for _, datasetNm := range datasetNmList {
				if isContainDatasetNm(data.getFile(), datasetNm) {
					annexContentInfoMap[datasetNm] = append(annexContentInfoMap[datasetNm], data.getAnnexContentInfo())
				}
			}
		}
	}
	return annexContentInfoMap, nil
}

func isContainDatasetNm(fileNm string, datasetNm string) bool {
	//FileNm(datasetフォルダーからのパス[datasetNm/folder/..../file])の左に指定データセット名があること
	return strings.HasPrefix(fileNm, datasetNm)
}

//git annex whereis (複数：JSON形式)をコンテンツロケーション名と一致するAnnexキーを取得
//
//@parame rawJson *[]byte git annex whereis のレスポンス
//
//@parame contentLocationList []string　抽出対象のコンテンツロケーション（ファイルパス）　完全一致の時のみ抽出
//
//@parame upperPath string　git annex whereis のレスポンスのファイル名項目の上部に追加する情報。 ない("")場合は、付けない
//　　　　　　　　　　　　　　ex: /OwenerNm/RepoNM/BranchNm
//　　　　　　　　　　　　　　ない("")場合は、付けない
func GetAnnexKeyListToContentLoc(rawJson *[]byte, contentLocationList []string, upperPath string) ([]string, error) {
	var keyList []string
	dataList, err := resolveAnnexWhereisResponseList(rawJson, upperPath)
	if err != nil {
		return nil, err
	}
	for _, location := range contentLocationList {
		data := dataList[location]
		keyList = append(keyList, data.Key)
	}
	return keyList, nil
}

//git annex whereis (複数：JSON形式)をファイル名と各コンテンツ情報を紐づける
func resolveAnnexWhereisResponseList(rawJson *[]byte, upperPath string) (map[string]AnnexWhereResponse, error) {
	fileNmMapToRes := map[string]AnnexWhereResponse{}
	reg := "\r\n|\n"
	strJson := *(*string)(unsafe.Pointer(rawJson))            //[]byte to string
	splitByline := regexp.MustCompile(reg).Split(strJson, -1) //改行分割
	for _, unitData := range splitByline {
		if jsonfunc.IsJSONString(unitData) {
			byteJson := []byte(unitData)
			var data AnnexWhereResponse
			if err := json.Unmarshal(byteJson, &data); err != nil {
				return nil, err
			}
			fileNm := data.getFile()
			if len(upperPath) > 0 {
				s := &fileNm
				*s = filepath.Join(upperPath, fileNm)
			}
			fileNmMapToRes[fileNm] = data
		}
	}
	return fileNmMapToRes, nil
}

func WhereisByKey(repoPath, key string) (AnnexContentInfo, error) {
	msg, err := git.NewCommand("annex", "whereis", "--json", "--key", key).RunInDir(repoPath)
	if err != nil {
		return AnnexContentInfo{}, fmt.Errorf("Failure git annex whereis by key[%v]", key)
	}
	log.Trace("[GIT-A WHEREIS Msg] %v", string(msg))
	return GetAnnexContentInfo(&msg)

}
