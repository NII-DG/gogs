package dataset_test

import (
	"fmt"
	"testing"

	"github.com/NII-DG/gogs/internal/db"
	mock_ipfs "github.com/NII-DG/gogs/internal/mocks/ipfs"
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

func TestGetDatasetAddress_異常系_意図しないエラーが発生した場合(t *testing.T) {
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

	//isDatasetFolderOnIPFS()のFilesIs() err("Internal Error")を返すMock
	rtnStrAry := []string{""}
	rtnErr := fmt.Errorf("Internal Error")
	mockIFIpfsOperation.EXPECT().FilesIs(datasetPath).Return(rtnStrAry, rtnErr)

	//実行
	d := dataset.DatasetCreater{}
	d.Operater = mockIFIpfsOperation
	_, err := d.GetDatasetAddress(datasetPath, datasetData)
	if err == nil {
		t.Fail()
	}
	expErr := fmt.Errorf("<%v>", rtnErr)
	assert.Equal(t, expErr, err)
}

func TestGetDatasetAddress_異常系_IPFSにフォルダが既に存在した場合(t *testing.T) {
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

	//isDatasetFolderOnIPFS()のFilesIs() errを返さないMock
	rtnStrAry := []string{""}
	mockIFIpfsOperation.EXPECT().FilesIs(datasetPath).Return(rtnStrAry, nil)

	//実行
	d := dataset.DatasetCreater{}
	d.Operater = mockIFIpfsOperation
	_, err := d.GetDatasetAddress(datasetPath, datasetData)

	//検証
	if err == nil {
		t.Fail()
	}
	expErr := fmt.Errorf("There Is A Possibility That Another User Upload Dataset To IPFS")
	assert.Equal(t, expErr, err)
}

func TestGetDatasetAddress_異常系_IPFSへのフォルダ構築の失敗(t *testing.T) {
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

	//createDatasetStructure() で errを返すようにする。FilesCopy() mock
	rtnErrCp := fmt.Errorf("Fail Copy")
	mockIFIpfsOperation.EXPECT().FilesCopy(inputList[0].Address, inputList[0].File).Return(rtnErrCp)

	//GetDatasetAddress()におけるFilesRemove() mock  errを返さない
	mockIFIpfsOperation.EXPECT().FilesRemove(datasetPath).Return(nil)

	d := dataset.DatasetCreater{}
	d.Operater = mockIFIpfsOperation

	//実行
	_, err := d.GetDatasetAddress(datasetPath, datasetData)

	//検証
	if err == nil {
		t.Fail()
	}
	expErr := fmt.Errorf("[Failure Create Foleder on IPFS] <%v>, Than Remove Creating Foleder", rtnErrCp)
	assert.Equal(t, expErr, err)
}

func TestGetDatasetAddress_異常系_IPFSへのフォルダ構築と削除を失敗(t *testing.T) {
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

	//createDatasetStructure() で errを返すようにする。FilesCopy() mock
	rtnErrCp := fmt.Errorf("Fail Copy")
	mockIFIpfsOperation.EXPECT().FilesCopy(inputList[0].Address, inputList[0].File).Return(rtnErrCp)

	//GetDatasetAddress()におけるFilesRemove() mock  errを返さない
	rtnErrRm := fmt.Errorf("Fail Remove")
	mockIFIpfsOperation.EXPECT().FilesRemove(datasetPath).Return(rtnErrRm)

	d := dataset.DatasetCreater{}
	d.Operater = mockIFIpfsOperation

	//実行
	_, err := d.GetDatasetAddress(datasetPath, datasetData)

	//検証
	if err == nil {
		t.Fail()
	}
	expErr := fmt.Errorf("[Failure Remove Creating Foleder on IPFS] <%v>,<%v>", rtnErrCp, rtnErrRm)
	assert.Equal(t, expErr, err)
}

func TestGetDatasetAddress_異常系_コンテンツアドレスの取得時エラー(t *testing.T) {
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
	mockIFIpfsOperation.EXPECT().FilesCopy(inputList[0].Address, inputList[0].File).Return(nil)
	mockIFIpfsOperation.EXPECT().FilesCopy(srcList[0].Address, srcList[0].File).Return(nil)
	mockIFIpfsOperation.EXPECT().FilesCopy(srcList[1].Address, srcList[1].File).Return(nil)
	mockIFIpfsOperation.EXPECT().FilesCopy(outputList[0].Address, outputList[0].File).Return(nil)

	//getUploadDatasetInfo()でinputフォルダのコンテンツを返すFilesStatus() Mock
	inputPath := datasetPath + "/" + db.INPUT_FOLDER_NM

	rtnErrStat := fmt.Errorf("Fail Stat")
	mockIFIpfsOperation.EXPECT().FilesStatus(inputPath).Return("", rtnErrStat)

	//getUploadDatasetInfo()でsrcフォルダのコンテンツを返すFilesStatus() Mock
	srcPath := datasetPath + "/" + db.SRC_FOLDER_NM
	rtnSrcAddress := "QmZzqfBnVLWEf8Mcvm1xfM5AEqaNHft5JgstXHX149TW4u"
	mockIFIpfsOperation.EXPECT().FilesStatus(srcPath).Return(rtnSrcAddress, nil)
	//getUploadDatasetInfo()でoutputフォルダのコンテンツを返すFilesStatus() Mock
	outputPath := datasetPath + "/" + db.OUTPUT_FOLDER_NM
	rtnOutputAddress := "QmUr3XYHNR1mRguGCG9z9jjAzG7P7XjBg38ubHBAWUUKdM"
	mockIFIpfsOperation.EXPECT().FilesStatus(outputPath).Return(rtnOutputAddress, nil)

	//実行
	d := dataset.DatasetCreater{}
	d.Operater = mockIFIpfsOperation

	_, err := d.GetDatasetAddress(datasetPath, datasetData)

	//検証
	if err == nil {
		t.Fail()
	}
	expErr := fmt.Errorf("[Failure Get Upload Dataset Address From IPFS] <INPUT : %v>, <SRC : %v>, <OUTPUT : %v>", rtnErrStat, nil, nil)
	assert.Equal(t, expErr, err)
}

func TestGetDatasetAddress_異常系_フォルダー構成削除が失敗(t *testing.T) {
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
	rtnErrRm := fmt.Errorf("Fail Remove")
	mockIFIpfsOperation.EXPECT().FilesRemove(datasetPath).Return(rtnErrRm)

	//実行
	d := dataset.DatasetCreater{}
	d.Operater = mockIFIpfsOperation

	_, err := d.GetDatasetAddress(datasetPath, datasetData)

	//検証
	if err == nil {
		t.Fail()
	}
	expErr := fmt.Errorf("[Failure Remove Created Foleder on IPFS] %v", rtnErrRm)
	assert.Equal(t, expErr, err)
}
