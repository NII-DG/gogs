package db

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/G-Node/libgin/libgin"
	"github.com/G-Node/libgin/libgin/annex"
	"github.com/gogs/git-module"
	"github.com/ivis-yoshida/gogs/internal/conf"
	"github.com/unknwon/com"
	"golang.org/x/crypto/bcrypt"
	log "gopkg.in/clog.v1"
	logv2 "unknwon.dev/clog/v2"
)

// StartIndexing sends an indexing request to the configured indexing service
// for a repository.
func StartIndexing(repo Repository) {
	go func() {
		if conf.Search.IndexURL == "" {
			log.Trace("Indexing not enabled")
			return
		}
		log.Trace("Indexing repository %d", repo.ID)
		ireq := libgin.IndexRequest{
			RepoID:   repo.ID,
			RepoPath: repo.FullName(),
		}
		data, err := json.Marshal(ireq)
		if err != nil {
			log.Error(2, "Could not marshal index request: %v", err)
			return
		}
		key := []byte(conf.Search.Key)
		encdata, err := libgin.EncryptString(key, string(data))
		if err != nil {
			log.Error(2, "Could not encrypt index request: %v", err)
		}
		req, err := http.NewRequest(http.MethodPost, conf.Search.IndexURL, strings.NewReader(encdata))
		if err != nil {
			log.Error(2, "Error creating index request")
		}
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Error(2, "Error submitting index request for [%d: %s]: %v", repo.ID, repo.FullName(), err)
			return
		}
	}()
}

// RebuildIndex sends all repositories to the indexing service to be indexed.
func RebuildIndex() error {
	indexurl := conf.Search.IndexURL
	if indexurl == "" {
		return fmt.Errorf("Indexing service not configured")
	}

	// collect all repo ID -> Path mappings directly from the DB
	repos := make(RepositoryList, 0, 100)
	if err := x.Find(&repos); err != nil {
		return fmt.Errorf("get all repos: %v", err)
	}
	log.Trace("Found %d repositories to index", len(repos))
	for _, repo := range repos {
		StartIndexing(*repo)
	}
	log.Trace("Rebuilding search index")
	return nil
}

func annexUninit(path string) {
	// walker sets the permission for any file found to 0660, to allow deletion
	var mode os.FileMode
	walker := func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		mode = 0660
		if info.IsDir() {
			mode = 0770
		}

		if err := os.Chmod(path, mode); err != nil {
			log.Error(3, "failed to change permissions on '%s': %v", path, err)
		}
		return nil
	}

	log.Trace("Uninit annex at '%s'", path)
	if msg, err := annex.Uninit(path); err != nil {
		log.Error(3, "uninit failed: %v (%s)", err, msg)
		if werr := filepath.Walk(path, walker); werr != nil {
			log.Error(3, "file permission change failed: %v", werr)
		}
	}
}

/**
UPDATE : 2022/02/01
AUTHOR : dai.tsukioka
*/
func annexSetup(path string) {
	logv2.Info("Running annex add (with filesize filter) in '%s'", path)

	// Initialise annex in case it's a new repository

	if msg, err := annex.Init(path); err != nil {
		logv2.Error("[Annex init failed] err : %v, msg : %s", err, msg)
		return
	} else {
		logv2.Info("[Annex init Success] err : %v, msg : %s", err, msg)
	}

	// Upgrade to v8 in case the directory was here before and wasn't cleaned up properly
	if msg, err := annex.Upgrade(path); err != nil {
		log.Error(2, "Annex upgrade failed: %v (%s)", err, msg)
		return
	}

	// Enable addunlocked for annex v8
	if msg, err := annex.SetAddUnlocked(path); err != nil {
		logv2.Error("Failed to set 'addunlocked' annex option: %v (%s)", err, msg)
	}

	// Set MD5 as default backend
	if msg, err := annex.MD5(path); err != nil {
		logv2.Error("Failed to set default backend to 'MD5': %v (%s)", err, msg)
	}

	// Set size filter in config
	//large fileのサイズを定義する（不要）
	// if msg, err := annex.SetAnnexSizeFilter(path, conf.Repository.Upload.AnnexFileMinSize*annex.MEGABYTE); err != nil {
	// 	logv2.Error("Failed to set size filter for annex: %v (%s)", err, msg)
	// }

	//Setting initremote ipfs
	if msg, err := setRemoteIPFS(path); err != nil {
		log.Error(2, "Annex initremote ipfs failed: %v (%s)", err, msg)
		return
	}
}

