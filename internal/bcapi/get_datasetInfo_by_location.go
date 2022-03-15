package bcapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var API_URL_GET_DATASET_INFO_BY_LOCATION string = "getDatasetToken"

type ResDatasetInfo struct {
	DatasetLocation string `json:"dataset_location"`
	InputAddress    string `json:"input_address"`
	SrcCodeAddress  string `json:"src_address"`
	OutputAddress   string `json:"output_address"`
}

//[GET] /getDatasetToken
func GetDatasetInfoByLocation(userCode, datasetLocation string) (ResDatasetInfo, error) {
	//リクエスト生成
	req, _ := createNewRequest(http.MethodGet, API_URL_GET_DATASET_INFO_BY_LOCATION, nil)
	//ヘッダー設定
	req.Header.Set("Authorization", "token")
	req.Header.Set("Content-Type", "application/json")
	//パラメータ設定
	params := req.URL.Query()
	params.Add("user_code", userCode)
	params.Add("dataset_location", datasetLocation)
	req.URL.RawQuery = params.Encode()

	//通信実行
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return ResDatasetInfo{}, fmt.Errorf("[HTTP conection error] url : %v, error msg : %v", API_URL_GET_DATASET_INFO_BY_LOCATION, err)
	}
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)
	jsonBytes := ([]byte)(byteArray)
	data := new(ResDatasetInfo)
	if err := json.Unmarshal(jsonBytes, data); err != nil {
		return ResDatasetInfo{}, fmt.Errorf("[JSON Unmarshal error] error msg : %v", err)
	}
	return *data, nil
}
