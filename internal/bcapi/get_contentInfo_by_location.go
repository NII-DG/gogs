package bcapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var API_URL_GET_CONTENT_INFO_BY_LOCATION string = "getContentInfoByLocation"

type ResContentInfo struct {
	UserCode    string    `json:"user_code"`
	ContentHash string    `json:"content_address"`
	AddDateTime time.Time `json:"add_date_time"`
}

//[GET] /getContentInfoByLocation
func GetContentInfoByLocation(userCode, fileLocation string) (ResContentInfo, error) {
	//リクエスト生成
	req, _ := createNewRequest(http.MethodGet, API_URL_GET_CONTENT_INFO_BY_LOCATION, nil)
	//ヘッダー設定
	req.Header.Set("Authorization", "token")
	req.Header.Set("Content-Type", "application/json")
	//パラメータ設定
	params := req.URL.Query()
	params.Add("user_code", userCode)
	params.Add("content_location", fileLocation)
	req.URL.RawQuery = params.Encode()

	//通信実行
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return ResContentInfo{}, fmt.Errorf("[HTTP conection error] url : %v, error msg : %v", API_URL_GET_CONTENT_INFO_BY_LOCATION, err)
	}
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)
	jsonBytes := ([]byte)(byteArray)
	data := new(ResContentInfo)
	if err := json.Unmarshal(jsonBytes, data); err != nil {
		return ResContentInfo{}, fmt.Errorf("[JSON Unmarshal error] error msg : %v", err)
	}
	return *data, nil
}
