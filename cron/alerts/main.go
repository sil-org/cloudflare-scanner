package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sesTypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/cloudflare/cloudflare-go"
)

// SESDefaultCharSet is the default set to use for AWS SES emails
const SESDefaultCharSet = "UTF-8"

type Scanner struct {
	SESCharSet      string
	SESReturnToAddr string
	Alerts          []Alert
}

type Alert struct {
	Title             string
	CFApiToken        string
	CFZoneNames       []string
	CFContainsStrings []string
	SESSubjectText    string
	RecipientEmails   []string
}

func newScanner() (*Scanner, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	applicationIdentifier := getEnv("APPLICATION_IDENTIFIER", "cloudflare scanner")
	configProfileIdentifier := getEnv("CONFIG_PROFILE_IDENTIFIER", "default")
	environment := getEnv("ENVIRONMENT", "prod")

	client := appconfigdata.NewFromConfig(cfg)
	session, err := client.StartConfigurationSession(ctx, &appconfigdata.StartConfigurationSessionInput{
		ApplicationIdentifier:          &applicationIdentifier,
		ConfigurationProfileIdentifier: &configProfileIdentifier,
		EnvironmentIdentifier:          &environment,
	})
	if err != nil {
		return nil, err
	}

	configuration, err := client.GetLatestConfiguration(ctx, &appconfigdata.GetLatestConfigurationInput{
		ConfigurationToken: session.InitialConfigurationToken,
	})
	if err != nil {
		return nil, err
	}

	var scanner Scanner
	return &scanner, json.Unmarshal(configuration.Configuration, &scanner)
}

func getCFRecordsWithSubstring(substring, zoneName string, recs []cloudflare.DNSRecord, results map[string][]string) {
	log.Printf("Records with '%s' in the name in zone: %s", substring, zoneName)

	subRecs := []string{}
	for _, r := range recs {
		if len(r.Name) > 0 && strings.Contains(r.Name, substring) {
			log.Print(" ", r.Name)
			subRecs = append(subRecs, r.Name+" ... "+r.Content)
		}
	}
	if len(subRecs) > 0 {
		log.Print("-----")
		results[zoneName] = subRecs
	}
}

func (s *Scanner) getCFRecords(alert Alert) map[string][]string {
	api, err := cloudflare.NewWithAPIToken(alert.CFApiToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting scan for %v", alert.CFZoneNames)

	results := map[string][]string{}

	for _, zoneName := range alert.CFZoneNames {
		zoneID, err := api.ZoneIDByName(zoneName)
		if err != nil {
			err = fmt.Errorf("error getting Cloudflare zone %s ... %v ", zoneName, err.Error())
			s.sendErrorEmails(alert, err)
			continue
		}

		// Fetch all records for a zone
		recs, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID),
			cloudflare.ListDNSRecordsParams{})
		if err != nil {
			err = fmt.Errorf("error getting Cloudflare dns records for zone %s ... %v ", zoneName, err.Error())
			s.sendErrorEmails(alert, err)
			continue
		}

		for _, ss := range alert.CFContainsStrings {
			ss = strings.Trim(ss, " ")
			getCFRecordsWithSubstring(ss, zoneName, recs, results)
		}
	}

	return results
}

func sendAnEmail(emailMsg sesTypes.Message, sender, recipient string) error {
	recipients := []string{recipient}

	input := &ses.SendEmailInput{
		Destination: &sesTypes.Destination{
			ToAddresses: recipients,
		},
		Message: &emailMsg,
		Source:  aws.String(sender),
	}

	// Create an SES session.
	svc := ses.New(ses.Options{})
	result, err := svc.SendEmail(context.Background(), input)
	log.Println(result)
	return err
}

func (s *Scanner) sendEmails(alert Alert, cfRecords map[string][]string) {
	msg := fmt.Sprintf("%s\n", alert.SESSubjectText)
	for zone, ps := range cfRecords {
		msg = fmt.Sprintf("%s\n Those found in %s", msg, zone)
		for _, p := range ps {
			msg = fmt.Sprintf("%s\n%s", msg, p)
		}
	}

	subject := alert.SESSubjectText

	emailMsg := makeSESMessage(s.SESCharSet, subject, msg)

	// Only report the last email error
	lastError := ""
	badRecipients := []string{}

	// Send emails to one recipient at a time to avoid one bad email sabotaging it all
	for _, address := range alert.RecipientEmails {
		err := sendAnEmail(emailMsg, address, s.SESReturnToAddr)
		if err != nil {
			lastError = err.Error()
			badRecipients = append(badRecipients, address)
		}
	}

	s.logLastError(lastError, badRecipients)
}

func (s *Scanner) sendErrorEmails(alert Alert, err error) {
	subject := "error attempting to scan Cloudflare."
	msg := fmt.Sprintf("The Cloudflare scanner failed with the following error. \n%s", err)

	emailMsg := makeSESMessage(s.SESCharSet, subject, msg)

	// Only report the last email error
	lastError := ""
	badRecipients := []string{}

	// Send emails to one recipient at a time to avoid one bad email sabotaging it all
	for _, address := range alert.RecipientEmails {
		err := sendAnEmail(emailMsg, address, s.SESReturnToAddr)
		if err != nil {
			lastError = err.Error()
			badRecipients = append(badRecipients, address)
		}
	}

	s.logLastError(lastError, badRecipients)
}

func makeSESMessage(charSet, subject, msg string) sesTypes.Message {
	if charSet == "" {
		charSet = SESDefaultCharSet
	}

	subjContent := sesTypes.Content{
		Charset: &charSet,
		Data:    &subject,
	}

	msgContent := sesTypes.Content{
		Charset: &charSet,
		Data:    &msg,
	}

	msgBody := sesTypes.Body{
		Text: &msgContent,
	}

	emailMsg := sesTypes.Message{
		Subject: &subjContent,
		Body:    &msgBody,
	}

	return emailMsg
}

func (s *Scanner) logLastError(lastError string, badRecipients []string) {
	if lastError == "" {
		return
	}

	addresses := strings.Join(badRecipients, ", ")
	msg := fmt.Sprintf(
		"\nError sending Cloudflare scanner email from %s to \n %s: \n %s",
		*aws.String(s.SESReturnToAddr),
		addresses,
		lastError,
	)
	log.Print(msg)
}

func handler() error {
	scanner, err := newScanner()
	if err != nil {
		return err
	}

	for _, alert := range scanner.Alerts {
		log.Printf("Starting scan for alert %q", alert.Title)
		cfRecords := scanner.getCFRecords(alert)

		if len(cfRecords) < 1 {
			log.Printf("\n No records found in Cloudflare containing any of these: %v", alert.CFContainsStrings)
			return nil
		}

		scanner.sendEmails(alert, cfRecords)
	}
	return nil
}

func main() {
	lambda.Start(handler)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
