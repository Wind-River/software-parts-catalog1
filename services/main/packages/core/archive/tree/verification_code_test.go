package tree

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func MustSha256(hexRepresentation string) [32]byte {
	slice, err := hex.DecodeString(hexRepresentation)
	if err != nil {
		panic(err)
	}

	return [32]byte(slice)
}

func MustMd5(hexRepresentation string) [16]byte {
	slice, err := hex.DecodeString(hexRepresentation)
	if err != nil {
		panic(err)
	}

	return [16]byte(slice)
}

func MustSha1(hexRepresentation string) [20]byte {
	slice, err := hex.DecodeString(hexRepresentation)
	if err != nil {
		panic(err)
	}

	return [20]byte(slice)
}

var treeContainerSymlinkError *Archive = &Archive{
	ArchiveIdentifiers: ArchiveIdentifiers{
		Sha256: MustSha256("837f0da343583b0995e51de26b6fb848103221b5fdfdd1d742746765042bd5ec"),
		Size:   920,
		Md5:    MustMd5("87683b1fb6e6fa1e3de887df15ac5de9"),
		Sha1:   MustSha1("dbf0265e662276ace37bab2917fee95edde31138"),
		Name:   "test.tar.bz2",
	},
	Files: []SubFile{
		{
			Path: "date.txt",
			File: &File{
				Sha256: MustSha256("80f3d9f67e1e3b664e50d1e932b5489b3a3e547d0a6ea97e0d0c888864d6dec6"),
				Size:   32,
				Md5:    MustMd5("679a9b342d4ce61a9df5073d14c39a07"),
				Sha1:   MustSha1("ef9d9765295a3a1ec0aae05efa2d22a51819d800"),
			},
		},
	},
	Archives: []SubArchive{
		{
			Path: "tar.utf8.tar.bz2",
			Archive: &Archive{
				ArchiveIdentifiers: ArchiveIdentifiers{
					Sha256: MustSha256("ee87767de57b973d084c9f0bcda3266bf443b8d18803ed4588064f31f9c48069"),
					Size:   519,
					Md5:    MustMd5("6f5f71b9112325ceab9d458ea23a85e3"),
					Sha1:   MustSha1("753f0d2905bbb73229177552a94efb8150729459"),
					Name:   "tar.utf8.tar.bz2",
				},
				Files: []SubFile{
					{
						Path: "/symlink",
						File: &File{
							Sha256: MustSha256("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
							Size:   0,
							Md5:    MustMd5("d41d8cd98f00b204e9800998ecf8427e"),
							Sha1:   MustSha1("da39a3ee5e6b4b0d3255bfef95601890afd80709"),
						},
					},
					{
						Path: "/crt",
						File: &File{
							Sha256: MustSha256("a33f0b33767fc513e888d765c05ca0e541c83c4908b0b1a62474ed827aa40844"),
							Size:   2106,
							Md5:    MustMd5("087176ee8cf10810f8f68f8dbd0d6632"),
							Sha1:   MustSha1("2ad82f4e421b26112b78f00983c9adedc0dea51e"),
						},
					},
				},
			},
		},
	},
}

func MustHex(hexRepresentation string) []byte {
	slice, err := hex.DecodeString(hexRepresentation)
	if err != nil {
		panic(err)
	}

	return slice
}

func TestCalculateVerificationCodes(t *testing.T) {
	type args struct {
		root *Archive
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantFVCode []byte
	}{
		{
			name: "sub-archive with symlink",
			args: args{
				root: treeContainerSymlinkError,
			},
			wantErr:    false,
			wantFVCode: MustHex("4656433200ca036383b3b7394126e7311bac9987a2d80fc42258ad61f84eaa096deb003eab"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CalculateVerificationCodes(tt.args.root); (err != nil) != tt.wantErr {
				t.Errorf("CalculateVerificationCodes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !bytes.Equal(tt.args.root.FileVerificationCode, tt.wantFVCode) {
				t.Errorf("CalculateVerificationCodes() got = %x, want %x", tt.args.root.FileVerificationCode, tt.wantFVCode)
			}
		})
	}
}
