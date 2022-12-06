package repo

import (
	"testing"

	"github.com/NII-DG/gogs/internal/context"
)

func Test_generateMaDmp(t *testing.T) {
	// TODO: mockの準備

	tests := []struct {
		name                string
		PrepareMockContexts func() context.AbstructContext
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generateMaDmp(tt.PrepareMockContexts())
		})
	}
}
