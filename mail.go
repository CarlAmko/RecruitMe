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

type Email struct {
	Subject string
	Body string
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

		// Extract and store the template's input qualifiers.
		subjectInputs := pattern.FindAllStringSubmatch(template.Subject, -1)
		bodyInputs := pattern.FindAllStringSubmatch(template.Body, -1)
		template.Inputs = make(map[string]string)

		for _, input := range subjectInputs {
			trimmed := input[1]
			if !fillDefaults(&template, trimmed) {
				template.Inputs[trimmed] = EMPTY_STRING
			}
		}

		for _, input := range bodyInputs {
			trimmed := input[1]
			if !fillDefaults(&template, trimmed) {
				template.Inputs[trimmed] = EMPTY_STRING
			}
		}

		// Name the template after the file name.
		name := strings.Replace(f.Name(), ".toml", EMPTY_STRING, -1)
		templates[name] = &template
	}

	return nil
}

func fillInputResponses() {
	reader := bufio.NewReader(os.Stdin)

	for _, template := range templates {
		for input := range template.Inputs {
			// Check if input information has already been provided.
			if template.Inputs[input] == EMPTY_STRING {
				// Otherwise, request input manually.
				fmt.Printf("Enter %s: ", input)
				text, _ := reader.ReadString('\n')
				template.Inputs[input] = strings.TrimRight(text, "\n")
			}
		}
	}
}

func fillDefaults(template * EmailTemplate, input string) bool {
	lower := strings.ToLower(input)
	if len(defaults.Author) > 0 && (lower == "author" || lower == "authorname") {
		template.Inputs[input] = defaults.Author
	} else if len(defaults.CompanyName) > 0 && (lower == "company" || lower == "companyname") {
		template.Inputs[input] = defaults.CompanyName
	}  else {
		return false
	}

	return true
}

func generateEmail(template * EmailTemplate, targetName string) (error, *Email) {
	err, name := generateSalutationFromName(targetName)
	if err != nil {
		return err, nil
	}
	email := &Email{template.Subject, template.Body}
	for input := range template.Inputs {
		data := template.Inputs[input]
		email.Subject = strings.Replace(email.Subject, fmt.Sprintf("$%s$", input), data, -1)
		email.Body = strings.Replace(email.Body, fmt.Sprintf("$%s$", input), data, -1)
	}

	email.Subject = strings.Replace(email.Subject, "#TargetName#", name, -1)
	email.Body = strings.Replace(email.Body, "#TargetName#", name, -1)
	fmt.Println(email.Subject)
	fmt.Println(email.Body)
	return nil, email
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
				 // Fill in template information.
				 err, email := generateEmail(template, target)
				 if err != nil {
					 fmt.Println(err)
					 continue
				 }

				 // Format email address to target based on company.
				 err, address := formatEmail(target, defaults.CompanyName)
				 if err != nil {
					 fmt.Println(err)
					 continue
				 }
				 fmt.Println(address)

				 msg := mg.NewMessage(
					 fmt.Sprintf("%s <mailgun@%s>", defaults.Author, auth.MailGunDomain),
					 email.Subject,
					 email.Body,
					 strings.ToLower(address))

				 for _, filename := range template.Attachments {
					 msg.AddAttachment(fmt.Sprintf("attachments/%s", filename))
				 }

				 // Send the message.
				 if _, _, err := mg.Send(msg); err != nil {
					 fmt.Println(err)
				 }
			 }
		}
	}
 }