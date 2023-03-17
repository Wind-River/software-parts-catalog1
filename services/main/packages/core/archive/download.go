package archive

import (
	"io"

	"wrs/tk/packages/blob/file"

	"github.com/pkg/errors"
)

func (p *ArchiveController) DownloadTo(archive *Archive, destination io.Writer) error {
	f, err := p.archiveStorage.Retrieve(archive.Sha256.Array())
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
	f, err := p.archiveStorage.Retrieve(archive.Sha256.Array())
	if err != nil {
		return f, err
	}

	return f, nil
}
