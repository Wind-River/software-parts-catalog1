package hash

import (
	"bytes"
	"crypto/sha1"
	"database/sql/driver"
	"encoding/hex"

	"github.com/pkg/errors"
)

type Sha1 [20]byte

func ParseSha1(s string) (*Sha1, error) {
	var ret Sha1
	if _, err := hex.Decode(ret[:], []byte(s)); err != nil {
		err = errors.Wrapf(err, "error parsing %s", s)
		return nil, err
	}

	return &ret, nil
}
func (h Sha1) IsValid() bool {
	return bytes.Compare(h.Bytes(), make([]byte, 20, 20)) != 0
}

func (h Sha1) IsEmpty() bool {
	return bytes.Compare(h.Bytes(), emptySha1().Bytes()) == 0
}

func emptySha1() *Sha1 {
	var h Sha1 = sha1.Sum(nil)
	return &h
}

func (h *Sha1) Scan(value interface{}) error {
	var err error
	if value == nil {
		h = emptySha1()
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

	return errors.Wrapf(err, "failed to scan Sha1")
}

func (h Sha1) Value() (driver.Value, error) {
	return h.Bytes(), nil
}

func (h Sha1) Bytes() []byte {
	return h[:]
}

func (h Sha1) Array() [20]byte {
	return h
}

func (h Sha1) Hex() string {
	return hex.EncodeToString(h.Bytes())
}
