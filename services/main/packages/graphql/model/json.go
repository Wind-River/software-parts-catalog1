package model

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

// Json, with MarshalJson and UnmarshalJson, is the type we use to prevent gqlgen from marshalling our JSON fields into a quoted string
type Json json.RawMessage

func MarshalJson(j Json) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, string(j))
	})
}

func UnmarshalJson(v interface{}) (Json, error) {
	switch v := v.(type) {
	case string:
		return Json(v), nil
	case []byte:
		return Json(v), nil
	case json.RawMessage:
		return Json(v), nil
	default:
		return nil, fmt.Errorf("unexpected type for Json: %t", v)
	}
}
