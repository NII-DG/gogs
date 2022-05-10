package encyrptfile

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"unsafe"

	"github.com/NII-DG/gogs/internal/ipfs"
	log "unknwon.dev/clog/v2"
)

//AES CTRモードの暗号化メソッド
//
//@param filepath　暗号化するファイルのパス
//
//@param password 暗号キー
func Encrypted(filepath, password string) (string, error) {
	//原本ファイルの取得
	plainText, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("[Cannot find file !!!, TARGET_FILE_PATH : %v]", filepath)
	}
	//共通キーの取得
	key := []byte(password)

	// Create new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("[Failure Creating new AES cipher block in Encrypting This FilePath : %v]", filepath)
	}

	// Create IV (cipherText : 暗号化データ)
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("[Failure Creating new IV in Encrypting This FilePath : %v]", filepath)
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

//AES CTRモードの復号化メソッド
//
//@param ipfsCid　暗号データを紐づくIPFSコンテンツアドレス
//
//@param password 復号キー
//
//@param filepath 復号したファイルの格納パス
func Decrypted(ipfsCid, password, filepath string) error {
	//暗号データの取得　from IPFS
	operater := ipfs.IpfsOperation{
		Commander: ipfs.NewCommand(),
	}
	cipherText, err := operater.Cat(ipfsCid)
	if err != nil {
		return err
	}

	//共通キーの取得
	key := []byte(password)

	// Create new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("[Failure Creating new AES cipher block in Dencrypting This IPFS CID : %v]", ipfsCid)
	}

	decryptedText := make([]byte, len(cipherText[aes.BlockSize:]))
	decryptStream := cipher.NewCTR(block, cipherText[:aes.BlockSize])
	decryptStream.XORKeyStream(decryptedText, cipherText[aes.BlockSize:])

	log.Trace("[Decrypted()]open file: %v", filepath)
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("[Cannot Open file : %v, Error Msg : %v]", filepath, err)
	}
	defer file.Close()
	_, err = file.Write(decryptedText)
	if err != nil {
		return fmt.Errorf("[Cannot Write file : %v, Error Msg : %v]", filepath, err)
	}
	return nil
}
