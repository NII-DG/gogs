package encyrptfile

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"unsafe"

	"github.com/NII-DG/gogs/internal/ipfs"
)

func Encrypted(filepath, password string) (string, error) {
	//原本ファイルの取得
	plainText, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("Cannot find file !!!, TARGET_FILE_PATH : %v", filepath)
	}
	//共通キーの取得
	key := []byte(password)

	// Create new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("Failure Creating new AES cipher block in Encrypting This FilePath [%v]", filepath)
	}

	// Create IV (cipherText : 暗号化データ)
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("Failure Creating new IV in Encrypting This FilePath [%v]", filepath)
	}

	// Encrypt
	encryptStream := cipher.NewCTR(block, iv)
	encryptStream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//cipherText(暗号化データ)をIPFSにアップロードする。
	address, err := ipfs.DirectlyAdd(*(*string)(unsafe.Pointer(&cipherText)))
	if err != nil {
		return "", err
	}
	return address, nil
}

func Decrypted() {

}
