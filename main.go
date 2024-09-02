package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/joanlopez/go-selfupdate-poc/kit/selfupdate"
	"github.com/joanlopez/go-selfupdate-poc/kit/semver"
)

const (
	Version = "v1.2.0"
	Slug    = "joanlopez/go-selfupdate-poc"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "update" {
		updated, to, err := selfupdate.From(context.Background(), semver.MustParse(Version), Slug)
		if err != nil {
			switch {
			case errors.Is(err, selfupdate.ErrRepositoryNotFound):
				// THIS SHOULD NEVER HAPPEN
			case errors.Is(err, selfupdate.ErrInvalidSlug):
				// THIS SHOULD NEVER HAPPEN
			case errors.Is(err, selfupdate.ErrReleaseNotDetected):
				fmt.Println("[WARN] No release found, perhaps there is a temporary error, try again later")
				os.Exit(0)
			}
			fmt.Println("[ERROR] Failed to update:", err)
			os.Exit(1)
		}

		if updated {
			fmt.Println("Updated to", to.Version)
			os.Exit(0)
		}

		fmt.Println("Already up-to-date")
		os.Exit(0)
	}

	fmt.Println("Version:", Version)
}
