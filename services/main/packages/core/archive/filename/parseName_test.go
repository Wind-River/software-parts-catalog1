package filename

import "testing"

func TestGetPkgNameVersion(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name        string
		args        args
		wantName    string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "libarchive",
			args:        args{filename: "libarchive-3.6.1.tar.xz"},
			wantName:    "libarchive",
			wantVersion: "3.6.1",
			wantErr:     false,
		},
		{
			name:        "foo",
			args:        args{filename: "foo.tar.bz2"},
			wantName:    "foo",
			wantVersion: "",
			wantErr:     false,
		},
		{
			name:        "example webtest",
			args:        args{filename: "WebTest-2.0.35.tar.gz"},
			wantName:    "webtest",
			wantVersion: "2.0.35",
			wantErr:     false,
		},
		{
			name:        "example urllib3",
			args:        args{filename: "urllib3-1.26.5.tar.gz"},
			wantName:    "urllib3",
			wantVersion: "1.26.5",
			wantErr:     false,
		},
		{
			name:        "empty string",
			args:        args{filename: ""},
			wantName:    "",
			wantVersion: "",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotVersion, err := GetPkgNameVersion(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPkgNameVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("GetPkgNameVersion() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("GetPkgNameVersion() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}
