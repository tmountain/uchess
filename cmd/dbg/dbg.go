package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	uchess "github.com/tmountain/uchess/pkg"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeStr(file string, data []byte) {
	err := ioutil.WriteFile(file, data, 0644)
	check(err)
}

func main() {
	themes := []uchess.ThemeHex{uchess.ThemeBasic.Hex()}
	for _, theme := range themes {
		b, err := json.Marshal(theme)
		filePath := fmt.Sprintf("themes/%v.json", theme.Name)
		writeStr(filePath, b)
		check(err)
	}
}
