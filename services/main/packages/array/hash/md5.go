package hash

import (
	"bytes"
	"crypto/md5"
	"database/sql/driver"
	"encoding/hex"

	"github.com/pkg/errors"
)

type Md5 [16]byte

func ParseMd5(s string) (*Md5, error) {
	var ret Md5
	if _, err := hex.Decode(ret[:], []byte(s)); err != nil {
		err = errors.Wrapf(err, "error parsing %s", s)
		return nil, err
	}

	return &ret, nil
}

func (h Md5) IsValid() bool {
	return bytes.Compare(h.Bytes(), make([]byte, 32, 32)) != 0
}

func (h Md5) IsEmpty() bool {
	return bytes.Compare(h.Bytes(), emptyMd5().Bytes()) == 0
}

func emptyMd5() *Md5 {
	var h Md5 = md5.Sum(nil)
	return &h
}

func (h *Md5) Scan(value interface{}) error {
	var err error
	if value == nil {
		h = emptyMd5()
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

	return errors.Wrapf(err, "failed to scan Md5")
}

func (h Md5) Value() (driver.Value, error) {
	return h.Bytes(), nil
}

func (h Md5) Bytes() []byte {
	return h[:]
}

func (h Md5) Array() [16]byte {
	return h
}

func (h Md5) Hex() string {
	return hex.EncodeToString(h.Bytes())
}
