package ipfs_test

import (
	"fmt"
	"testing"

	"github.com/NII-DG/gogs/internal/ipfs"
	mock_ipfs "github.com/NII-DG/gogs/internal/mocks/ipfs"
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

func TestFilesStatus_正常系(t *testing.T) {
	//FilesStatus()の引数の定義
	folderPath := "/urs01/repo01/master/dataset1"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIFIpfsCommand := mock_ipfs.NewMockIFIpfsCommand(ctrl)
	mockIFIpfsCommand.EXPECT().RemoveArgs()
	mockIFIpfsCommand.EXPECT().AddArgs("files", "stat", folderPath)
	msg := `item1
	item2
	item3`
	rtn := []byte(msg)
	mockIFIpfsCommand.EXPECT().Run().Return(rtn, nil)

	i := ipfs.IpfsOperation{}
	i.Commander = mockIFIpfsCommand
	rtnStr, err := i.FilesStatus(folderPath)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, "item1", rtnStr)

}

func TestFilesStatus_異常系(t *testing.T) {
	//FilesStatus()の引数の定義
	folderPath := "/urs01/repo01/master/dataset1"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIFIpfsCommand := mock_ipfs.NewMockIFIpfsCommand(ctrl)
	mockIFIpfsCommand.EXPECT().RemoveArgs()
	mockIFIpfsCommand.EXPECT().AddArgs("files", "stat", folderPath)
	msg := ``
	rtn := []byte(msg)
	rtnErr := fmt.Errorf("return error")
	mockIFIpfsCommand.EXPECT().Run().Return(rtn, rtnErr)

	i := ipfs.IpfsOperation{}
	i.Commander = mockIFIpfsCommand
	_, err := i.FilesStatus(folderPath)
	if err == nil {
		t.Fail()
	}
	exErr := fmt.Errorf("[Failure ipfs files stat ...] FolderPath : %v", folderPath)
	assert.Equal(t, exErr, err)

}

func TestFilesRemove_正常系(t *testing.T) {
	//FilesStatus()の引数の定義
	folderPath := "/urs01/repo01/master/dataset1"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIFIpfsCommand := mock_ipfs.NewMockIFIpfsCommand(ctrl)
	mockIFIpfsCommand.EXPECT().RemoveArgs()
	mockIFIpfsCommand.EXPECT().AddArgs("files", "rm", "-r", folderPath)
	msg := `item1
	item2
	item3`
	rtn := []byte(msg)
	mockIFIpfsCommand.EXPECT().Run().Return(rtn, nil)

	i := ipfs.IpfsOperation{}
	i.Commander = mockIFIpfsCommand
	err := i.FilesRemove(folderPath)
	if err != nil {
		t.Fail()
	}
}

func TestFilesRemove_異常系(t *testing.T) {
	//FilesStatus()の引数の定義
	folderPath := "/urs01/repo01/master/dataset1"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIFIpfsCommand := mock_ipfs.NewMockIFIpfsCommand(ctrl)
	mockIFIpfsCommand.EXPECT().RemoveArgs()
	mockIFIpfsCommand.EXPECT().AddArgs("files", "rm", "-r", folderPath)
	msg := ``
	rtn := []byte(msg)
	rtnErr := fmt.Errorf("return error")
	mockIFIpfsCommand.EXPECT().Run().Return(rtn, rtnErr)

	i := ipfs.IpfsOperation{}
	i.Commander = mockIFIpfsCommand
	err := i.FilesRemove(folderPath)
	if err == nil {
		t.Fail()
	}
	exErr := fmt.Errorf("[Failure ipfs file rm ...] FolderPath : %v", folderPath)
	assert.Equal(t, exErr, err)

}

func TestFilesIs_正常系(t *testing.T) {
	//FilesStatus()の引数の定義
	folderPath := "/urs01/repo01/master/dataset1"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIFIpfsCommand := mock_ipfs.NewMockIFIpfsCommand(ctrl)
	mockIFIpfsCommand.EXPECT().RemoveArgs()
	mockIFIpfsCommand.EXPECT().AddArgs("files", "ls", folderPath)
	msg := "item1\r\nitem2\nitem3"
	rtn := []byte(msg)
	mockIFIpfsCommand.EXPECT().Run().Return(rtn, nil)

	i := ipfs.IpfsOperation{}
	i.Commander = mockIFIpfsCommand
	rtnStr, err := i.FilesIs(folderPath)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, "item1", rtnStr[0])
	assert.Equal(t, "item2", rtnStr[1])
	assert.Equal(t, "item3", rtnStr[2])

}

func TestFilesIs_異常系(t *testing.T) {
	//FilesStatus()の引数の定義
	folderPath := "/urs01/repo01/master/dataset1"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIFIpfsCommand := mock_ipfs.NewMockIFIpfsCommand(ctrl)
	mockIFIpfsCommand.EXPECT().RemoveArgs()
	mockIFIpfsCommand.EXPECT().AddArgs("files", "ls", folderPath)
	msg := ``
	rtn := []byte(msg)
	rtnErr := fmt.Errorf("return error")
	mockIFIpfsCommand.EXPECT().Run().Return(rtn, rtnErr)

	i := ipfs.IpfsOperation{}
	i.Commander = mockIFIpfsCommand
	_, err := i.FilesIs(folderPath)
	if err == nil {
		t.Fail()
	}
	exErr := fmt.Errorf("[Failure ipfs file is ...] <%v>, FolderPath : %v", rtnErr, folderPath)
	assert.Equal(t, exErr, err)

}
