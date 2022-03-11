package bcapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type UploadDatasetInfo struct {
	InputAddress  string
	SrcAddress    string
	OutputAddress string
}

var API_URL_CREATE_DATASET_TOKEN string = "createDatasetToken"

type ReqCreateDatasetToken struct {
	UserCode    string `json:"user_code"`
	DatasetList []struct {
		DatasetLocation string    `json:"content_location"`
		InputAddress    string    `json:"content_address"`
		SrcAddress      string    `json:"src_address"`
		OutputAddress   string    `json:"output_address"`
		AddDateTime     time.Time `json:"add_date_time"`
	} `json:"dataset_list"`
}

type ResNotCreateDatasetToken struct {
	DatasetList []struct {
		DatasetLocation string `json:"content_location"`
		InputAddress    string `json:"content_address"`
		SrcAddress      string `json:"src_address"`
		OutputAddress   string `json:"output_address"`
	} `json:"dataset_list"`
}

//[POST] /createDatasetToken
func CreateDatasetToken(user_code string, contentMap map[string]UploadDatasetInfo) (ResNotCreateDatasetToken, error) {
	//登録日時の取得
	now := time.Now()
	//リクエストボディ定義
	reqStr := ReqCreateDatasetToken{}
	reqStr.UserCode = user_code
	for k, v := range contentMap {
		reqStr.DatasetList = append(reqStr.DatasetList, struct {
			DatasetLocation string    "json:\"content_location\""
			InputAddress    string    "json:\"content_address\""
			SrcAddress      string    "json:\"src_address\""
			OutputAddress   string    "json:\"output_address\""
			AddDateTime     time.Time "json:\"add_date_time\""
		}{k, v.InputAddress, v.SrcAddress, v.OutputAddress, now})
	}
	reqBody, err := json.Marshal(reqStr)
	if err != nil {
		return ResNotCreateDatasetToken{}, fmt.Errorf("[Fail convert Json] %v", err)
	}
	//リクエスト生成
	req, _ := createNewRequest(http.MethodPost, API_URL_CREATE_DATASET_TOKEN, reqBody)
	//ヘッダー設定
	req.Header.Set("Authorization", "token")
	req.Header.Set("Content-Type", "application/json")

	//通信実行
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return ResNotCreateDatasetToken{}, fmt.Errorf("[HTTP conection error] url : %v, error msg : %v", API_URL_CREATE_DATASET_TOKEN, err)
	}
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return ResNotCreateDatasetToken{}, fmt.Errorf("[Error in BC-API] status code : %v, response : %s", resp.StatusCode, byteArray)
	}
	//リクエストボディ取得
	jsonBytes := ([]byte)(byteArray)
	data := new(ResNotCreateDatasetToken)
	if err := json.Unmarshal(jsonBytes, data); err != nil {
		return ResNotCreateDatasetToken{}, fmt.Errorf("[JSON Unmarshal error] error msg : %v", err)
	}
	return *data, nil
}
