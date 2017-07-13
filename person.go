package main

import (
	"fmt"
	"strings"
)

type Gender uint8
const (
	MALE Gender = iota
	FEMALE
)

func generateSalutationFromName(name string, gender Gender) string {
	surname := strings.SplitAfter(name, " ")[0]

	address := "Ms."
	if gender == MALE {
		address = "Mr."
	}
	return fmt.Sprintf("%s %s", address, surname)
}