/**
UPDATE : 2022/02/01
AUTHOR : dai.tsukioka
NOTE : Setting initremote ipfs
*/
func setRemoteIPFS(path string) ([]byte, error) {
	cmd := git.NewCommand("annex", "initremote")
	cmd.AddArgs("ipfs", "type=external", "externaltype=ipfs", "encryption=none")
	msg, err := cmd.RunInDir(path)
	logv2.Info("[Initremoto IPFS] path : %v, msg : %s, error : %v", path, msg, err)
	return msg, err
}

func annexAdd(repoPath string, all bool, files ...string) error {
	cmd := git.NewCommand("annex", "add")
	if all {
		cmd.AddArgs(".")
	}
	msg, err := cmd.AddArgs(files...).RunInDir(repoPath)
	logv2.Info("[AnnexAdd] msg : %s, err : %v", msg, err)
	return err
}

/**
UPDATE : 2022/02/01
AUTHOR : dai.tsukioka
NOTE : methods : [sync and copy] locations are invert
*/
func annexUpload(repoPath, remote string) error {
	//IPFSの所在確認（デバック用）
	logv2.Info("[git annex whereis1-1] path : %v", repoPath)
	if msg, err := git.NewCommand("annex", "whereis").RunInDir(repoPath); err != nil {
		logv2.Error("[git annex whereis Error] err : %v", err)
	} else {
		logv2.Info("[git annes whereis Info] msg : %s", msg)
	}

	//ipfsへ実データをコピーする。
	logv2.Info("[Uploading annexed data to %v] path : %v", remote, repoPath)
	cmd := git.NewCommand("annex", "copy", "--to", remote)
	if msg, err := cmd.RunInDir(repoPath); err != nil {
		return fmt.Errorf("[Failure git annex copy to %v] err : %v ,fromPath : %v", remote, err, repoPath)
	} else {
		logv2.Info("[Success copy to ipfs] msg : %s, fromPath : %v", msg, repoPath)
	}

	//IPFSの所在確認（デバック用）
	logv2.Info("[git annex whereis1-2] path : %v", repoPath)
	if msg, err := git.NewCommand("annex", "whereis").RunInDir(repoPath); err != nil {
		logv2.Error("[git annex whereis Error] err : %v", err)
	} else {
		logv2.Info("[git annes whereis Info] msg : %s", msg)
	}

	//リモートと同期（メタデータを更新）
	log.Info("Synchronising annex info : %v", repoPath)
	if msg, err := git.NewCommand("annex", "sync").RunInDir(repoPath); err != nil {
		logv2.Error("[Failure git-annex sync] err : %v, msg : %s", err, msg)
	} else {
		logv2.Info("[Success git-annex sync] msg : %s", msg)
	}
	return nil
}

// isAddressAllowed returns true if the email address is allowed to sign up
// based on the regular expressions found in the email filter file
// (custom/addressfilters).
// In case of errors (opening or reading file) or no matches, the function
// defaults to 'true'.
func isAddressAllowed(email string) bool {
	fpath := filepath.Join(conf.CustomDir(), "addressfilters")
	if !com.IsExist(fpath) {
		// file doesn't exist: default allow everything
		return true
	}

	f, err := os.Open(fpath)
	if err != nil {
		log.Error(2, "Failed to open file %q: %v", fpath, err)
		// file read error: default allow everything
		return true
	}
	defer f.Close()

	emailBytes := []byte(email)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Check provided email address against each line regex
		// Failure to match any line returns true (allowed)
		// Matching a line prefixed with + returns true (allowed)
		// Matching a line prefixed with - returns false (blocked)
		// Erroneous patterns are logged and ignored
		var allow bool
		line := scanner.Text()
		if line[0] == '-' {
			allow = false
		} else if line[0] == '+' {
			allow = true
		} else {
			log.Error(2, "Invalid line in addressfilters: %s", line)
			log.Error(2, "Prefix invalid (must be '-' or '+')")
			continue
		}
		pattern := strings.TrimSpace(line[1:])
		match, err := regexp.Match(pattern, emailBytes)
		if err != nil {
			log.Error(2, "Invalid line in addressfilters: %s", line)
			log.Error(2, "Invalid pattern: %v", err)
		}
		if match {
			log.Trace("New user email %q matched filter rule %q (Allow: %t)", email, line, allow)
			return allow
		}
	}

	// No match: Default to allow
	return true
}

func (u *User) OldGinVerifyPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Passwd), []byte(plain))
	return err == nil
}
