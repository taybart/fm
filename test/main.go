package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	path := "/Users/taylor/dev"
	f, err := os.Open(path)
	if err != nil {
		if os.IsPermission(err) {
			return
		}
		w := fmt.Sprintf("filename: %s", path)
		fmt.Println(err, w)
	}
	defer f.Close()

	files, err := f.Readdir(0) // Or f.Readdir(1)
	if err == io.EOF {
		fmt.Println("Nothing there")
	}
	for _, f := range files {
		fmt.Printf("%+v\n", f.Name())
	}
}
