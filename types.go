package envparser

import (
	"encoding/base64"
	"os"
)

type B64UrlString string

func (s *B64UrlString) UnmarshalEnv(val string) error {
	data, err := base64.RawURLEncoding.DecodeString(val)
	if err != nil {
		return err
	}
	*s = B64UrlString(data)
	return nil
}

type BytesFromFile []byte

func (b *BytesFromFile) UnmarshalEnv(val string) error {
	data, err := os.ReadFile(val)
	if err != nil {
		return err
	}
	*b = BytesFromFile(data)
	return nil
}

type StringFromFile string

func (s *StringFromFile) UnmarshalEnv(val string) error {
	var b BytesFromFile
	if err := b.UnmarshalEnv(val); err != nil {
		return err
	}
	*s = StringFromFile(b)
	return nil
}
