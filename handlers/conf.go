package handlers

import "github.com/taybart/log"

func toggleHidden() {
	log.Verbose("toggleHidden")
	conf.ShowHidden = !conf.ShowHidden
}
