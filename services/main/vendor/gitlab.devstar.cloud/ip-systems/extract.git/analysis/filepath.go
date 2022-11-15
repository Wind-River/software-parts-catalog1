package analysis

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// If test not in filepath, failure not expected
// Assumes everything before hash of package, may be irrelevant
// So if test after hash, expect failure
// Else doubt exists
func IsFailureExpected(filePath string, enclosed bool) (float64, error) {
	if !enclosed {
		return isFailureExpected(filePath, enclosed, "")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return 0, errors.Wrap(err, "unable to open filePath")
	}
	defer f.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, f); err != nil {
		return 0, errors.Wrap(err, "unable to read filePath")
	}

	sha1Hex := hex.EncodeToString(hash.Sum(nil))

	return isFailureExpected(filePath, enclosed, sha1Hex)
}

func isFailureExpected(filePath string, enclosed bool, sha1Hex string) (float64, error) {
	testIndex := strings.LastIndex(filePath, "test")

	if testIndex == -1 {
		return 0, nil
	}

	shaIndex := strings.Index(filePath, sha1Hex)

	if shaIndex != -1 {
		if testIndex < shaIndex {
			return 0.5, nil
		}
	}

	return 1.0, nil
}
