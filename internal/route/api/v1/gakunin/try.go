package gakunin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NII-DG/gogs/internal/context"
)

type UserInfo struct {
	IdpUserName string `json:"user" binding:"Required"`
	ServerName  string `json:"server" binding:"Required"`
}

func DeleteConteiner(c *context.APIContext, ui UserInfo) {

	//path := "http://jupyter.cs.rcos.nii.ac.jp/hub/api/users/" + ui.IdpUserName + "/servers/" + ui.ServerName
	path := "http://163.220.176.50:10180/hub/api/users/" + ui.IdpUserName + "/servers/" + ui.ServerName
	//body := bytes.NewReader([]byte(`{"remove": true}`))

	fmt.Println("path:", path)
	//fmt.Println(c.Req.Header)

	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		c.Error(err, "error")
	}
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("create request")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		c.ErrorStatus(resp.StatusCode, err)
		return
	}
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)

	var orgsInfo interface{}
	_ = json.Unmarshal(contents, &orgsInfo)
	c.JSONSuccess(orgsInfo)

	fmt.Printf("%+v\n", resp)
}
