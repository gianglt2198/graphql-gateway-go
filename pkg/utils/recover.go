package utils

import "fmt"

func RecoverFn() {
	if r := recover(); r != nil {
		fmt.Println("Recovered in f", r)
	}
}
