package main

import (
	"gopkg.in/mgo.v2"
	"github.com/BurntSushi/toml"
	"fmt"
)

type DBConfig struct {

}

var DBconfig DBConfig

func connect() {
	// Parse DB config TOML.
	if _, err := toml.DecodeFile("DBconfig.toml", &DBconfig); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Printf("Database configuration parsed successfully:\n -Connection URI is %s.\n", DBconfig.DBConnectionURI)


	// Connect to Mongo database.
	dbSession, err := mgo.Dial(DBconfig.DBConnectionURI)
	if err != nil {
		panic(err)
	}
	defer dbSession.Close()
}