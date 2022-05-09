package encyrptfile

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func Encrypted(filepath string) (string, error) {
	//原本ファイルの取得
	plainText, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("Cannot find file !!!, TARGET_FILE_PATH : %v", filepath)
	}
	//共通キーの取得
	key := []byte("passw0rdpassw0rdpassw0rdpassw0rd")

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

}

func encryptedDataToIpfs(encryptedData []byte) (string, error) {
	//create command
	echoCmd := exec.Command("echo", string(encryptedData))
	addCmd := exec.Command("ipfs", "add")

	//make a pipe
	reader, writer := io.Pipe()
	var buf bytes.Buffer

	//set the output of "cat" command to pipe writer
	echoCmd.Stdout = writer
	//set the input of the "wc" command pipe reader

	addCmd.Stdin = reader

	//cache the output of "wc" to memory
	addCmd.Stdout = &buf

	//start to execute "cat" command
	echoCmd.Start()

	//start to execute "wc" command
	addCmd.Start()

	//waiting for "cat" command complete and close the writer
	echoCmd.Wait()
	writer.Close()

	//waiting for the "wc" command complete and close the reader
	addCmd.Wait()
	reader.Close()
	//copy the buf to the standard output
	io.Copy(os.Stdout, &buf)
}
