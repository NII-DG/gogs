package utils

import (
	"fmt"
	"os/exec"
	"time"

	log "unknwon.dev/clog/v2"
)

func ClearDirectoryCache(path string) error {

	log.Info("[DEBUG LOG BY RCOS] Start Cache Clear. Time : [%s], Path : [%s]", time.Now(), path)
	cmd := exec.Command("sync", path)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clear directory cache: %s", err)
	}
	log.Info("[DEBUG LOG BY RCOS] END Cache Clear. Time : [%s], Path : [%s]", time.Now(), path)
	return nil
}
