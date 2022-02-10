package annex_ipfs

//git annex whereis --jsonの構造体
type annexWhereResponse struct {
	Command   string   `json:"command"`
	Note      string   `json:"note"`
	Success   bool     `json:"success"`
	Untrusted []string `json:"untrusted"`
	Key       string   `json:"key"`
	Whereis   []struct {
		Here        bool     `json:"here"`
		Uuid        string   `json:"uuid"`
		Urls        []string `json:"urls"`
		Description []string `json:"description"`
	} `json:"whereis"`
	File string `json:"file"`
}
