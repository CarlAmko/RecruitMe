package main

import (
	"gopkg.in/mailgun/mailgun-go.v1"
	"github.com/BurntSushi/toml"
	"fmt"
	"regexp"
	"os"
	"path/filepath"
	"strings"
	"bufio"
)

type Auth struct {
	MailGunAPIKey string
	MailGunPublicAPIKey string
	MailGunDomain string
}

type Config struct {
	AttachResume bool
	ResumeFormat string
	Author string
}

type EmailTemplate struct {
	Subject string
	Body string
	Inputs map[string]string
}

// Mailgun instance object
var mg mailgun.Mailgun

// Auth object
var auth Auth

// Config object
var config Config

var templates map[string]*EmailTemplate = make(map[string]*EmailTemplate)

func parseTemplates(path string, f os.FileInfo, err error) error {
	// Create RegEx pattern to capture template inputs.
	pattern := regexp.MustCompile("[$]([a-zA-Z]+?)[$]")

	// Filter out non-TOML files.
	if strings.HasSuffix(f.Name(), ".toml") {

		// Decode and parse the TOML to an email template.
		var template EmailTemplate
		if _, err := toml.DecodeFile(path, &template); err != nil {
			fmt.Printf("Error while parsing auth TOML: %s.\n", err)
			return err
		}

		// Name the template after the file name.
		name := strings.Replace(f.Name(), ".toml", "", -1)

		// Extract and store the template's input qualifiers.
		subjectInputs := pattern.FindAllStringSubmatch(template.Subject, -1)
		bodyInputs := pattern.FindAllStringSubmatch(template.Body, -1)
		template.Inputs = make(map[string]string)

		for _, input := range subjectInputs {
			template.Inputs[input[1]] = ""
		}

		for _, input := range bodyInputs {
			template.Inputs[input[1]] = ""
		}

		templates[name] = &template
		fmt.Println(name)
		fmt.Println(template)
	}

	return nil
}

func fillInputResponses() {
	reader := bufio.NewReader(os.Stdin)

	for _, template := range templates {
		for input := range template.Inputs {
			// Attempt to fill input field from config.
			if !fillTemplateFromConfig(template, input) {
				// Otherwise, request input manually.
				fmt.Printf("Enter %s: ", input)
				text, _ := reader.ReadString('\n')
				template.Inputs[input] = strings.TrimRight(text, "\n")
			}

			data := template.Inputs[input]
			template.Subject = strings.Replace(template.Subject, fmt.Sprintf("$%s$", input), data, -1)
			template.Body = strings.Replace(template.Body, fmt.Sprintf("$%s$", input), data, -1)
		}

		fmt.Println(template)
	}
}

func fillTemplateFromConfig(template * EmailTemplate, input string) bool{
	lower := strings.ToLower(input)
	if lower == "author" || lower == "authorname" {
		template.Inputs[input] = config.Author
	} else {
		return false
	}

	return true
}

func main() {
	if _, err := toml.DecodeFile("auth.toml", &auth); err != nil {
		fmt.Printf("Error while parsing auth TOML: %s.\n", err)
		return
	}

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Printf("Error while parsing config TOML: %s.\n", err)
		return
	}

	// Connect to MailGun with parsed Auth configuration.
	mg = mailgun.NewMailgun(auth.MailGunDomain, auth.MailGunAPIKey, auth.MailGunPublicAPIKey)

	// Parses all email templates located in "templates" sub-directory.
	filepath.Walk("./templates", parseTemplates)

	// Generate input responses.
	fillInputResponses()

	//mg.NewMessage(
	//	fmt.Sprintf("%s <mailgun@%s>", auth.AuthorName, auth.MailGunDomain),
	//	""
	//)

}