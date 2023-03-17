package tree

import (
	"bytes"
	"sort"

	"gitlab.devstar.cloud/ip-systems/verification-code.git/code"
)

func CalculateVerificationCodes(root *Archive) error {
	if _, err := calculateVerificationCodes(root); err != nil {
		return err
	}

	return nil
}

func calculateVerificationCodes(root *Archive) ([][32]byte, error) {
	hasher := code.NewVersionTwo().(*code.VersionTwoHasher)

	sha256Accumulator := make([][32]byte, 0)
	for _, file := range root.Files {
		sha256Accumulator = append(sha256Accumulator, file.Sha256)
		if err := hasher.AddSha256(file.Sha256[:]); err != nil {
			return sha256Accumulator, err
		}
	}

	if root.Archives != nil && len(root.Archives) > 0 {
		for _, subArchive := range root.Archives {
			subSha256s, err := calculateVerificationCodes(subArchive.Archive)
			if err != nil {
				return sha256Accumulator, err
			}

			for i := range subSha256s { // index is used to prevent the slice we create from changing on the next iteration
				sha256Accumulator = append(sha256Accumulator, subSha256s[i])
				if err := hasher.AddSha256(subSha256s[i][:]); err != nil {
					return sha256Accumulator, err
				}
			}
		}
	}

	// root.FileVerificationCode = hasher.Sum()
	sort.Slice(sha256Accumulator, func(i, j int) bool {
		return bytes.Compare(sha256Accumulator[i][:], sha256Accumulator[j][:]) < 0
	})

	root.FileVerificationCode = hasher.Sum()

	return sha256Accumulator, nil
}
