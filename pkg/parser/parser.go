package parser

import (
	"encoding/json"
	"io"
)

func ParseBody[T any](data io.ReadCloser, o *T) error {
	body, err := io.ReadAll(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, o)
	if err != nil {
		return err
	}

	return nil
}
