package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fuegoio/sunred/go/sdk/sunred"
)

// parseInt64 parses a string to int64, exiting on error.
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// printError prints an API error response to stderr.
func printError(status int, errModel *sunred.ErrorModel) {
	if errModel != nil && errModel.Detail != nil {
		fmt.Fprintf(os.Stderr, "error (%d): %s\n", status, *errModel.Detail)
		return
	}
	if errModel != nil && errModel.Title != nil {
		fmt.Fprintf(os.Stderr, "error (%d): %s\n", status, *errModel.Title)
		return
	}
	fmt.Fprintf(os.Stderr, "error (%d)\n", status)
}
