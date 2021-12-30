// Package main provides a typing test
package main

import (
	"github.com/docopt/docopt-go"
	"github.com/noqqe/serra/src/serra"
)

// Main Loop
func main() {

	usage := `Serra

Usage:
  serra add <card>...
  serra list
  serra sets
  serra update

Options:
  -h --help     Show this screen.
  --version     Show version.
`

	args, _ := docopt.ParseDoc(usage)

	if args["add"].(bool) {
		serra.Add(args["<card>"].([]string))
	}

	if args["list"].(bool) {
		serra.List()
	}

	if args["sets"].(bool) {
		serra.Sets()
	}

}
