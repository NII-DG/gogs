package bcapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var API_URL_GET_CONTENT_BY_FOLDER string = "getContentByFolder"

type ResContentsInFolder struct {
	ContentsInFolder []struct {
		UserCode        string    `json:"user_code"`
		ContentLocation string    `json:"content_location"`
		ContentAddress  string    `json:"content_address"`
		AddDateTime     time.Time `json:"add_date_time"`
	} `json:"contents_in_folder"`
}

//[GET] /getContentByFolder
func GetContentByFolder(userCode, folderPath string) (ResContentsInFolder, error) {
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
		return ResContentsInFolder{}, fmt.Errorf("[HTTP conection error] url : %v, error msg : %v", API_URL_GET_CONTENT_BY_FOLDER, err)
	}
	defer resp.Body.Close()
	//リクエストボディの取得
	byteArray, _ := ioutil.ReadAll(resp.Body)
	jsonBytes := ([]byte)(byteArray)
	data := new(ResContentsInFolder)
	if err := json.Unmarshal(jsonBytes, data); err != nil {
		return ResContentsInFolder{}, fmt.Errorf("[JSON Unmarshal error] error msg : %v", err)
	}
	return *data, nil
}
