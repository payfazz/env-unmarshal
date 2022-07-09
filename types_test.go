package envparser

import (
	"errors"
	"os"
	"testing"
)

func TestParserTypes(t *testing.T) {
	fakeEnv := map[string]string{
		"b64":       "YXNkZg",
		"FileBytes": "testdata/test.txt",
	}

	for k, v := range fakeEnv {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range fakeEnv {
			os.Unsetenv(k)
		}
	}()

	var config struct {
		B64       Base64 `env:"b64"`
		FileBytes File
	}

	err := Unmarshal(&config)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	if string(config.B64) != "asdf" ||
		string(config.FileBytes) != "hello\n" {
		t.FailNow()
	}
}

func TestTypesError(t *testing.T) {
	fakeEnv := map[string]string{
		"b64":       "a",
		"FileBytes": "testdata/nonexisted",
	}

	for k, v := range fakeEnv {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range fakeEnv {
			os.Unsetenv(k)
		}
	}()

	var config struct {
		B64       Base64 `env:"b64"`
		FileBytes File
	}

	err := Unmarshal(&config)
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

	if string(config.B64) != "" ||
		string(config.FileBytes) != "" {
		t.FailNow()
	}
}
