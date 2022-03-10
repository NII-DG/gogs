package jsonfunc

import "encoding/json"

func IsJSONString(s string) bool {
	var js json.RawMessage
	err := json.Unmarshal([]byte(s), &js)
	return err == nil
}
