package model

import "testing"

func TestTypeToLTree(t *testing.T) {
	type args struct {
		partType string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "archive",
			args:    args{partType: "archive"},
			want:    "archive",
			wantErr: false,
		},
		{
			name:    "archive with junk",
			args:    args{partType: "archive/foo/bar/zap"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "archive with custom",
			args:    args{partType: "archive/custom/foo/bar/zap"},
			want:    "archive.custom.foo.bar.zap",
			wantErr: false,
		},
		{
			name:    "archive with empty custom",
			args:    args{partType: "archive/custom"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "container without sub-type",
			args:    args{partType: "container"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "container with junk",
			args:    args{partType: "container/foo/bar/zap"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "container image",
			args:    args{partType: "container/image"},
			want:    "container.image",
			wantErr: false,
		},
		{
			name:    "container image with junk",
			args:    args{partType: "container/image/foo/bar/zap"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "container source",
			args:    args{partType: "container/source"},
			want:    "container.source",
			wantErr: false,
		},
		{
			name:    "container source with junk",
			args:    args{partType: "container/source/foo/bar/zap"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "container with custom",
			args:    args{partType: "container/custom/foo/bar/zap"},
			want:    "container.custom.foo.bar.zap",
			wantErr: false,
		},
		{
			name:    "container with empty custom",
			args:    args{partType: "container/custom"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TypeToLTree(tt.args.partType)
			if (err != nil) != tt.wantErr {
				t.Errorf("TypeToLTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TypeToLTree() = %v, want %v", got, tt.want)
			}
		})
	}
}
