package context_test

// ↑cyclicError回避のための対応
// refs: https://qiita.com/raharu0425/items/518a140c323b430c4506

import (
	"testing"

	"github.com/ivis-yoshida/gogs/internal/context"
)

// NOTE: 他ライブラリのリファクタリングが必要と見られるため、
//       実装を中断している
func Test_hasFileInRepo(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	// mockCtrl := gomock.NewController(t)
	// defer mockCtrl.Finish()

	// mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	// mockCtxRepo := mock_context.NewMockAbstructCtxRepository(mockCtrl)
	// mockDbRepo := mock_db.NewMockAbstructDbRepository(mockCtrl)

	// mockGitRepo := mock_context.NewMockAbstructGitRepository(mockCtrl)
	// gitCommit := &git.Commit{}

	tests := []struct {
		name                string
		PrepareMockContexts func() context.AbstructContext
		filePath            string
		want                bool
	}{
		// {
		// 	name: "hasFileInRepoがtrueを返す（正常終了する）ことを確認する",
		// 	PrepareMockContexts: func() context.AbstructContext {
		// 		mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)

		// 		mockCtxRepo.EXPECT().GetGitRepo().Return(mockGitRepo)
		// 		mockCtxRepo.EXPECT().GetDbRepo().Return(mockDbRepo)
		// 		mockDbRepo.EXPECT().GetDefaultBranch().Return("master")
		// 		mockGitRepo.EXPECT().BranchCommit("master")

		// 		// gitCommit.EXPECT().Blob("./dmp.json")

		// 		return mockCtx
		// 	},
		// 	filePath: "./dmp.json",
		// 	want:     true,
		// },
		// {
		// 	name: "hasFileInRepoがfalseを返すことを確認する",
		// 	PrepareMockContexts: func() context.AbstructContext {
		// 		mockCtx.EXPECT().GetRepo().Return(mockCtxRepo)

		// 		mockCtxRepo.EXPECT().GetGitRepo().Return(mockGitRepo)
		// 		mockCtxRepo.EXPECT().GetDbRepo().Return(mockDbRepo)
		// 		mockDbRepo.EXPECT().GetDefaultBranch().Return("master")
		// 		mockGitRepo.EXPECT().BranchCommit("master").Return(nil, fmt.Errorf("これは想定されたエラーです"))

		// 		return mockCtx
		// 	},
		// 	filePath: "./dmp.json",
		// 	want:     false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := context.HasFileInRepo(tt.PrepareMockContexts(), tt.filePath); got != tt.want {
				t.Errorf("hasFileInRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}
