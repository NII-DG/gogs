package bcapi

import (
	//"encoding/json"
	"fmt"
	//"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"
)

type Req_json struct {
	UserCode        string `json:"user_code"`
	ContentHistorys []struct {
		ContentLocation string    `json:"content_location"`
		ContentHash     string    `json:"content_hash"`
		AddDateTime     time.Time `json:"add_date_time"`
	} `json:"content_historys"`
}

type Res_json struct {
	ContentHistorys []struct {
		UserCode    string    `json:"user_code"`
		ContentHash string    `json:"content_Hash"`
		AddDateTime time.Time `json:"add_date_time"`
	} `json:"content_historys"`
}

func creatContenthistorys(user string, contenLocation string) {
	url := "http://localhost8080/provenances"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	//ヘッダーセット
	req.Header.Set("Authorization", "token")
	req.Header.Set("Content-Type", "application/json")
	//パラメータ追加
	params := req.URL.Query()
	params.Add("user_code", user)
	params.Add("content_location", contenLocation)
	req.URL.RawQuery = params.Encode()

	client := new(http.Client)
	//通信実行
	resp, err := client.Do(req)

	dumpResp, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("%s", dumpResp)
	fmt.Printf("%v", err)

}
