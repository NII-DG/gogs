package json_struct

import "time"

type ResContentsInFolder struct {
	ContentsInFolder []struct {
		UserCode        string    `json:"user_code"`
		ContentLocation string    `json:"content_location"`
		FullContentHash string    `json:"full_content_hash"`
		IpfsCid         string    `json:"ipfs_cid"`
		AddDateTime     time.Time `json:"add_date_time"`
	} `json:"contents_in_folder"`
}
