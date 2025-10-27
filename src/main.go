package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sesTypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/cloudflare/cloudflare-go"
	"github.com/getsentry/sentry-go"
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
	SESCharSet        string
	SESReturnToAddr   string
	SESSubjectText    string
	RecipientEmails   []string
}

func newScanner() (*Scanner, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	applicationIdentifier := getEnv("APPLICATION_IDENTIFIER", "cloudflare-scanner")
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
	log.Printf("searching for %q in zone %q", substring, zoneName)
	var subRecs []string
	for _, r := range recs {
		if len(r.Name) > 0 && strings.Contains(r.Name, substring) {
			log.Printf("found %q in zone %q", substring, zoneName)
			subRecs = append(subRecs, r.Name+" ... "+r.Content)
		}
	}
	if len(subRecs) > 0 {
		results[zoneName] = subRecs
	}
}

func (a *Alert) getCFRecords() map[string][]string {
	api, err := cloudflare.NewWithAPIToken(a.CFApiToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("scanning zones: %s", strings.Join(a.CFZoneNames, ", "))

	results := map[string][]string{}

	for _, zoneName := range a.CFZoneNames {
		zoneID, err := api.ZoneIDByName(zoneName)
		if err != nil {
			err = fmt.Errorf("error getting Cloudflare zone %s: %w ", zoneName, err)
			a.sendErrorEmails(err)
			continue
		}

		// Fetch all records for a zone
		recs, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID),
			cloudflare.ListDNSRecordsParams{})
		if err != nil {
			err = fmt.Errorf("error getting Cloudflare dns records for zone %s: %w ", zoneName, err)
			a.sendErrorEmails(err)
			continue
		}

		for _, ss := range a.CFContainsStrings {
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
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	svc := ses.NewFromConfig(cfg)
	_, err = svc.SendEmail(context.Background(), input)
	if err != nil {
		return fmt.Errorf("send email failed: %w", err)
	}
	log.Printf("sent %q email to %q", *emailMsg.Subject.Data, recipient)
	return nil
}

func (a *Alert) sendEmails(cfRecords map[string][]string) {
	msg := fmt.Sprintf("%s\n", a.SESSubjectText)
	for zone, ps := range cfRecords {
		msg = fmt.Sprintf("%s\n Those found in %s", msg, zone)
		for _, p := range ps {
			msg = fmt.Sprintf("%s\n%s", msg, p)
		}
	}

	subject := a.SESSubjectText

	emailMsg := makeSESMessage(a.SESCharSet, subject, msg)

	// Only report the last email error
	lastError := ""
	var badRecipients []string

	// Send emails to one recipient at a time to avoid one bad email sabotaging it all
	for _, address := range a.RecipientEmails {
		err := sendAnEmail(emailMsg, address, a.SESReturnToAddr)
		if err != nil {
			log.Printf("error sending alert email %s: %s", msg, err)
			lastError = err.Error()
			badRecipients = append(badRecipients, address)
		}
	}

	if lastError != "" {
		a.logEmailError(lastError, badRecipients)
	}
}

func (a *Alert) sendErrorEmails(err error) {
	sentry.CaptureException(err)

	subject := "error attempting to scan Cloudflare."
	msg := fmt.Sprintf("The Cloudflare scanner failed with the following error. \n%s", err)

	emailMsg := makeSESMessage(a.SESCharSet, subject, msg)

	// Only report the last email error
	lastError := ""
	var badRecipients []string

	// Send emails to one recipient at a time to avoid one bad email sabotaging it all
	for _, address := range a.RecipientEmails {
		err := sendAnEmail(emailMsg, address, a.SESReturnToAddr)
		if err != nil {
			log.Printf("error sending error email %s: %s", msg, err)
			lastError = err.Error()
			badRecipients = append(badRecipients, address)
		}
	}

	if lastError != "" {
		a.logEmailError(lastError, badRecipients)
	}
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

func (a *Alert) logEmailError(errorMessage string, badRecipients []string) {
	addresses := strings.Join(badRecipients, ", ")
	msg := fmt.Sprintf("Error sending Cloudflare scanner email from %q to %q: %s",
		*aws.String(a.SESReturnToAddr),
		addresses,
		errorMessage,
	)
	log.Println(msg)
	sentry.CaptureException(errors.New(msg))
}

func handler() error {
	scanner, err := newScanner()
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	for _, alert := range scanner.Alerts {
		log.Printf("Starting scan for alert %q", alert.Title)

		if alert.SESCharSet == "" {
			alert.SESCharSet = scanner.SESCharSet
		}
		if alert.SESReturnToAddr == "" {
			alert.SESReturnToAddr = scanner.SESReturnToAddr
		}

		cfRecords := alert.getCFRecords()

		if len(cfRecords) < 1 {
			log.Printf("No records found in Cloudflare containing any of these: %v", alert.CFContainsStrings)
		} else {
			alert.sendEmails(cfRecords)
		}
	}
	return nil
}

func main() {
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		initSentry(dsn)
		defer sentry.Flush(2 * time.Second)
	}

	lambda.Start(handler)
}

func initSentry(dsn string) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: getEnv("APP_ENV", "prod"),
	})
	if err != nil {
		log.Printf("sentry.Init failure: %s", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
