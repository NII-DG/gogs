package dataset_test

import (
	"fmt"
	"testing"

	"github.com/NII-DG/gogs/internal/db"
	mock_ipfs "github.com/NII-DG/gogs/internal/ipfs/mock"
	"github.com/NII-DG/gogs/internal/route/dataset"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetDatasetAddress_正常系(t *testing.T) {
	//GetDatasetAddress()の引数の定義
	datasetPath := "/user01/demo01/master/dataset1"
	inputList := []db.ContentInfo{
		{
			File:    "/user01/demo01/master/dataset1/input/input_data.txt",
			Address: "QmWs4uF1aseAtAXMNwhmE2sX9qENbWjSz98m48QWZk9gjw",
		},
	}
	srcList := []db.ContentInfo{
		{
			File:    "/user01/demo01/master/dataset1/src/main/src_data.txt",
			Address: "Qmf13tWcQt471ckdf3RX7zbhBbZq1KurX9zbNvWeFZFFKM",
		},
		{
			File:    "/user01/demo01/master/dataset1/src/main/src_data2.txt",
			Address: "QmdpvhPobPzsCwB793PhP2RLUFdqFqBjhKyoJ8uQ1Qeuq4",
		},
	}
	outputList := []db.ContentInfo{
		{
			File:    "/user01/demo01/master/dataset1/output/output_data.txt",
			Address: "QmTfoE7CGczRcm32h5hDAxpqiCW3uyxvsx3NgWPUHEFETX",
		},
	}
	datasetData := db.DatasetInfo{
		InputList:  inputList,
		SrcList:    srcList,
		OutputList: outputList,
	}

	//Mock IFIpfsOperation定義
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIFIpfsOperation := mock_ipfs.NewMockIFIpfsOperation(ctrl)

	//isDatasetFolderOnIPFS()でfalse, nilを返すようにするFilesIs() Mock
	rtnStrAry := []string{""}
	rtnErr := fmt.Errorf("file does not exist")
	mockIFIpfsOperation.EXPECT().FilesIs(datasetPath).Return(rtnStrAry, rtnErr)

	//createDatasetStructure() で nilを返すようにする。FilesCopy() mock
	// mockIFIpfsOperation.EXPECT().FilesCopy(gomock.Any(), gomock.Any()).Return(nil)
	mockIFIpfsOperation.EXPECT().FilesCopy(inputList[0].Address, inputList[0].File).Return(nil)
	mockIFIpfsOperation.EXPECT().FilesCopy(srcList[0].Address, srcList[0].File).Return(nil)
	mockIFIpfsOperation.EXPECT().FilesCopy(srcList[1].Address, srcList[1].File).Return(nil)
	mockIFIpfsOperation.EXPECT().FilesCopy(outputList[0].Address, outputList[0].File).Return(nil)

	//getUploadDatasetInfo()でinputフォルダのコンテンツを返すFilesStatus() Mock
	inputPath := datasetPath + "/" + db.INPUT_FOLDER_NM
	rtnInputAddress := "QmSrFnBYKjKW3PfqNLvoBmLUA9FXZCPHhWo2WeVLZn6URK"
	mockIFIpfsOperation.EXPECT().FilesStatus(inputPath).Return(rtnInputAddress, nil)

	//getUploadDatasetInfo()でsrcフォルダのコンテンツを返すFilesStatus() Mock
	srcPath := datasetPath + "/" + db.SRC_FOLDER_NM
	rtnSrcAddress := "QmZzqfBnVLWEf8Mcvm1xfM5AEqaNHft5JgstXHX149TW4u"
	mockIFIpfsOperation.EXPECT().FilesStatus(srcPath).Return(rtnSrcAddress, nil)
	//getUploadDatasetInfo()でoutputフォルダのコンテンツを返すFilesStatus() Mock
	outputPath := datasetPath + "/" + db.OUTPUT_FOLDER_NM
	rtnOutputAddress := "QmUr3XYHNR1mRguGCG9z9jjAzG7P7XjBg38ubHBAWUUKdM"
	mockIFIpfsOperation.EXPECT().FilesStatus(outputPath).Return(rtnOutputAddress, nil)

	//GetDatasetAddress()におけるFilesRemove() mock  errを返さない
	mockIFIpfsOperation.EXPECT().FilesRemove(datasetPath).Return(nil)

	d := dataset.DatasetCreater{}
	d.Operater = mockIFIpfsOperation

	rtnData, err := d.GetDatasetAddress(datasetPath, datasetData)
	if err != nil {
		fmt.Println("[Test Error TestGetDatasetAddress_正常系] ", err)
		t.Fail()
	}
	assert.Equal(t, rtnInputAddress, rtnData.InputAddress)
	assert.Equal(t, rtnSrcAddress, rtnData.SrcAddress)
	assert.Equal(t, rtnOutputAddress, rtnData.OutputAddress)
}
