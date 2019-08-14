package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config uration
type Config struct {
	ShowHidden   bool   `json:"showHidden"`
	WrapText     bool   `json:"wrapText"`
	ColumnWidth  int    `json:"columnWidth"`
	ColumnRatios []int  `json:"columnRatios"`
	PreviewRegex string `json:"previewRegex"`
	JumpAmount   int    `json:"jumpAmount"`
	Folder       string `json:"folder"`
}

// Load configuration file
func Load(name string) (c *Config, err error) {
	home := os.Getenv("HOME")
	if _, err = os.Stat(home + "/.config/fm"); os.IsNotExist(err) {
		err = os.MkdirAll(home+"/.config/fm", os.ModePerm)
		if err != nil {
			return
		}
	}
	j, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}

	defer j.Close()

	// Default Config
	c = &Config{
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
		return
	}
	json.Unmarshal(jb, &c)
	return
}
