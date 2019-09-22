package hlsdl

import (
	"io/ioutil"
	"os"
)

func Decrypt(segment *Segment) ([]byte, error) {

	file, err := os.Open(segment.Path)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(file)
}
