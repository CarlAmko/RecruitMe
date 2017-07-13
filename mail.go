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

		fmt.Println(template.Subject)
		fmt.Println(template.Body)
	}
}

func fillTemplateFromConfig(template * EmailTemplate, input string) bool{
	lower := strings.ToLower(input)
	if len(defaults.Author) > 0 && (lower == "author" || lower == "authorname") {
		template.Inputs[input] = defaults.Author
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

	if _, err := toml.DecodeFile("template_defaults.toml", &defaults); err != nil {
		fmt.Printf("Error while parsing template defaults TOML: %s.\n", err)
	}

	// Connect to MailGun with parsed Auth configuration.
	mg = mailgun.NewMailgun(auth.MailGunDomain, auth.MailGunAPIKey, auth.MailGunPublicAPIKey)
	if mg == nil {
		fmt.Println("Error while attempting to make MailGun connection.")
		return
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
				 msg := mg.NewMessage(
					 fmt.Sprintf("%s <mailgun@%s>", defaults.Author, auth.MailGunDomain),
					 template.Subject,
					 template.Body,
					 target)

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