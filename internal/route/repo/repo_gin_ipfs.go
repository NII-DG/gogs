package repo

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unsafe"

	"github.com/NII-DG/gogs/internal/bcapi"
	"github.com/NII-DG/gogs/internal/context"
	encyrptfile "github.com/NII-DG/gogs/internal/ipfs/encyrpt_file"
	"github.com/NII-DG/gogs/internal/tool"
	"github.com/gogs/git-module"
	log "unknwon.dev/clog/v2"
)

//指定されたコンテンツをIPFSから取得する。この時、ブロックチェーン（BC）における存在証明の検証が行われる。
//さらに、BCからの取得内容に従い、暗号化コンテンツは復号されて、返される
//
//@  contentLocation string　コンテンツロケーション
func resolveAnnexedContentFromIPFS(c *context.Context, buf []byte, contentLocation string) ([]byte, error) {
	if !tool.IsAnnexedFile(buf) {
		// not an annex pointer file; return as is
		return buf, nil
	}
	repoPath := c.Repo.Repository.RepoPath()

	//ベアレポジトリをIPFSへ連携有効化
	if _, err := git.NewCommand("annex", "enableremote", "ipfs").RunInDir(repoPath); err != nil {
		log.Error("[Failure enable remote(ipfs)] err : %v, repoPath : %v", err, repoPath)
	} else {
		log.Info("[Success enable remote(ipfs)] repoPath : %v", repoPath)
	}

	//BCAPI通信（コンテンツパスからBC上のコンテンツ情報を取得）
	bcContentInfo, err := bcapi.GetContentInfoByLocation(c.User.Name, contentLocation)
	if err != nil {
		log.Error("[BC-API HTTP error] %v", err)
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}

	//指定のコンテンツの暗号化の有無の判定する。
	if bcContentInfo.IsPrivate {
		//暗号ファイルの処理
		return resolveEncyrptedContent(c, buf, bcContentInfo)
	} else {
		//公開データの処理
		return resolvePublicContent(c, buf, bcContentInfo)
	}
}

