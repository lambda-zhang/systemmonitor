package systemmonitor

import (
	"io/ioutil"
	"strings"
)

func readFile2String(filepath string) (string, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	return strings.Replace(string(b), "\n", "", -1), nil
}

func listdir(pathname string) ([]string, error) {
	lists := make([]string, 0)
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		lists = append(lists, fi.Name())
	}
	return lists, err
}
