package archive

import (
	"io"

	"wrs/tk/packages/blob/file"

	"github.com/pkg/errors"
)

func (p *ArchiveController) DownloadTo(archive *Archive, destination io.Writer) error {
	sha256, err := file.ParseSha256(archive.Sha256.String)
	if err != nil {
		return err
	} else if sha256 == nil {
		return errors.New("sha256 is nil")
	}

	f, err := p.archiveStorage.Retrieve(*sha256)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(destination, f); err != nil {
		return errors.Wrapf(err, "error copying file")
	}

	return nil
}

func (p *ArchiveController) Download(archive *Archive) (*file.File, error) {
	sha256, err := file.ParseSha256(archive.Sha256.String)
	if err != nil {
		return nil, err
	} else if sha256 == nil {
		return nil, errors.New("sha256 is nil")
	}

	f, err := p.archiveStorage.Retrieve(*sha256)
	if err != nil {
		return f, err
	}

	return f, nil
}
