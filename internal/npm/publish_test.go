package npm

import (
	"strings"
	"testing"
)

func TestNpmError(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantSub string
	}{
		{
			name:    "EOTP triggers 2FA message",
			input:   "npm ERR! code EOTP\nnpm ERR! This operation requires a one-time password",
			wantSub: "2FA",
		},
		{
			name:    "ENEEDAUTH triggers auth message",
			input:   "npm ERR! code ENEEDAUTH\nnpm ERR! need auth",
			wantSub: "not authenticated",
		},
		{
			name:    "E401 triggers auth message",
			input:   "npm ERR! code E401\nnpm ERR! Unauthorized",
			wantSub: "not authenticated",
		},
		{
			name:    "E403 triggers permission message",
			input:   "npm ERR! code E403\nnpm ERR! Forbidden",
			wantSub: "permission denied",
		},
		{
			name:    "E409 triggers version exists message",
			input:   "npm ERR! code E409\nnpm ERR! conflict",
			wantSub: "already exists",
		},
		{
			name:    "EPUBLISHCONFLICT triggers version exists message",
			input:   "npm ERR! code EPUBLISHCONFLICT\nnpm ERR! conflict",
			wantSub: "already exists",
		},
		{
			name:    "ENOTFOUND triggers network message",
			input:   "npm ERR! code ENOTFOUND\nnpm ERR! network error",
			wantSub: "network error",
		},
		{
			name:    "ETIMEDOUT triggers network message",
			input:   "npm ERR! code ETIMEDOUT\nnpm ERR! timed out",
			wantSub: "network error",
		},
		{
			name:    "ECONNREFUSED triggers network message",
			input:   "npm ERR! code ECONNREFUSED\nnpm ERR! connection refused",
			wantSub: "network error",
		},
		{
			name:    "EUSAGE with provenance triggers provenance message",
			input:   "npm ERR! code EUSAGE\nnpm ERR! provenance is not supported outside CI",
			wantSub: "provenance",
		},
		{
			name:    "EUSAGE without provenance falls through to generic error",
			input:   "npm ERR! code EUSAGE\nnpm ERR! unrelated usage error",
			wantSub: "publish failed",
		},
		{
			name:    "unknown error falls back to raw output",
			input:   "something went wrong completely unexpectedly",
			wantSub: "publish failed",
		},
		{
			name:    "empty input falls back to raw output",
			input:   "",
			wantSub: "publish failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := npmError([]byte(tt.input))
			if !strings.Contains(got, tt.wantSub) {
				t.Errorf("npmError() = %q, want substring %q", got, tt.wantSub)
			}
		})
	}
}
