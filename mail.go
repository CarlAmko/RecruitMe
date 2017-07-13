package RecruitMe

import (
	"gopkg.in/mailgun/mailgun-go.v1"
	"github.com/BurntSushi/toml"
	"fmt"
)

type Auth struct {
	MailGunAPIKey string
	MailGunPublicAPIKey string
	MailGunDomain string
}

type Config struct {
	AttachResume bool
	ResumeFormat string
}

// Mailgun instance object
var mg mailgun.Mailgun

// Auth object
var auth Auth

// Config object
var config Config

func main() {
	if _, err := toml.DecodeFile("auth.toml", &auth); err != nil {
		fmt.Printf("Error while parsing auth TOML: %s.\n", err)
		return
	}

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Printf("Error while parsing config TOML: %s.\n", err)
		return
	}

	mg = mailgun.NewMailgun(auth.MailGunDomain, auth.MailGunAPIKey, auth.MailGunPublicAPIKey)


}