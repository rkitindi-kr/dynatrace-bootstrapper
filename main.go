package main

import (
	"os"

	bootstrapper "github.com/rkitindi-kr/dynatrace-bootstrapper/cmd"
	"github.com/spf13/afero"
)

func main() {
	err := bootstrapper.New(afero.NewOsFs()).Execute()
	if err != nil {
		os.Exit(1)
	}
}
