package extract

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/pkg/errors"
)

/*
#cgo pkg-config: libarchive
#cgo pkg-config: libcrypto
#include "lib.h"
#include "stdlib.h"
*/
import "C"

var Extensions []string
var MabyeExtensions []string

const (
	archiveEOF    = 1   // Found end of archive.
	archiveOK     = 0   // Operation was successful.
	archiveRetry  = -10 // Retry might succeed.
	archiveWarn   = -20 // Partial success.
	archiveFailed = -25 // Current operation cannot complete.
	archiveFatal  = -30 // No more operations are possible.

	archiveOpenError  = -100 // Error on opening file
	archiveNextError  = -200 // Error on reading next header
	archiveCopyError  = -300 // Error from copy_data
	archiveWriteError = -400 // Error finishing archive write
)

func init() {
	Extensions = []string{".ar", ".arj", ".cpio", ".dump", ".jar", ".7z", ".zip", ".pack", ".pack2000", ".tar", ".bz2", ".gz", ".lzma", ".snz", ".xz", ".z", ".tgz", ".rpm", ".gem", ".deb", ".whl", ".apk"}
}

func IsExtractable(file string) float64 {
	fileName, _, ext := SplitExt(file)

	if ext == ".pack" {
		var hasIdx bool
		var inObjectsDir bool

		if _, err := os.Stat(fileName + ".idx"); err == nil {
			hasIdx = true
		}
		parentPath, _ := filepath.Split(file)
		parent := path.Base(parentPath)

		inObjectsDir = parent == "objects"

		ret := 1.0
		if hasIdx {
			ret -= 0.5
		}
		if inObjectsDir {
			ret -= 0.5
		}
		return ret
	}

	for _, v := range Extensions {
		if v == ext {
			return 1.0
		}
	}

	return 0.0
}

func RecognizeExtension(file string) bool {
	_, _, ext := SplitExt(file)

	for _, v := range Extensions {
		if v == ext {
			return true
		}
	}

	return false
}

//filepath -> file name, full extension (including .tar), extension
func SplitExt(s string) (string, string, string) {
	ext := filepath.Ext(s)
	fileName := strings.TrimSuffix(s, ext)
	var fullExt string

	if filepath.Ext(fileName) == ".tar" {
		fullExt = ".tar" + ext
		fileName = strings.TrimSuffix(fileName, ".tar")
	} else {
		fullExt = ext
	}

	return fileName, fullExt, ext
}

type Extract struct {
	source     string
	filename   string
	target     string
	isEnclosed bool
}

func (e Extract) Source() string {
	return e.source
}
func (e Extract) Target() string {
	return e.target
}
func (e Extract) IsEnclosed() bool {
	return e.isEnclosed
}

func New(source, filename string) (*Extract, error) {
	if filename == "" {
		filename = source
	}
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "extract.New(%s, %s).Stat(%s)", source, filename, source)
	}

	source, err := filepath.Abs(source)
	if err != nil {
		return nil, errors.Wrapf(err, "extract.New(%s, %s).Abs(%s)", source, filename, source)
	}
	e := Extract{source, filename, "", false}

	return &e, nil
}

func NewAt(source, filename, target string) (*Extract, error) {
	if filename == "" {
		filename = source
	}
	e, err := New(source, filename)
	if err != nil {
		return e, err
	}

	if err = os.Mkdir(target, 0755); err != nil && !os.IsExist(err) {
		return e, errors.Wrapf(err, "extract.NewAt(%s, %s, %s).Mkdir(%s)", source, filename, target, target)
	}

	target, err = filepath.Abs(target)
	if err != nil {
		return e, errors.Wrapf(err, "extract.NewAt(%s, %s, %s).Abs(%s)", source, filename, target, target)
	}
	e.target = target
	return e, nil
}

func (e *Extract) Enclose() error {
	if e.isEnclosed {
		return nil
	}
	if e.source == "" {
		return errors.New("No source file provided")
	}

	f, err := os.Open(e.source)
	if err != nil {
		return errors.Wrapf(err, "Extract.Enclose().Open(%s)", e.source)
	}
	defer f.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return errors.Wrapf(err, "EXtract.Enclose().Copy(sha1.New(), %s)", e.source)
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	targetDir := filepath.Join(e.target, hash)
	if err = os.MkdirAll(targetDir, 0755); err != nil && !os.IsExist(err) {
		return errors.Wrapf(err, "Extract.Enclose().MkdirAll(%s)", targetDir)
	}
	e.target = targetDir
	e.isEnclosed = true

	return nil
}

func (e Extract) Extract() (string, error) {
	log.Debug().Str(zerolog.CallerFieldName, "extract.Extract{}.Extract()").Interface("extract", e).Msg("extracting")

	originalDirectory := ""
	if e.target != "" {
		cur, err := os.Getwd()
		if err != nil {
			log.Warn().Err(err).Str(zerolog.CallerFieldName, "extract.Extract{}.Extract()").Str("target", e.target).Msg("defaulting to \".\"")
			cur = "."
		}
		originalDirectory = cur

		err = os.Chdir(e.target)
		if err != nil {
			return "", errors.Wrapf(err, "Extract.Extract().Chdir(%s)", e.target)
		}
		// defer os.Chdir(cur)
	}

	cs := C.CString(e.source)
	n, f, ex := SplitExt(e.filename)
	var exit C.status
	if f != ex || (len(ex) > 0 && ex[1] == 'c') { // f contains tar
		exit = C.extractOne(cs, nil, nil, false)
	} else {
		cn := C.CString(n)
		exit = C.extractOne(cs, cn, nil, false)
	}
	defer C.status_free(exit)
	if exit == nil {
		log.Debug().Str(zerolog.CallerFieldName, "extract.Extract{}.Extract()").Msg("exit is null")
	}
	log.Debug().Str(zerolog.CallerFieldName, "extract.Extract{}.Extract()").Str("code", fmt.Sprintf("%d", exit.code)).Send()

	if exit.code < 0 {
		if originalDirectory != "" {
			os.Chdir(originalDirectory)
		}
		return e.target, errors.New(fmt.Sprintf("extract returned with status: %d\n%s\n%s\n", exit.code, C.GoString(exit.message), C.GoString(exit.tag)))
	}

	// ret := C.GoString(exit.tag)
	ret := e.Target()
	log.Debug().Str(zerolog.CallerFieldName, "extract.Extract{}.Extract()").Interface("extract", e).Str("target", ret).Msg("extracted")
	if originalDirectory != "" {
		os.Chdir(originalDirectory)
	}
	return ret, nil
}
