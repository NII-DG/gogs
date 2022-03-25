package ipfs_test

import (
	"fmt"
	"testing"

	"github.com/NII-DG/gogs/internal/ipfs"
	mock_ipfs "github.com/NII-DG/gogs/internal/ipfs/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFilesCopy_正常系(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIFIpfsCommand := mock_ipfs.NewMockIFIpfsCommand(ctrl)

	contentAddress := "Qmflojfsidfljl"
	fullRepoFilePath := "/owner/repo/branch/dataset1/innput/test1.txt"
	contentParam := "/ipfs/" + contentAddress

	mockIFIpfsCommand.EXPECT().RemoveArgs()
	mockIFIpfsCommand.EXPECT().AddArgs("files", "cp", contentParam, "-p", fullRepoFilePath)

	msg := ""
	rtn := []byte(msg)
	mockIFIpfsCommand.EXPECT().Run().Return(rtn, nil)

	i := ipfs.IpfsOperation{}
	i.Commander = mockIFIpfsCommand
	err := i.FilesCopy(contentAddress, fullRepoFilePath)

	if err != nil {
		t.Fail()
	}
}

func TestFilesCopy_異常系(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIFIpfsCommand := mock_ipfs.NewMockIFIpfsCommand(ctrl)

	contentAddress := "Qmflojfsidfljl"
	fullRepoFilePath := "/owner/repo/branch/dataset1/innput/test1.txt"
	contentParam := "/ipfs/" + contentAddress

	mockIFIpfsCommand.EXPECT().RemoveArgs()
	mockIFIpfsCommand.EXPECT().AddArgs("files", "cp", contentParam, "-p", fullRepoFilePath)

	msg := ""
	rtn := []byte(msg)
	rtnErr := fmt.Errorf("return error")
	mockIFIpfsCommand.EXPECT().Run().Return(rtn, rtnErr)

	i := ipfs.IpfsOperation{}
	i.Commander = mockIFIpfsCommand
	err := i.FilesCopy(contentAddress, fullRepoFilePath)
	if err == nil {
		t.Fail()
	}
	expect := fmt.Errorf("[Failure ipfs files cp ...] Content Adress : %v, FullRepoFilePath : %v, err : %v", contentAddress, fullRepoFilePath, rtnErr)
	assert.Equal(t, expect, err)
}
