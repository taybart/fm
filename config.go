package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type config struct {
	ShowHidden  bool `json:"showHidden"`
	ColumnWidth int  `json:"columnWidth"`
}

func loadConfig(name string) (config, error) {
	j, err := os.Open(name)
	if err != nil {
		return config{}, err
	}

	defer j.Close()

	// Default Config
	c := config{
		ShowHidden:  false,
		ColumnWidth: 20,
	}

	jb, err := ioutil.ReadAll(j)
	if err != nil {
		return config{}, err
	}
	json.Unmarshal(jb, &c)
	return c, nil
}
