package filename

import (
	"fmt"
	"os/exec"
	"strings"

	_ "embed"

	"github.com/pkg/errors"
)

// the original python code
//
//go:embed src_analysis_utils_functions.py
var getPkgNameVersionSrc string

// bit of script to append to the original that calls getPkgNameVersion
var getPkgNameVersionMain string = `
name, version = getPkgNameVersion(sys.argv[1])
print(f"{name}\n{version}", end="")
`

// the assembled script to run
var getPkgNameVersionScript string

func init() {
	getPkgNameVersionScript = getPkgNameVersionSrc + getPkgNameVersionMain
}

// GetPkgNameVersion returns package name and version based on the package name.
// Output will be in lowercase
func GetPkgNameVersion(filename string) (name string, version string, err error) {
	if filename == "" {
		// return empty strings without having to invoke shell
		return "", "", nil
	}

	cmd := exec.Command("python3", "-c", getPkgNameVersionScript, filename)

	// capture stdout to parse name and version
	var stdoutPipe strings.Builder
	cmd.Stdout = &stdoutPipe
	// capture stderr in case of error
	var stderrPipe strings.Builder
	cmd.Stderr = &stderrPipe

	if err := cmd.Run(); err != nil {
		// use stderr as error message
		if stderr := stderrPipe.String(); stderr != "" {
			return "", "", errors.Wrapf(err, stderr)
		}

		// if nothing in stderr, use generic error message
		return "", "", errors.Wrapf(err, "error running command")
	}

	// stdout expected to be name\nversion
	lines := strings.Split(string(stdoutPipe.String()), "\n")

	if len(lines) != 2 {
		return "", "", errors.New(fmt.Sprintf("Unexpected lines: %#v", lines))
	}

	return lines[0], lines[1], nil
}
