package bcapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var API_URL_CREATE_CONTENT_HISTORY_TOKEN = "createContentHistoryToken"

type ReqCreateContentHistory struct {
	UserCode        string `json:"user_code"`
	ContentHistorys []struct {
		ContentLocation string    `json:"content_location"`
		ContentAddress  string    `json:"content_address"`
		AddDateTime     time.Time `json:"add_date_time"`
	} `json:"content_history_list"`
}

//[POST] /createContentHistory
func CreateContentHistory(user_code string, contentMap map[string]string) error {
	//登録日時の取得
	now := time.Now()
	//リクエストボディ定義
	reqStr := ReqCreateContentHistory{}
	reqStr.UserCode = user_code
	for k, v := range contentMap {
		reqStr.ContentHistorys = append(reqStr.ContentHistorys, struct {
			ContentLocation string    "json:\"content_location\""
			ContentAddress  string    "json:\"content_address\""
			AddDateTime     time.Time "json:\"add_date_time\""
		}{k, v, now})
	}
	reqBody, err := json.Marshal(reqStr)
	if err != nil {
		return fmt.Errorf("[Fail convert Json] %v", err)
	}
	//リクエスト生成
	req, _ := createNewRequest(http.MethodPost, API_URL_CREATE_CONTENT_HISTORY_TOKEN, reqBody)
	//ヘッダー設定
	req.Header.Set("Authorization", "token")
	req.Header.Set("Content-Type", "application/json")

	//通信実行
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("[HTTP conection error] url : %v, error msg : %v", API_URL_CREATE_CONTENT_HISTORY_TOKEN, err)
	}
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("[Error in BC-API] status code : %v, response : %s", resp.StatusCode, byteArray)
	}
	return nil
}
