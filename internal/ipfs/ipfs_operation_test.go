package ipfs_test

import (
	"testing"

	"github.com/NII-DG/gogs/internal/ipfs"
	mock_main "github.com/NII-DG/gogs/internal/ipfs/mock"
	"github.com/golang/mock/gomock"
)

func TestFilesCopy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIFIpfsCommand := mock_main.NewMockIFIpfsCommand(ctrl)

	contentAddress := "Qmflojfsidfljl"
	fullRepoFilePath := "/owner/repo/branch/dataset1/innput/test1.txt"
	contentParam := "/ipfs/" + contentAddress

	mockIFIpfsCommand.EXPECT().AddArgs("files", "cp", contentParam, "-p", fullRepoFilePath)

	msg := ""
	rtn := []byte(msg)
	mockIFIpfsCommand.EXPECT().Run().Return(rtn, nil)

	i := ipfs.IpfsOperation{}
	i.Command = mockIFIpfsCommand
	err := i.FilesCopy(contentAddress, fullRepoFilePath)

	if err != nil {
		t.Fail()
	}
}
