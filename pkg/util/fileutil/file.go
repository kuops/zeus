package fileutil

import (
	"io/ioutil"
)

func FileContent(fileName string) (string, error) {
	bytes, err := ioutil.ReadFile(fileName)
	return string(bytes), err
}
