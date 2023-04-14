package tree

import (
	"bytes"
	"sort"

	"gitlab.devstar.cloud/ip-systems/verification-code.git/code"
)

func CalculateVerificationCodes(root Node) error {
	if _, err := calculateVerificationCodes(root); err != nil {
		return err
	}

	return nil
}

func calculateVerificationCodes(root Node) ([][32]byte, error) {
	hasher := code.NewVersionTwo().(*code.VersionTwoHasher)

	sha256Accumulator := make([][32]byte, 0)
	for _, file := range root.GetFiles() {
		sha256Accumulator = append(sha256Accumulator, file.Sha256)
		if err := hasher.AddSha256(file.Sha256[:]); err != nil {
			return sha256Accumulator, err
		}
	}

	if root.GetNodes() != nil && len(root.GetNodes()) > 0 {
		for _, subArchive := range root.GetNodes() {
			subSha256s, err := calculateVerificationCodes(subArchive.Node)
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

	root.SetFileVerificationCode(hasher.Sum())

	if len(root.GetFiles()) == 0 && // if archive has no files
		(root.GetNodes() != nil && len(root.GetNodes()) == 1) { // and archive has one sub-archive
		// if sub-archive has same files, steal sub-archive's files and archives, and move sub-archive to duplicate
		if err := root.Merge(root.GetNodes()[0].Node); err != nil {
			return nil, err
		}
	}

	return sha256Accumulator, nil
}
