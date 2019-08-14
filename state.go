package main

type fm struct {
	cd            string
	dir           []pseudofile
	dt            directoryTree
	cmd           string
	cmdIndex      int
	active        pseudofile
	mode          mode
	copySource    pseudofile
	copyBufReady  bool
	moveFile      bool
	lastInput     rune
	selectedFiles map[string]pseudofile
}