//コンテンツの存在証明検証し、IPFSから暗号データを取得、復号して、復号データを返す。
//
//@param
func resolvePublicContent(c *context.Context, buf []byte, bcContentInfo bcapi.ResContentInfo) ([]byte, error) {
	repoPath := c.Repo.Repository.RepoPath()

	keyparts := strings.Split(strings.TrimSpace(string(buf)), "/")
	key := keyparts[len(keyparts)-1]

	addressByAnnex, err := GetIpfsHashValueByAnnexKey(key, repoPath)
	if err != nil {
		log.Error("[Cannot Get IPFS Hash] key : %v, err : %v", key, err)
	} else {
		log.Info("[Get IPFS Hash From AnnexKey] key : %v To hash : %v", key, addressByAnnex)
	}

	//BC-IPFSハッシュ値とAnnex-IPFSハッシュ値を比較
	if addressByAnnex != bcContentInfo.IpfsCid {
		log.Error("[Not math AnnexContentAddress to BcContentAddress] AnnexContentAddress : %v, BcContentAddress : %v", addressByAnnex, bcContentInfo.IpfsCid)
		c.Data["IsAnnexedFile"] = true
		return buf, fmt.Errorf("[Not math AnnexContentAddress to BcContentAddress]")
	}

	//ipfsからオブジェクトを取得
	if _, err := git.NewCommand("annex", "copy", "--from", "ipfs", "--key", key).RunInDir(repoPath); err != nil {
		log.Error("[Failure copy dataObject from ipfs] err : %v, repoPath : %v", err, repoPath)
	} else {
		log.Info("[Success copy dataObject from ipfs] key : %v, repoPath : %v", key, repoPath)
	}

	contentPath, err := git.NewCommand("annex", "contentlocation", key).RunInDir(repoPath)
	if err != nil {
		log.Error("Failed to find content location for key %q, err : %v", key, err)
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	// always trim space from output for git command
	contentPath = bytes.TrimSpace(contentPath)
	///home/ivis/gogs-repositories/user1/demo2.git +
	afp, err := os.Open(filepath.Join(repoPath, string(contentPath)))
	if err != nil {
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	info, err := afp.Stat()
	if err != nil {
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	annexDataReader := bufio.NewReader(afp)
	annexBuf := make([]byte, 1024)
	n, _ := annexDataReader.Read(annexBuf)
	annexBuf = annexBuf[:n]
	c.Data["FileSize"] = info.Size()
	log.Trace("Annexed file size: %d B", info.Size())
	//メモrepopath + /annex を削除する。
	//gogs-repositories/user1/demo4.git/annex
	return annexBuf, nil
}

func resolveEncyrptedContent(c *context.Context, buf []byte, bcContentInfo bcapi.ResContentInfo) ([]byte, error) {
	repoPath := c.Repo.Repository.RepoPath()

	keyparts := strings.Split(strings.TrimSpace(string(buf)), "/")
	key := keyparts[len(keyparts)-1]

	//暗号化ファイルの処理
	if _, err := git.NewCommand("annex", "copy", "--from", "ipfs", "--key", key).RunInDir(repoPath); err != nil {
		log.Error("[Failure copy dataObject from ipfs] err : %v, repoPath : %v", err, repoPath)
	} else {
		log.Info("[Success copy dataObject from ipfs] key : %v, repoPath : %v", key, repoPath)
	}
	contentPath, err := git.NewCommand("annex", "contentlocation", key).RunInDir(repoPath)
	if err != nil {
		log.Error("Failed to find content location for key %q, err : %v", key, err)
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	log.Trace("contentPath: %v", string(contentPath))
	contentPath = bytes.TrimSpace(contentPath)
	filepath := filepath.Join(repoPath, string(contentPath))
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Error("[ioutil.ReadFile(filepath.Join(repoPath, string(contentPath)))] err : %v, repoPath : %v", err, repoPath)
	}
	fullContentHash := string(b)

	//ファイルコンテンツハッシュを比較
	if fullContentHash != bcContentInfo.FullContentHash {
		log.Error("[Not match AnnexFullContentHash to BcFullContentHash] FullContentHash : %v, BcFullContentHash : %v", fullContentHash, bcContentInfo.FullContentHash)
		c.Data["IsAnnexedFile"] = true
		return buf, fmt.Errorf("[Not math AnnexContentAddress to BcContentAddress]")
	}

	//ファイルコンテンツハッシュが一致した場合
	//IPFSから暗号データを取得して、復元する。
	log.Trace("[Match AnnexFullContentHash to BcFullContentHash] fullContentHash : %v, BCFullContentHash : %v", fullContentHash, bcContentInfo.FullContentHash)
	//復号ファイルの格納ディレクトリパスの定義
	dirDecryptedData := strings.Replace(filepath, "objects", "decrypt", 1)
	err = encyrptfile.DecryptedFromIPFS(bcContentInfo.IpfsCid, c.Repo.Repository.Password, dirDecryptedData)
	if err != nil {
		log.Error("[Cannot Decrypting data] File : %v, Error Msg : %v", dirDecryptedData, err)
		c.Data["IsAnnexedFile"] = true
		return buf, fmt.Errorf("[Cannot Decrypting data] File : %v, Error Msg : %v", dirDecryptedData, err)
	}
	//復号したデータを取得する。
	afp, err := os.Open(dirDecryptedData)
	if err != nil {
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	info, err := afp.Stat()
	if err != nil {
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	annexDataReader := bufio.NewReader(afp)
	annexBuf := make([]byte, 1024)
	n, _ := annexDataReader.Read(annexBuf)
	annexBuf = annexBuf[:n]
	c.Data["FileSize"] = info.Size()
	log.Trace("Annexed file size: %d B", info.Size())
	return annexBuf, nil
}

//Convert Annex Key to IPFS hash
func GetIpfsHashValueByAnnexKey(key string, repoPath string) (string, error) {
	var hash string
	var err error
	if msg, err := git.NewCommand("annex", "whereis", "--key", key).RunInDir(repoPath); err == nil {
		strMsg := *(*string)(unsafe.Pointer(&msg)) //[]byte to string
		reg := "\r\n|\n"
		arr1 := regexp.MustCompile(reg).Split(strMsg, -1) //改行分割
		for _, s := range arr1 {
			if strings.Contains(s, "ipfs:") {
				index := strings.LastIndex(s, ":")
				trimStr := &hash
				*trimStr = s[index+1:]
			}
		}
	}
	return hash, err
}
