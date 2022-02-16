// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ivis-yoshida/gogs/internal/context"
	mock_context "github.com/ivis-yoshida/gogs/internal/testutil/mock/context"
	mock_db "github.com/ivis-yoshida/gogs/internal/testutil/mock/db"
	mock_repo "github.com/ivis-yoshida/gogs/internal/testutil/mock/repo"
)

func TestCreateDmp(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type args struct {
		c context.AbstructContext
	}
	tests := []struct {
		name    string
		context context.AbstructContext
		// args args
	}{
		{
			name:    "verification",
			context: successPattern(mockCtrl),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CreateDmp(tt.context)
		})
	}
}

func Test_fetchDmpSchema(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockRepoUtil := mock_repo.NewMockAbstructRepoUtil(mockCtrl)

	tests := []struct {
		name           string
		blobPath       string
		PrepareMockFn  func() AbstructRepoUtil
		PrepareMockCtx func() context.AbstructContext
		wantErr        bool
	}{
		{
			name:     "succeed_fetchDmpSchema",
			blobPath: "https://right.pattern.com/dmp/orgs/meti",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://right.pattern.com/dmp/orgs/meti").Return([]byte(`{"content":"SGVsbG8sIHdvcmxkLg=="}`), nil)
				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`{"content":"SGVsbG8sIHdvcmxkLg=="}`)).Return("Hello, world.", nil)
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(2).Return(make(map[string]interface{}))
				return mockCtx
			},
			wantErr: false,
		},
		{
			name:     "failed_FetchContentsOnGithub",
			blobPath: "https://wrong.pattern.com/dmp/orgs/meti",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://wrong.pattern.com/dmp/orgs/meti").Return(nil, fmt.Errorf("DesiredError"))
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(0)
				return mockCtx
			},
			wantErr: true,
		},
		{
			name:     "failed_DecodeBlobContent",
			blobPath: "https://right.pattern.com/dmp/orgs/meti",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://right.pattern.com/dmp/orgs/meti").Return([]byte(`{"data":"SGVsbG8sIHdvcmxkLg=="}`), nil)
				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`{"data":"SGVsbG8sIHdvcmxkLg=="}`)).Return("", fmt.Errorf("DesiredError"))
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(0)
				return mockCtx
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d dmpUtil
			if err := d.fetchDmpSchema(tt.PrepareMockCtx(), tt.PrepareMockFn(), tt.blobPath); (err != nil) != tt.wantErr {
				t.Errorf("fetchDmpSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_bidingDmpSchemaList(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockRepoUtil := mock_repo.NewMockAbstructRepoUtil(mockCtrl)

	tests := []struct {
		name           string
		treePath       string
		PrepareMockFn  func() AbstructRepoUtil
		PrepareMockCtx func() context.AbstructContext
		wantErr        bool
	}{
		{
			name:     "succeed_bindingDmpSchemaList",
			treePath: "https://right.pattern.com/dmp/schema",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://right.pattern.com/dmp/schema").Return([]byte(`[{"name":"dummySchema"}]`), nil)
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(1).Return(make(map[string]interface{}))
				return mockCtx
			},
			wantErr: false,
		},
		{
			name:     "failed_FetchContentsOnGithub",
			treePath: "https://wrong.pattern.com/dmp/schema",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://wrong.pattern.com/dmp/schema").Return(nil, fmt.Errorf("DesiredError"))
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(0)
				return mockCtx
			},
			wantErr: true,
		},
		{
			name:     "failed_Unmarshal",
			treePath: "https://right.pattern.com/dmp/schema",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://right.pattern.com/dmp/schema").Return([]byte(`wrong JSON`), nil)
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(0)
				return mockCtx
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d dmpUtil
			if err := d.bidingDmpSchemaList(tt.PrepareMockCtx(), tt.PrepareMockFn(), tt.treePath); (err != nil) != tt.wantErr {
				t.Errorf("bidingDmpSchemaList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func successPattern(mockCtrl *gomock.Controller) context.AbstructContext {
	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockCtxRepo := mock_context.NewMockAbstructCtxRepository(mockCtrl)
	mockDbRepo := mock_db.NewMockAbstructDbRepository(mockCtrl)

	mockCtx.EXPECT().PageIs("Edit")
	mockCtx.EXPECT().RequireHighlightJS()
	mockCtx.EXPECT().RequireSimpleMDE()
	mockCtx.EXPECT().QueryEscape("schema").Return("meti")
	mockCtx.EXPECT().GetRepo().Return(mockCtxRepo)
	mockCtx.EXPECT().CallData().MaxTimes(2).Return(make(map[string]interface{}))
	mockCtx.EXPECT().Success("repo/editor/edit")

	mockCtxRepo.EXPECT().GetTreePath().Return("")
	// mockCtxRepo.EXPECT().GetRepoLink().Return("localhost:3080/dummyUser/dummyRepo")
	// mockCtxRepo.EXPECT().GetBranchName().Return("master")
	// mockCtxRepo.EXPECT().GetDbRepo().Return(mockDbRepo)
	// mockCtxRepo.EXPECT().GetCommitId().Return(&git.SHA1{})

	mockDbRepo.EXPECT().FullName().Return("dummyRepo")

	return mockCtx
}
