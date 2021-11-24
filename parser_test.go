package envparser

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func getLookupFn(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

type addone int

func (a *addone) Parse(e Env) error {
	if err := json.Unmarshal([]byte(e), a); err != nil {
		return err
	}
	*a += 1
	return nil
}

type sliceAdder []int

func (a sliceAdder) Parse(e Env) error {
	var v int
	if err := json.Unmarshal([]byte(e), &v); err != nil {
		return err
	}
	for i := range a {
		a[i] += v
	}
	return nil
}

func TestParser(t *testing.T) {
	fakeEnv := map[string]string{
		"TestKey":    "test value",
		"TestKey2":   "12",
		"Composite":  `{"A": 11, "B": true}`,
		"ADD_ONE":    "22",
		"SliceAdder": "4",
	}

	var config struct {
		TestKey    string
		TestKey2   int
		TestKey3   bool
		unexported string
		Composite  struct {
			A int
			B bool
		}
		AddOne     addone `env:"ADD_ONE"`
		SliceAdder sliceAdder
	}
	config.SliceAdder = []int{1, 2, 3}

	err := parse(&config, getLookupFn(fakeEnv))
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	if config.TestKey != "test value" ||
		config.TestKey2 != 12 ||
		config.Composite.A != 11 ||
		config.Composite.B != true ||
		config.AddOne != 23 ||
		config.SliceAdder[0] != 5 ||
		config.SliceAdder[1] != 6 ||
		config.SliceAdder[2] != 7 {
		t.FailNow()
	}
}

func TestRealEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "12")
	defer os.Unsetenv("TEST_KEY")

	var config struct {
		Asdf int `env:"TEST_KEY"`
	}

	err := Parse(&config)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	if config.Asdf != 12 {
		t.FailNow()
	}
}

func TestInvalidTarget(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.FailNow()
		}
	}()

	var config map[string]string

	Parse(&config)
}

func TestError(t *testing.T) {
	fakeEnv := map[string]string{
		"TestKey2":   "aa",
		"ADD_ONE":    "aa",
		"SliceAdder": "aa",
	}

	var config struct {
		TestKey2   int
		AddOne     addone `env:"ADD_ONE"`
		SliceAdder sliceAdder
	}
	config.TestKey2 = 22
	config.AddOne = 44
	config.SliceAdder = []int{1, 2, 3}

	err := parse(&config, getLookupFn(fakeEnv))
	if err == nil || err.Error() == "" {
		t.FailNow()
	}

	var parseError *ParseError
	if !errors.As(err, &parseError) {
		t.FailNow()
	}

	if config.TestKey2 != 22 ||
		config.AddOne != 44 ||
		config.SliceAdder[0] != 1 ||
		config.SliceAdder[1] != 2 ||
		config.SliceAdder[2] != 3 {
		t.FailNow()
	}
}
