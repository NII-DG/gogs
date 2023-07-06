package healthcheck

import (
	"fmt"
	"path/filepath"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/utils"
	log "unknwon.dev/clog/v2"
)

/*
 Check connecting File System
*/
func CheckFileSystem() error {
	if len(conf.DG.HealthFilePath) <= 0 || len(conf.DG.HealthFileName) <= 0 {
		log.Error("health check to file system is skip")
		return nil
	}
	filepath := filepath.Join(conf.DG.HealthFilePath, conf.DG.HealthFileName)
	if utils.ExistData(filepath) {
		return nil
	} else {
		return fmt.Errorf("not exist %s for health check", filepath)
	}
}
