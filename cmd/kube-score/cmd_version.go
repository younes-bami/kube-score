package main

import (
	"fmt"
)

// These variables are overwritten by goreleaser
var version = "development"
var commit = "N/A"
var date = "N/A"

func cmdVersion() {
	fmt.Printf("kube-score version: %s, commit: %s, built: %s\n", version, commit, date)
}
