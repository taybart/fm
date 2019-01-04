package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type config struct {
	ShowHidden   bool   `json:"showHidden"`
	ColumnWidth  int    `json:"columnWidth"`
	ColumnRatios []int  `json:"columnWidths"`
	PreviewRegex string `json:"previewRegex"`
}

func loadConfig(name string) (config, error) {
	j, err := os.Open(name)
	if err != nil {
		return config{}, err
	}

	defer j.Close()

	// Default Config
	c := config{
		ShowHidden:   false,
		ColumnWidth:  -1,
		ColumnRatios: []int{2, 5, 3},
		PreviewRegex: "",
	}

	jb, err := ioutil.ReadAll(j)
	if err != nil {
		return config{}, err
	}
	json.Unmarshal(jb, &c)
	return c, nil
}
