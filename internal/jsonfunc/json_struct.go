package jsonfunc

import "time"

type ResContentsInFolder struct {
	ContentsInFolder []struct {
		UserCode        string    `json:"user_code"`
		ContentLocation string    `json:"content_location"`
		FullContentHash string    `json:"full_content_hash"`
		IpfsCid         string    `json:"ipfs_cid"`
		IsPrivate       bool      `json:"is_private"`
		AddDateTime     time.Time `json:"add_date_time"`
	} `json:"contents_in_folder"`
}
