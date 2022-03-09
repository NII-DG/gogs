// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"testing"

	"github.com/ivis-yoshida/gogs/internal/context"
	"github.com/ivis-yoshida/gogs/internal/db"
)

// NOTE: 他ライブラリのリファクタリングが必要と見られるため、
//       実装を中断している
func Test_exploreMetadata(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	// mockCtrl := gomock.NewController(t)
	// defer mockCtrl.Finish()

	// mockDbUtil := mock_db.NewMockAbstructDbUtil(mockCtrl)

	// mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	// // mockCtxRepo := mock_context.NewMockAbstructCtxRepository(mockCtrl)

	// mockDbRepo := mock_db.NewMockAbstructDbRepository(mockCtrl)
	// mockDbUser := mock_db.NewMockAbstructDbUser(mockCtrl)

	// // c.CallData()された値を確認するための変数
	// dummyData := make(map[string]interface{})

	tests := []struct {
		name                string
		PrepareMockDbUtil   func() db.AbstructDbUtil
		PrepareMockContexts func() context.AbstructContext
	}{
		// {
		// 	name: "exploreMetadataの正常終了を確認する（変数page==0）",
		// 	PrepareMockDbUtil: func() db.AbstructDbUtil {
		// 		mockRepos := []*mock_db.MockAbstructDbRepository{mockDbRepo}

		// 		mockDbUtil.EXPECT().SearchRepositoryByName(&db.SearchRepoOptions{
		// 			Keyword:  "",
		// 			UserID:   int64(1),
		// 			OrderBy:  "updated_unix DESC",
		// 			Page:     1,
		// 			PageSize: conf.UI.ExplorePagingNum,
		// 		}).Return(mockRepos, int64(1), nil)

		// 		return mockDbUtil
		// 	},
		// 	PrepareMockContexts: func() context.AbstructContext {
		// 		mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)
		// 		mockCtx.EXPECT().Tr("explore")
		// 		mockCtx.EXPECT().GetUser().Return(mockDbUser)

		// 		mockDbUser.EXPECT().GetType().Return(db.UserFA)

		// 		mockCtx.EXPECT().Query("selectKey").Return("schema")
		// 		mockCtx.EXPECT().Query("q").Return("meti")
		// 		mockCtx.EXPECT().QueryInt("page").Return(0)
		// 		mockCtx.EXPECT().UserID().Return(int64(1))

		// 		// FIXME: return-value is depend on actual repository
		// 		mockDbRepo.EXPECT().RepoPath().Return("ivis-kuwata/test1")

		// 		mockCtx.EXPECT().Success("explore/metadata")

		// 		return mockCtx
		// 	},
		// },

		// {
		// 	name: "db.SearchRepositoryByNameで失敗することを確認する",
		// },
		// {
		// 	name: "git.Open(repo.RepoPath())で失敗することを確認する",
		// },
		// {
		// 	name: "gitRepo.CatFileCommitで失敗することを確認する",
		// },
		// {
		// 	name: "commit.Blobで失敗することを確認する",
		// },
		// {
		// 	name: "entry.Bytesで失敗することを確認する",
		// },
		// {
		// 	name: "db.RepositoryList(repos).LoadAttributesで失敗することを確認する",
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exploreMetadata(tt.PrepareMockContexts(), tt.PrepareMockDbUtil())
		})
	}
}

func Test_isContained(t *testing.T) {
	type args struct {
		bufStr      string
		selectedKey string
		keyword     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "isContainedがtrueを返すことを確認する",
			args: args{
				bufStr:      "\"schema\": \"meti\"",
				selectedKey: "schema",
				keyword:     "meti",
			},
			want: true,
		},
		{
			name: "isContainedがfalseを返すことを確認する",
			args: args{
				bufStr:      "\"schema\": \"meti\"",
				selectedKey: "worngkey",
				keyword:     "meti",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isContained(tt.args.bufStr, tt.args.selectedKey, tt.args.keyword); got != tt.want {
				t.Errorf("isContained() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dmpBrowsing(t *testing.T) {
	type args struct {
		c context.AbstructContext
		d db.AbstructDbUtil
	}
	tests := []struct {
		name string
		args args
	}{
		// NOTE: 他ライブラリのリファクタリングが必要と見られるため、
		//       まだ実装していない
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dmpBrowsing(tt.args.c, tt.args.d)
		})
	}
}
