package bcapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NII-DG/gogs/internal/jsonfunc"
)

var API_URL_GET_CONTENT_BY_FOLDER string = "getContentByFolder"

//[GET] /getContentByFolder
func GetContentByFolder(userCode, folderPath string) (jsonfunc.ResContentsInFolder, error) {
	//リクエスト生成
	req, _ := createNewRequest(http.MethodGet, API_URL_GET_CONTENT_BY_FOLDER, nil)
	//ヘッダー設定
	req.Header.Set("Authorization", "token")
	req.Header.Set("Content-Type", "application/json")
	//パラメータ設定
	params := req.URL.Query()
	params.Add("user_code", userCode)
	params.Add("folder_path", folderPath)
	req.URL.RawQuery = params.Encode()

	//通信実行
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return jsonfunc.ResContentsInFolder{}, fmt.Errorf("[HTTP conection error] url : %v, error msg : %v", API_URL_GET_CONTENT_BY_FOLDER, err)
	}
	defer resp.Body.Close()
	//リクエストボディの取得
	byteArray, _ := ioutil.ReadAll(resp.Body)
	jsonBytes := ([]byte)(byteArray)
	data := new(jsonfunc.ResContentsInFolder)
	if err := json.Unmarshal(jsonBytes, data); err != nil {
		return jsonfunc.ResContentsInFolder{}, fmt.Errorf("[JSON Unmarshal error] error msg : %v", err)
	}
	return *data, nil
}
