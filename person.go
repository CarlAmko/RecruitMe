package main

import (
	"fmt"
	"strings"
	"errors"
)

type Person struct {
	Salutation string
	FirstName string
	LastName string
}

func splitName(name string) (error, *Person) {
	split := strings.Split(name, " ")

	if count := len(split); count < 2 {
		return errors.New("Name not formatted correctly!"), nil
	} else if count == 2 {
		return nil, &Person{EMPTY_STRING,strings.Title(split[0]), strings.Title(split[1])}
	} else {
		return nil, &Person{strings.Title(split[0]), strings.Title(split[1]), strings.Title(split[2])}
	}
}

func generateSalutationFromName(name string) (error, string){
	err, person := splitName(name)
	if err != nil {
		return err, EMPTY_STRING
	}

	return nil, fmt.Sprintf("%s %s", person.Salutation, person.LastName)
}