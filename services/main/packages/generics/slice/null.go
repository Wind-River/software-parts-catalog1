package slice

import (
	"fmt"

	"github.com/pkg/errors"
)

type NullSlice[E any] []E

func (slice *NullSlice[E]) Scan(value interface{}) error {
	var err error
	if value == nil {
		slice = nil
		return nil
	}

	switch t := value.(type) {
	case []interface{}:
		var ret []E

		for i, v := range t {
			sqlValue, ok := v.(E)
			if !ok {
				return errors.New(fmt.Sprintf("error scanning element %d", i))
			}

			if ret == nil {
				ret = make([]E, 0)
			}

			ret = append(ret, sqlValue)
		}

		if ret == nil {
			return errors.Wrapf(err, "failed to scan slice")
		}

		*slice = ret
		return nil
	default:
		return errors.New(fmt.Sprintf("expected slice: got %T: %#v", t, t))
	}
}
