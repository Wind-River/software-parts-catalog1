package hash

import (
	"bytes"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"

	"github.com/pkg/errors"
)

type Sha256 [32]byte

func ParseSha256(s string) (*Sha256, error) {
	var ret Sha256
	if _, err := hex.Decode(ret[:], []byte(s)); err != nil {
		err = errors.Wrapf(err, "error parsing %s", s)
		return nil, err
	}

	return &ret, nil
}

func (h Sha256) IsValid() bool {
	return bytes.Compare(h.Bytes(), make([]byte, 32, 32)) != 0
}

func (h Sha256) IsEmpty() bool {
	return bytes.Compare(h.Bytes(), emptySha256().Bytes()) == 0
}

func emptySha256() *Sha256 {
	var h Sha256 = sha256.Sum256(nil)
	return &h
}

func (h *Sha256) Scan(value interface{}) error {
	var err error
	if value == nil {
		h = emptySha256()
		return nil
	}

	var sv driver.Value
	sv, err = driver.String.ConvertValue(value)
	if err == nil {
		if v, ok := sv.([]byte); ok {
			if _, err := hex.Decode(h[:], v); err != nil {
				copy(h[:], v)
			}
		}
	}

	return errors.Wrapf(err, "failed to scan Sha256")
}

func (h Sha256) Value() (driver.Value, error) {
	return h.Bytes(), nil
}

func (h Sha256) Bytes() []byte {
	return h[:]
}

func (h Sha256) Array() [32]byte {
	return h
}

func (h Sha256) Hex() string {
	return hex.EncodeToString(h.Bytes())
}
