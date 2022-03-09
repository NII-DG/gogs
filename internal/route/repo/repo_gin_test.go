package repo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ivis-yoshida/gogs/internal/context"
	"github.com/ivis-yoshida/gogs/internal/db"
	mock_context "github.com/ivis-yoshida/gogs/internal/mocks/context"
	mock_db "github.com/ivis-yoshida/gogs/internal/mocks/db"
	mock_repo "github.com/ivis-yoshida/gogs/internal/mocks/repo"
)

func Test_generateMaDmp(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	// mockCtrl := gomock.NewController(t)
	// defer mockCtrl.Finish()

	// mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	// mockCtxRepo := mock_context.NewMockAbstructCtxRepository(mockCtrl)

	tests := []struct {
		name                string
		PrepareMockContexts func() context.AbstructContext
		PrepareMockRepoUtil func() AbstructRepoUtil
	}{
		// add test content
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generateMaDmp(tt.PrepareMockContexts(), tt.PrepareMockRepoUtil())
		})
	}
}

func Test_fetchContentsOnGithub(t *testing.T) {
	wantByte := []byte(`{"name":"maDMP_for_test.ipynb","path":"maDMP_for_test.ipynb","sha":"859552c7e0503b939e70987e097dd2e9d236a99a","size":764,"url":"https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/maDMP_for_test.ipynb?ref=unittest","html_url":"https://github.com/ivis-kuwata/maDMP-template/blob/unittest/maDMP_for_test.ipynb","git_url":"https://api.github.com/repos/ivis-kuwata/maDMP-template/git/blobs/859552c7e0503b939e70987e097dd2e9d236a99a","download_url":"https://raw.githubusercontent.com/ivis-kuwata/maDMP-template/unittest/maDMP_for_test.ipynb","type":"file","content":"ewogImNlbGxzIjogWwogIHsKICAgImNlbGxfdHlwZSI6ICJtYXJrZG93biIs\nCiAgICJtZXRhZGF0YSI6IHt9LAogICAic291cmNlIjogWwogICAgIiMg5Y2Y\n5L2T44OG44K544OI55SobWFETVDjg4bjg7Pjg5fjg6zjg7zjg4hcbiIsCiAg\nICAiXG4iLAogICAgIuOBk+OCjOOBr+WNmOS9k+ODhuOCueODiOOBrueCuuOB\nrm1hRE1Q44OG44Oz44OX44Os44O844OI44Gn44GZ44CC44OG44K544OI57WQ\n5p6c44Gr5b2x6Z+/44KS5Y+K44G844GZ44Gf44KB44CB6Kix5Y+v44Gq44GP\n57eo6ZuG44O75YmK6Zmk44GX44Gq44GE44Gn44GP44Gg44GV44GE44CCIgog\nICBdCiAgfQogXSwKICJtZXRhZGF0YSI6IHsKICAia2VybmVsc3BlYyI6IHsK\nICAgImRpc3BsYXlfbmFtZSI6ICJQeXRob24gMyAoaXB5a2VybmVsKSIsCiAg\nICJsYW5ndWFnZSI6ICJweXRob24iLAogICAibmFtZSI6ICJweXRob24zIgog\nIH0sCiAgImxhbmd1YWdlX2luZm8iOiB7CiAgICJjb2RlbWlycm9yX21vZGUi\nOiB7CiAgICAibmFtZSI6ICJpcHl0aG9uIiwKICAgICJ2ZXJzaW9uIjogMwog\nICB9LAogICAiZmlsZV9leHRlbnNpb24iOiAiLnB5IiwKICAgIm1pbWV0eXBl\nIjogInRleHQveC1weXRob24iLAogICAibmFtZSI6ICJweXRob24iLAogICAi\nbmJjb252ZXJ0X2V4cG9ydGVyIjogInB5dGhvbiIsCiAgICJweWdtZW50c19s\nZXhlciI6ICJpcHl0aG9uMyIsCiAgICJ2ZXJzaW9uIjogIjMuOC4xMiIKICB9\nCiB9LAogIm5iZm9ybWF0IjogNCwKICJuYmZvcm1hdF9taW5vciI6IDIKfQo=\n","encoding":"base64","_links":{"self":"https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/maDMP_for_test.ipynb?ref=unittest","git":"https://api.github.com/repos/ivis-kuwata/maDMP-template/git/blobs/859552c7e0503b939e70987e097dd2e9d236a99a","html":"https://github.com/ivis-kuwata/maDMP-template/blob/unittest/maDMP_for_test.ipynb"}}`)

	type args struct {
		blobPath string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "succeed fetch blob",
			args: args{
				blobPath: "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/maDMP_for_test.ipynb?ref=unittest",
			},
			want:    wantByte,
			wantErr: false,
		},
		{
			name: "failed fetch blob",
			args: args{
				blobPath: "https://api.github.com/repos/no-exists/maDMP-template/contents/maDMP_for_test.ipynb",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f repoUtil
			got, err := f.fetchContentsOnGithub(tt.args.blobPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchContentsOnGithub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				// if !bytes.Equal(got, wantByte) {
				t.Errorf("fetchContentsOnGithub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeBlobContent(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDecodeStringer := mock_repo.NewMockDecodeStringer(mockCtrl)

	rightBlobInfo := []byte(`{"content":"SGVsbG8sIHdvcmxkLg=="}`)
	rightDecordedBlob := "Hello, world."

	wrongJsonInfo := []byte(`{"content":"SGVsbG8sIHdvcmxkLg=="`)

	type args struct {
		blobInfo []byte
	}
	tests := []struct {
		name                      string
		args                      args
		PrepareMockDecodeStringer func() DecodeStringer
		want                      string
		wantErr                   bool
	}{
		{
			name: "SucceedDecording",
			args: args{
				blobInfo: rightBlobInfo,
			},
			PrepareMockDecodeStringer: func() DecodeStringer {
				mockDecodeStringer.EXPECT().DecodeString("SGVsbG8sIHdvcmxkLg==").Return([]byte("Hello, world."), nil)
				return mockDecodeStringer
			},
			want:    rightDecordedBlob,
			wantErr: false,
		},
		{
			name: "FailUnmarshal",
			args: args{
				blobInfo: wrongJsonInfo,
			},
			PrepareMockDecodeStringer: func() DecodeStringer {
				return mockDecodeStringer
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "FailDecodeString",
			args: args{
				blobInfo: rightBlobInfo,
			},
			PrepareMockDecodeStringer: func() DecodeStringer {
				mockDecodeStringer.EXPECT().DecodeString("SGVsbG8sIHdvcmxkLg==").Return(nil, fmt.Errorf("これは想定されたエラーです"))
				return mockDecodeStringer
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f repoUtil
			got, err := f.decodeBlobContent(tt.args.blobInfo, tt.PrepareMockDecodeStringer())
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeBlobContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("decodeBlobContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_FailedGenereteMaDmp(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockCtxRepo := mock_context.NewMockAbstructCtxRepository(mockCtrl)
	mockFlash := mock_context.NewMockAbstructFlash(mockCtrl)

	type args struct {
		c   context.AbstructContext
		msg string
	}
	tests := []struct {
		name                string
		PrepareMockContexts func() context.AbstructContext
		msg                 string
	}{
		{
			name: "GetFlash, Redirectの呼び出し確認",
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().GetFlash().Return(mockFlash) //.Error("errortest")
				mockFlash.EXPECT().Error("errortest")
				mockCtx.EXPECT().GetRepo().Return(mockCtxRepo)
				mockCtxRepo.EXPECT().GetRepoLink().Return("dummyrepo")
				mockCtx.EXPECT().Redirect("dummyrepo")
				return mockCtx
			},
			msg: "errortest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f repoUtil
			f.FailedGenereteMaDmp(tt.PrepareMockContexts(), tt.msg)
		})
	}
}

func Test_repoUtil_fetchContentsOnGithub(t *testing.T) {

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux := http.NewServeMux()
		mux.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	})
	// 別goroutine上でリッスンが開始される
	ts := httptest.NewServer(h)
	defer ts.Close()

	responseBody := `{"name":"test.txt","path":"test.txt","sha":"d16645b06f063607594a4c43f5e2f8aaa1e8b5ad","size":40,"url":"https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/test.txt?ref=unittest","html_url":"https://github.com/ivis-kuwata/maDMP-template/blob/unittest/test.txt","git_url":"https://api.github.com/repos/ivis-kuwata/maDMP-template/git/blobs/d16645b06f063607594a4c43f5e2f8aaa1e8b5ad","download_url":"https://raw.githubusercontent.com/ivis-kuwata/maDMP-template/unittest/test.txt","type":"file","content":"IyBUaGlzIGlzIHVzZWQgdGVzdCBHb2dzLgojIERvbid0IGVkaXQuCg==\n","encoding":"base64","_links":{"self":"https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/test.txt?ref=unittest","git":"https://api.github.com/repos/ivis-kuwata/maDMP-template/git/blobs/d16645b06f063607594a4c43f5e2f8aaa1e8b5ad","html":"https://github.com/ivis-kuwata/maDMP-template/blob/unittest/test.txt"}}`

	type args struct {
		blobPath string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "期待通りの返り値が得られることを確認する",
			args: args{
				blobPath: "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/test.txt?ref=unittest",
			},
			want:    []byte(responseBody),
			wantErr: false,
		},
		{
			name: "client.Doで失敗することを確認する",
			args: args{
				blobPath: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "StatusNotFoundが得られた場合エラーを返すことを確認する",
			args: args{
				blobPath: "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/WRONG_CONTENT.txt?ref=unittest",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "verify",
			args: args{
				blobPath: "/test",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f repoUtil
			got, err := f.fetchContentsOnGithub(tt.args.blobPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("repoUtil.fetchContentsOnGithub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("repoUtil.fetchContentsOnGithub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readDmpJson(t *testing.T) {
	type args struct {
		c context.AbstructContext
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readDmpJson(tt.args.c)
		})
	}
}

func Test_fetchDockerfile(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepoUtil := mock_repo.NewMockAbstructRepoUtil(mockCtrl)
	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockCtxRepo := mock_context.NewMockAbstructCtxRepository(mockCtrl)
	mockDbRepo := mock_db.NewMockAbstructDbRepository(mockCtrl)
	mockDbUser := mock_db.NewMockAbstructDbUser(mockCtrl)

	// strings for test
	rightBlobInfo := []byte(`{"content":"SGVsbG8sIHdvcmxkLg=="}`)
	rightDecordedBlob := "Hello, world."
	// wrongJsonInfo := []byte(`{"content":"SGVsbG8sIHdvcmxkLg=="`)

	tests := []struct {
		name                string
		PrepareMockRepoUtil func() AbstructRepoUtil
		PrepareMockContexts func() context.AbstructContext
	}{
		{
			name: "successFetchDockerfile",
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/Dockerfile").Return([]byte(rightBlobInfo), nil)

				mockRepoUtil.EXPECT().DecodeBlobContent(rightBlobInfo).Return(rightDecordedBlob, nil)
				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtxRepo.EXPECT().GetDbRepo().Return(mockDbRepo)

				// argument of "UpdateRepoFile"
				mockCtx.EXPECT().GetUser().Return(mockDbUser)
				mockCtxRepo.EXPECT().GetLastCommitIdStr().Return("0000")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("test")

				mockDbRepo.EXPECT().UpdateRepoFile(mockDbUser, db.UpdateRepoFileOptions{
					LastCommitID: "0000",
					OldBranch:    "test",
					NewBranch:    "test",
					OldTreeName:  "",
					NewTreeName:  "Dockerfile",
					Message:      "[GIN] fetch Dockerfile",
					Content:      "Hello, world.",
					IsNewFile:    true,
				})
				return mockCtx
			},
		},
		{
			name: "fail_at_FetchContentsOnGithub",
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/Dockerfile").Return(nil, fmt.Errorf("これは想定されたエラーです"))

				mockRepoUtil.EXPECT().FailedGenereteMaDmp(mockCtx, "Sorry, faild gerate maDMP: fetching template failed(Dockerfile)")
				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				return mockCtx
			},
		},
		{
			name: "fail_at_DecodeBlobContent",
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/Dockerfile").Return([]byte(rightBlobInfo), nil)

				mockRepoUtil.EXPECT().DecodeBlobContent(rightBlobInfo).Return("", fmt.Errorf("これは想定されたエラーです"))

				mockRepoUtil.EXPECT().FailedGenereteMaDmp(mockCtx, "Sorry, faild gerate maDMP: fetching template failed(Dockerfile)")
				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				return mockCtx
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetchDockerfile(tt.PrepareMockContexts(), tt.PrepareMockRepoUtil())
		})
	}
}
