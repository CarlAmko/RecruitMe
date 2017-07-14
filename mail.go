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
	"log"
)

type Auth struct {
	MailGunAPIKey string
	MailGunPublicAPIKey string
	MailGunDomain string
}

type Defaults struct {
	Author string
	CompanyName string
}

type EmailTemplate struct {
	Targets []string
	Subject string
	Body string
	Attachments []string
	Inputs map[string]string
}

var mg mailgun.Mailgun
var auth Auth
var defaults Defaults

var templates map[string]*EmailTemplate = make(map[string]*EmailTemplate)
const EMPTY_STRING string = ""

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
		name := strings.Replace(f.Name(), ".toml", EMPTY_STRING, -1)

		// Extract and store the template's input qualifiers.
		subjectInputs := pattern.FindAllStringSubmatch(template.Subject, -1)
		bodyInputs := pattern.FindAllStringSubmatch(template.Body, -1)
		template.Inputs = make(map[string]string)

		for _, input := range subjectInputs {
			template.Inputs[input[1]] = EMPTY_STRING
		}

		for _, input := range bodyInputs {
			template.Inputs[input[1]] = EMPTY_STRING
		}

		templates[name] = &template
	}

	return nil
}

func fillInputResponses() {
	reader := bufio.NewReader(os.Stdin)

	for _, template := range templates {
		for input := range template.Inputs {
			// Check if input information has already been provided.
			if !prefillTemplate(template, input) {
				// Otherwise, request input manually.
				fmt.Printf("Enter %s: ", input)
				text, _ := reader.ReadString('\n')
				template.Inputs[input] = strings.TrimRight(text, "\n")
			}

			data := template.Inputs[input]
			template.Subject = strings.Replace(template.Subject, fmt.Sprintf("$%s$", input), data, -1)
			template.Body = strings.Replace(template.Body, fmt.Sprintf("$%s$", input), data, -1)
		}

		fmt.Println(template.Subject)
		fmt.Println(template.Body)
	}
}

func prefillTemplate(template * EmailTemplate, input string) bool{
	if template.Inputs[input] != EMPTY_STRING {

		lower := strings.ToLower(input)
		if len(defaults.Author) > 0 && (lower == "author" || lower == "authorname") {
			template.Inputs[input] = defaults.Author
		} else if len(defaults.CompanyName) > 0 && (lower == "company" || lower == "companyname") {
			template.Inputs[input] = defaults.CompanyName
		} else if len(defaults.CompanyName) > 0 && (lower == "company" || lower == "companyname") {
			template.Inputs[input] = defaults.CompanyName
		} else {
			return false
		}
	}

	return true
}

func fillTarget(template * EmailTemplate, targetName string) {
	//@TODO
}

func main() {
	if _, err := toml.DecodeFile("config/auth.toml", &auth); err != nil {
		fmt.Printf("Error while parsing auth TOML: %s.\n", err)
		return
	}

	if _, err := toml.DecodeFile("template_defaults.toml", &defaults); err != nil {
		fmt.Printf("Error while parsing template defaults TOML: %s.\n", err)
	}

	// Connect to MailGun with parsed Auth configuration.
	mg = mailgun.NewMailgun(auth.MailGunDomain, auth.MailGunAPIKey, auth.MailGunPublicAPIKey)
	if mg == nil {
		fmt.Println("Error while attempting to make MailGun connection.")
		return
	}

	// Connect to email format database.
	if err, session := connect(); err != nil {
		fmt.Printf("Error while attempting to connect to format DB: %s\n", err)
		return
	} else {
		defer session.Close()
	}

	// Parses all email templates located in "templates" sub-directory.
	filepath.Walk("./templates", parseTemplates)

	// Generate input responses.
	fillInputResponses()

	for _, template := range templates {
		// Send message to each target.
		for _, target := range template.Targets {
			// Ensure that the target email is valid.
			 if validateEmail(target) {
				 // Fill in target information.
				 fillTarget(template, target)

				 // Format email address to target based on company.
				 err, address := formatEmail(defaults.Author, defaults.CompanyName)
				 if err != nil {
					 log.Fatal(err)
				 }
				 fmt.Println(address)

				 msg := mg.NewMessage(
					 fmt.Sprintf("%s <mailgun@%s>", defaults.Author, auth.MailGunDomain),
					 template.Subject,
					 template.Body,
					 address)

				 for _, filename := range template.Attachments {
					 msg.AddAttachment(fmt.Sprintf("attachments/%s", filename))
				 }

				 // Send the message.
				 if _, _, err := mg.Send(msg); err != nil {
					 log.Fatal(err)
				 }
			 }
		}
	}
 }