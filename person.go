package main

import (
	"fmt"
	"strings"
	"errors"
)

type Gender uint8
const (
	MALE Gender = iota
	FEMALE
	UNKNOWN
)

func splitName(name string) (error, string, string) {
	split := strings.Split(name, " ")

	if len(split) < 2 {
		return errors.New("Name not formatted correctly!"), EMPTY_STRING, EMPTY_STRING
	}
	return nil, strings.Title(split[0]), strings.Title(split[1])
}

func generateSalutationFromName(name string, gender Gender) (error, string){
	err, first, last := splitName(name)
	if err != nil {
		var address string
		if gender == MALE {
			address = "Mr."
		} else if gender == FEMALE {
			address = "Ms."
		} else {
			address = first
		}
		return nil, fmt.Sprintf("%s %s", address, last)
	}

	return err, EMPTY_STRING
}