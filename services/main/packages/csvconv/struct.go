// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package csvconv

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"strconv"
)

type CSVTransformer struct {
	ImplicitConversion bool // On the first data row, try int, float, or just string conversions, and record the first successful for subsequent rows
}

// ToJSONMap converts a csv to a array of field, taken from the header row, to value maps
func (transformer CSVTransformer) ToJSONMap(_r io.Reader) ([]map[string]interface{}, error) {
	r := csv.NewReader(_r)

	header, err := r.Read()
	if err != nil {
		return nil, err
	}

	var types []string
	if transformer.ImplicitConversion {
		types = make([]string, len(header))
	}

	data := make([]map[string]interface{}, 0)
	for record, err := r.Read(); err != io.EOF; record, err = r.Read() {
		if err != nil {
			return nil, err
		}
		row := make(map[string]interface{})

		if transformer.ImplicitConversion && types[0] == "" {
			// try to convert fields to int or float, if successful, record in types
			for i, v := range record {
				if n, err := strconv.ParseInt(v, 10, 64); err == nil {
					types[i] = "int"
					row[header[i]] = n
				} else if f, err := strconv.ParseFloat(v, 64); err == nil {
					types[i] = "float"
					row[header[i]] = f
				} else {
					types[i] = "string"
					row[header[i]] = v
				}
			}
		} else if transformer.ImplicitConversion {
			for i, v := range record {
				switch types[i] {
				case "int":
					n, err := strconv.ParseInt(v, 10, 64)
					if err != nil {
						return nil, err
					}
					row[header[i]] = n
				case "float":
					f, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return nil, err
					}
					row[header[i]] = f
				case "string":
					fallthrough
				default:
					row[header[i]] = v
				}
			}
		} else { // don't try converting and just treat all values as string
			for i, v := range record {
				row[header[i]] = v
			}
		}

		data = append(data, row)
	}

	return data, nil
}

// ToJSON takes an extra step by encoding the results of ToJSONMap to a byte array
func (transformer CSVTransformer) ToJSON(_r io.Reader) ([]byte, error) {
	data, err := transformer.ToJSONMap(_r)
	if err != nil {
		return nil, err
	}

	bs := make([]byte, 0)
	ret := bytes.NewBuffer(bs)
	if err = json.NewEncoder(ret).Encode(data); err != nil {
		return nil, err
	}

	return ret.Bytes(), nil
}
