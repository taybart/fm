package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type config struct {
	ShowHidden   bool   `json:"showHidden"`
	WrapText     bool   `json:"wrapText"`
	ColumnWidth  int    `json:"columnWidth"`
	ColumnRatios []int  `json:"columnRatios"`
	PreviewRegex string `json:"previewRegex"`
	JumpAmount   int    `json:"jumpAmount"`
	Folder       string `json:"folder"`
}

func loadConfig(name string) (config, error) {
	home := os.Getenv("HOME")
	if _, err := os.Stat(home + "/.config/fm"); os.IsNotExist(err) {
		err := os.MkdirAll(home+"/.config/fm", os.ModePerm)
		if err != nil {
			return config{}, err
		}
	}
	j, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return config{}, err
	}

	defer j.Close()

	// Default Config
	c := config{
		ShowHidden:   false,
		WrapText:     true,
		ColumnWidth:  -1,
		ColumnRatios: []int{2, 3, 5},
		JumpAmount:   5,
		PreviewRegex: "",
		Folder:       home + "/.config/fm",
	}

	jb, err := ioutil.ReadAll(j)
	if err != nil {
		return config{}, err
	}
	json.Unmarshal(jb, &c)
	return c, nil
}
