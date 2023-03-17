package annex

import (
	"encoding/json"
	"fmt"

	"github.com/gogs/git-module"
)

func GitAnnexMetadata(repoPath, key string) ([]byte, error) {
	cmd := git.NewCommand("annex", "metadata", "--key", key, "--json")
	raw_msg, err := cmd.RunInDir(repoPath)
	if err != nil {
		return nil, fmt.Errorf("[%v]. exec cmd : [%v]. exec dir : [%s]", err, cmd.String(), repoPath)
	}
	return raw_msg, nil
}

func SetAnnexMetadata(repoPath, key string, size int64, hash, mimetype string) error {

	content_size := fmt.Sprintf("content_size=%d", size)
	sha256 := fmt.Sprintf("sha256=%s", hash)
	mime_type := fmt.Sprintf("mime_type=%s", mimetype)
	cmd := git.NewCommand("annex", "metadata", "--key", key, "-s", content_size, "-s", sha256, "-s", mime_type)
	_, err := cmd.RunInDir(repoPath)
	if err != nil {
		return fmt.Errorf("[%v]. exec cmd : [%v]. exec dir : [%s]", err, cmd.String(), repoPath)
	}
	return nil
}

type Field struct {
	ContentSize     string
	Sha256          string
	SdDatePublished string
	EncodingFormat  string
}

func GetFieldsFromMetadata(repoPath, key string) (Field, error) {
	raw_msg, err := GitAnnexMetadata(repoPath, key)
	if err != nil {
		return Field{}, err
	}

	var jsonObj interface{}
	_ = json.Unmarshal(raw_msg, &jsonObj)
	raw_field := jsonObj.(map[string]interface{})["fields"].(map[string]interface{})

	field := Field{}
	if val, ok := raw_field["content_size"]; ok {
		field.ContentSize = val.([]interface{})[0].(string)
	} else {
		field.ContentSize = ""
	}

	if val, ok := raw_field["sha256"]; ok {
		field.Sha256 = val.([]interface{})[0].(string)
	} else {
		field.Sha256 = ""
	}

	if val, ok := raw_field["sd_date_published"]; ok {
		field.SdDatePublished = val.([]interface{})[0].(string)
	} else {
		field.SdDatePublished = ""
	}

	if val, ok := raw_field["mime_type"]; ok {
		field.EncodingFormat = val.([]interface{})[0].(string)
	} else {
		field.EncodingFormat = ""
	}
	return field, nil
}
