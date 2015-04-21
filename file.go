package toml

import (
	"io/ioutil"
	"os"
)

func UnmarshalFile(filename string, v interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = Unmarshal(buf, v)

	return err
}
