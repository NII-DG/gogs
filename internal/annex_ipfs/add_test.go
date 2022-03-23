package annex_ipfs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAnnexAddInfo_正常系(t *testing.T) {
	var strJson = `{"command":"add1","note":"note1","success":false,"key":"key1","file":"file1"}
	{"command":"add2","note":"note2","success":false,"key":"key2","file":"file2"}`
	byteJson := []byte(strJson)
	res, err := GetAnnexAddInfo(&byteJson)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	} else {
		assert.Equal(t, "add1", res[0].Command)
		assert.Equal(t, "note1", res[0].Note)
		assert.Equal(t, false, res[0].Success)
		assert.Equal(t, "key1", res[0].Key)
		assert.Equal(t, "file1", res[0].File)

		assert.Equal(t, "add2", res[1].Command)
		assert.Equal(t, "note2", res[1].Note)
		assert.Equal(t, false, res[1].Success)
		assert.Equal(t, "key2", res[1].Key)
		assert.Equal(t, "file2", res[1].File)
	}

}
