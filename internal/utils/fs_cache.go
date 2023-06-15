package utils

import (
	"fmt"
	"os/exec"

	log "unknwon.dev/clog/v2"
)

func clearDirectoryCache(path string) error {

	log.Info("[DEBUG LOG BY RCOS] Start Cache Clear")
	cmd := exec.Command("sync", path)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clear directory cache: %s", err)
	}

	return nil
}
