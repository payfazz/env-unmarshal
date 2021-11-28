package envparser

import (
	"errors"
	"testing"
)

func TestParserTypes(t *testing.T) {
	fakeEnv := map[string]string{
		"b64":        "YXNkZg",
		"FileBytes":  "testdata/test.txt",
		"FileString": "testdata/test.txt",
	}

	var config struct {
		B64        B64UrlString `env:"b64"`
		FileBytes  BytesFromFile
		FileString StringFromFile
	}

	err := unmarshal(&config, getLookupFn(fakeEnv))
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	if config.B64 != "asdf" ||
		string(config.FileBytes) != "hello\n" ||
		config.FileString != "hello\n" {
		t.FailNow()
	}
}

func TestTypesError(t *testing.T) {
	fakeEnv := map[string]string{
		"b64":        "a",
		"FileBytes":  "testdata/nonexisted",
		"FileString": "testdata/nonexisted",
	}

	var config struct {
		B64        B64UrlString `env:"b64"`
		FileBytes  BytesFromFile
		FileString StringFromFile
	}

	err := unmarshal(&config, getLookupFn(fakeEnv))
	if err == nil || err.Error() == "" {
		t.FailNow()
	}

	var parseError *ParseError
	if !errors.As(err, &parseError) {
		t.FailNow()
	}

	if len(parseError.Items) != len(fakeEnv) {
		for _, v := range parseError.Items {
			if _, ok := fakeEnv[v.Key]; !ok {
				t.FailNow()
			}
		}
	}

	if config.B64 != "" ||
		string(config.FileBytes) != "" ||
		config.FileString != "" {
		t.FailNow()
	}
}
