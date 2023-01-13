package envparser

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

type addOne int

func (a *addOne) UnmarshalEnv(e string) error {
	if err := json.Unmarshal([]byte(e), a); err != nil {
		return err
	}
	*a += 1
	return nil
}

type addSlice []int

func (a addSlice) UnmarshalEnv(e string) error {
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
		"TestKey":     "test value",
		"TestKey2":    "12",
		"unexported":  "some text",
		"Composite":   `{"A": 11, "B": true}`,
		"ADD_ONE":     "22",
		"AddSlice":    "4",
		"Time":        "2021-09-14T12:13:14.123123+09:00",
		"StringSlice": "a, b, c",
		"IntSlice":    "1,2,3",
		"Dur":         "1m30s",
		"Loc":         "Asia/Jakarta",
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
		TestKey    string
		TestKey2   int
		TestKey3   bool
		unexported string
		Composite  struct {
			A int
			B bool
		}
		AddOne      addOne `env:"ADD_ONE"`
		AddSlice    addSlice
		Time        time.Time
		StringSlice []string
		IntSlice    []int
		Dur         time.Duration
		Loc         *time.Location
	}
	config.AddSlice = []int{1, 2, 3}
	config.unexported = "unexported"

	err := Unmarshal(&config)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	if config.TestKey != "test value" ||
		config.TestKey2 != 12 ||
		config.TestKey3 != false ||
		config.unexported != "unexported" ||
		config.Composite.A != 11 ||
		config.Composite.B != true ||
		config.AddOne != 23 ||
		config.AddSlice[0] != 5 ||
		config.AddSlice[1] != 6 ||
		config.AddSlice[2] != 7 ||
		config.Time.UTC() != time.Date(2021, 9, 14, 3, 13, 14, 123123000, time.UTC) ||
		config.StringSlice[0] != "a" ||
		config.StringSlice[1] != "b" ||
		config.StringSlice[2] != "c" ||
		config.IntSlice[0] != 1 ||
		config.IntSlice[1] != 2 ||
		config.IntSlice[2] != 3 ||
		config.Dur != 1*time.Minute+30*time.Second ||
		config.Loc.String() != "Asia/Jakarta" {
		t.FailNow()
	}
}

func TestRealEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "12")
	defer os.Unsetenv("TEST_KEY")

	var config struct {
		Asdf int `env:"TEST_KEY"`
	}

	err := Unmarshal(&config)
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

	Unmarshal(&config)
}

func TestError(t *testing.T) {
	fakeEnv := map[string]string{
		"TestKey2":   "aa",
		"ADD_ONE":    "aa",
		"SliceAdder": "aa",
		"Time":       "aa",
		"IntSlice":   "1,aa,3",
		"Dur":        "aa",
		"Loc":        "Asia/Somewhere",
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
		TestKey2   int
		AddOne     addOne `env:"ADD_ONE"`
		SliceAdder addSlice
		Time       time.Time
		IntSlice   []int
		Dur        time.Duration
		Loc        *time.Location
	}
	config.TestKey2 = 22
	config.AddOne = 44
	config.SliceAdder = []int{1, 2, 3}
	config.Dur = 3 * time.Minute

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

	defTime := time.Time{}

	if config.TestKey2 != 22 ||
		config.AddOne != 44 ||
		config.SliceAdder[0] != 1 ||
		config.SliceAdder[1] != 2 ||
		config.SliceAdder[2] != 3 ||
		config.Time != defTime ||
		len(config.IntSlice) != 0 ||
		config.Dur != 3*time.Minute ||
		config.Loc != nil {
		t.FailNow()
	}
}

func TestListEnvName(t *testing.T) {
	var config struct {
		TestKey2   int
		AddOne     addOne `env:"ADD_ONE"`
		SliceAdder addSlice
		Time       time.Time
		IntSlice   []int
		Dur        time.Duration
		Loc        *time.Location
	}
	names := ListEnvName(&config)
	if !reflect.DeepEqual(names, []string{
		"TestKey2",
		"ADD_ONE",
		"SliceAdder",
		"Time",
		"IntSlice",
		"Dur",
		"Loc",
	}) {
		t.Fatalf("invalid ListEnvName")
	}
}
