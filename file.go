package toml

import (
	"io/ioutil"
	"os"
)

// UnmarshalFile parses the TOML data in file filename on disk and stores it in
// the value pointed to by v
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
