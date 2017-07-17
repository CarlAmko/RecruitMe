package main

import (
	"gopkg.in/mgo.v2"
	"github.com/BurntSushi/toml"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"github.com/pkg/errors"
)

type DBConfig struct {
	DBConnectionURI string
}

type CompanyEntry struct {
	Company string
	Metro string
	State string
	Format string
	Domain string
	Example string
	Notes string
	Phone string
}

var DBconfig DBConfig
var formats *mgo.Collection

func connect() (error, *mgo.Session){
	// Parse DB config TOML.
	if _, err := toml.DecodeFile("config/DBconfig.toml", &DBconfig); err != nil {
		fmt.Printf("Error: %s\n", err)
		return err, nil
	}
	fmt.Printf("Database configuration parsed successfully:\n-Connection URI is %s.\n", DBconfig.DBConnectionURI)

	// Connect to Mongo database.
	dbSession, err := mgo.Dial(DBconfig.DBConnectionURI)
	if err != nil {
		return err, nil
	}

	formats = dbSession.DB("email-formats").C("formats")
	return nil, dbSession
}

func findCompaniesByName(name string) (error, []*CompanyEntry){

	query := formats.Find(bson.M{"company" : &bson.RegEx{Pattern: name, Options: "i"}})
	if count, err := query.Count(); count == 0 {
		fmt.Println(fmt.Sprintf("Company %s does not exist!", name))
		return err, nil
	}

	var entries []*CompanyEntry = []*CompanyEntry{}
	err := query.All(&entries)
	if err != nil {
		fmt.Println(err)
		return err, nil
	}

	for _, entry := range entries {
		fmt.Println(entry)
	}
	return nil, entries
}

func selectCompany(companies []*CompanyEntry) *CompanyEntry {
	//@TODO
	return companies[0]
}

func formatEmail(name string, company string) (error, string){
	err, companies := findCompaniesByName(company)
	if err != nil {
		return err, EMPTY_STRING
	}

	var selection * CompanyEntry
	if count := len(companies); count == 0 {
		return errors.Errorf("Company '%s' not found!", company), EMPTY_STRING
	} else if count == 1 {
		selection = companies[0]
	} else {
		selection = selectCompany(companies)
	}

	err, person := splitName(name)
	if err != nil {
		return err, EMPTY_STRING
	}

	lower := strings.ToLower(selection.Format)
	format := strings.Replace(lower, "firstname", person.FirstName, -1)
	format = strings.Replace(format, "lastname", person.LastName, -1)
	format = strings.Replace(format, "firstinitial", person.FirstName[:1], -1)
	format = strings.Replace(format, "lastinitial", person.LastName[:1], -1)

	return nil, fmt.Sprintf("%s@%s", strings.TrimSpace(format), strings.TrimSpace(selection.Domain))
}