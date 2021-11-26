package envparser

import "encoding/base64"

type B64UrlString string

func (s *B64UrlString) UnmarshalEnv(val string) error {
	data, err := base64.RawURLEncoding.DecodeString(val)
	if err != nil {
		return err
	}
	*s = B64UrlString(data)
	return nil
}
