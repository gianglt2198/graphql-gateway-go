package utils

import (
	"os"

	"golang.org/x/mod/modfile"
)

func GetModuleName() (string, error) {
	goModBytes, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	modName := modfile.ModulePath(goModBytes)

	return modName, nil
}
