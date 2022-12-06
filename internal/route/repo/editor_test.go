// Copyright 2016 The Gogs Authors. All rights reserved.

// Use of this source code is governed by a MIT-style

// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/NII-DG/gogs/internal/context"
	mock_context "github.com/NII-DG/gogs/internal/mocks/context"
	mock_db "github.com/NII-DG/gogs/internal/mocks/db"
	mock_repo "github.com/NII-DG/gogs/internal/mocks/repo"
	"github.com/gogs/git-module"
	"github.com/golang/mock/gomock"
)

func Test_dmpUtil_CreateDmp(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDmpUtil := mock_repo.NewMockAbstructDmpUtil(mockCtrl)
	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockCtxRepo := mock_context.NewMockAbstructCtxRepository(mockCtrl)
	mockDbRepo := mock_db.NewMockAbstructDbRepository(mockCtrl)

	org := "jst"

	// c.CallData()された値を確認するための変数
	dummyData := make(map[string]interface{})

	tests := []struct {
		name                string
		PrepateMockDmpUtil  func() AbstructDmpUtil
		PrepareMockContexts func() context.AbstructContext
		wantErr             bool
	}{
		// TODO: Add test cases.
		{
			name: "正常系：エラーが発生しない",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx).Return(nil)
				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, org).Return(nil)
				mockDmpUtil.EXPECT().GetCombinedDmp(org).Return("Success", nil)
				return mockDmpUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return(org)
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)

				mockCtxRepo.EXPECT().GetTreePath().AnyTimes().Return("")
				mockCtxRepo.EXPECT().GetRepoLink().AnyTimes().Return("localhost:3080/dummyUser/dummyRepo")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("master")
				mockCtxRepo.EXPECT().GetDbRepo().AnyTimes().Return(mockDbRepo)
				mockCtxRepo.EXPECT().GetCommitId().AnyTimes().Return(&git.SHA1{})

				mockDbRepo.EXPECT().FullName().Return("dummyRepo")

				mockCtx.EXPECT().Success("repo/editor/edit")

				return mockCtx
			},
			wantErr: false,
		},
		{
			name: "異常系:BidingDmpSchemaList()でエラーが発生する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx).Return(fmt.Errorf("Error"))
				return mockDmpUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return(org)
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)

				mockCtxRepo.EXPECT().GetTreePath().AnyTimes().Return("")
				mockCtxRepo.EXPECT().GetRepoLink().AnyTimes().Return("localhost:3080/dummyUser/dummyRepo")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("master")
				mockCtxRepo.EXPECT().GetDbRepo().AnyTimes().Return(mockDbRepo)
				mockCtxRepo.EXPECT().GetCommitId().AnyTimes().Return(&git.SHA1{})

				return mockCtx
			},
			wantErr: true,
		},
		{
			name: "異常系:FetchDmpSchema()でエラーが発生する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx).Return(nil)
				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, org).Return(fmt.Errorf("Error"))
				return mockDmpUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return(org)
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)

				mockCtxRepo.EXPECT().GetTreePath().AnyTimes().Return("")
				mockCtxRepo.EXPECT().GetRepoLink().AnyTimes().Return("localhost:3080/dummyUser/dummyRepo")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("master")
				mockCtxRepo.EXPECT().GetDbRepo().AnyTimes().Return(mockDbRepo)
				mockCtxRepo.EXPECT().GetCommitId().AnyTimes().Return(&git.SHA1{})

				return mockCtx
			},
			wantErr: true,
		},
		{
			name: "異常系:GetCombinedDmp()でエラーが発生する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx).Return(nil)
				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, org).Return(nil)
				mockDmpUtil.EXPECT().GetCombinedDmp(org).Return("", fmt.Errorf("Error"))
				return mockDmpUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return(org)
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)

				mockCtxRepo.EXPECT().GetTreePath().AnyTimes().Return("")
				mockCtxRepo.EXPECT().GetRepoLink().AnyTimes().Return("localhost:3080/dummyUser/dummyRepo")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("master")
				mockCtxRepo.EXPECT().GetDbRepo().AnyTimes().Return(mockDbRepo)
				mockCtxRepo.EXPECT().GetCommitId().AnyTimes().Return(&git.SHA1{})

				return mockCtx
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createDmp(tt.PrepareMockContexts(), tt.PrepateMockDmpUtil())
			if !tt.wantErr {
				// 正常系の際はdummyDataの中身を確認
				if reflect.DeepEqual(fmt.Sprintf("%v", dummyData["FileContent"]), `{"name":"dummySchema"}`) {
					t.Errorf("DMP Schema is wrong,  got:  %v,  want: `[{\"name\":\"dummySchema\"}]`", dummyData["FileContent"])
				}
			}
		})
	}
}
