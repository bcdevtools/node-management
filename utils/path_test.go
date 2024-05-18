package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTryExtractUserHomeDirFromPath(t *testing.T) {
	tests := []struct {
		path    string
		want    string
		wantErr bool
	}{
		{
			path:    "/",
			wantErr: true,
		},
		{
			path:    "/root",
			want:    "/root",
			wantErr: false,
		},
		{
			path:    "/root/",
			want:    "/root",
			wantErr: false,
		},
		{
			path:    "/root/a",
			want:    "/root",
			wantErr: false,
		},
		{
			path:    "/home",
			wantErr: true,
		},
		{
			path:    "/home/",
			wantErr: true,
		},
		{
			path:    "/home/abc",
			want:    "/home/abc",
			wantErr: false,
		},
		{
			path:    "/home/abc/",
			want:    "/home/abc",
			wantErr: false,
		},
		{
			path:    "/home/abc/def",
			want:    "/home/abc",
			wantErr: false,
		},
		{
			path:    "/Users",
			wantErr: true,
		},
		{
			path:    "/Users/",
			wantErr: true,
		},
		{
			path:    "/Users/abc",
			want:    "/Users/abc",
			wantErr: false,
		},
		{
			path:    "/Users/abc/",
			want:    "/Users/abc",
			wantErr: false,
		},
		{
			path:    "/Users/abc/def",
			want:    "/Users/abc",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := TryExtractUserHomeDirFromPath(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
