package annex

import (
	"fmt"

	"github.com/gogs/git-module"
)

func GitAnnexUnlock(repoPath, filePath string) error {
	cmd := git.NewCommand("annex", "unlock", filePath)
	_, err := cmd.RunInDir(repoPath)
	if err != nil {
		return fmt.Errorf("[%v]. exec cmd : [%v]. exec dir : [%s]", err, cmd.String(), repoPath)
	}
	return nil
}
