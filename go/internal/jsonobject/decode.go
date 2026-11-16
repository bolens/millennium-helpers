// Package jsonobject decodes configuration documents without rounding numbers.
package jsonobject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Decode requires exactly one non-null JSON object and preserves numeric text.
func Decode(data []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil {
		return nil, err
	}
	if object == nil {
		return nil, fmt.Errorf("expected a non-null JSON object")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("expected exactly one JSON object")
	}
	return object, nil
}
