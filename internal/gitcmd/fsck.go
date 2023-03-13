package gitcmd

import (
	"fmt"

	"github.com/NII-DG/gogs/internal/utils"
	"github.com/gogs/git-module"
)

func GitFsck(repoPath string) (string, error) {
	cmd := git.NewCommand("fsck", "--full")
	raw_msg, err := cmd.RunInDir(repoPath)
	if err != nil {
		return "", fmt.Errorf("[%v]. exec cmd : [%v]. exec dir : [%s]", err, cmd.String(), repoPath)
	}
	return utils.BytesToString(raw_msg), nil
}
