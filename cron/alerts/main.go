package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/cloudflare/cloudflare-go"
)

// SESCharSet is the set to use for AWS SES emails
const SESCharSet = "UTF-8"

// EnvListDelimiter is the environment variable for
// the character that splits environment variables with list values
const EnvListDelimiter = ","

// EnvKeyCFApiToken is the environment variable for
// the API token needed to access the Cloudflare API
const EnvKeyCFApiToken = "CF_API_TOKEN"

// EnvKeyCFContainsStrings is the environment variable for
// the substrings (comma separated) that this app should be using to identify
// certain Cloudflare record names
const EnvKeyCFContainsStrings = "CF_CONTAINS_STRINGS"

// EnvKeyCFZoneNames is the environment variable for
// the zone names you want the app to get records from (comma separated list)
const EnvKeyCFZoneNames = "CF_ZONE_NAMES"

// EnvKeySESSubject is the environment variable for
// the subject text of the emails that get sent out
const EnvKeySESSubject = "SES_SUBJECT"

// EnvKeySESReturnToAddress is the environment variable for
// the return-to address of the emails that get sent out
const EnvKeySESReturnToAddress = "SES_RETURN_TO_ADDR"

// EnvKeySESRecipients is the environment variable for
// the list of email addresses (comma separated) that the emails should get sent to
const EnvKeySESRecipients = "SES_RECIPIENT_EMAILS"

// EnvKeyAWSRegion is the environment variable for
// the AWS region where the lambda should ne run
const EnvKeyAWSRegion = "SES_AWS_REGION"

func getSESAWSRegion() string {
	region := os.Getenv(EnvKeyAWSRegion)
	if region == "" {
		region = "us-east-1"
	}
	return region
}

func splitStringList(compoundValue string) []string {
	initialList := strings.Split(compoundValue, EnvListDelimiter)
	output := []string{}
	for _, s := range initialList {
		output = append(output, strings.Trim(s, " "))
	}
	return output
}

func getZoneNames(a *AlertsConfig) error {
	gluedZoneNames := os.Getenv(EnvKeyCFZoneNames)
	if gluedZoneNames == "" {
		return fmt.Errorf("required value missing for environment variable %s", EnvKeyCFZoneNames)
	}
	a.CFZoneNames = splitStringList(gluedZoneNames)

	return nil
}

func getRecipientAddresses(a *AlertsConfig) error {
	gluedRecipients := os.Getenv(EnvKeySESRecipients)
	if gluedRecipients == "" {
		return fmt.Errorf("required value missing for environment variable %s", EnvKeySESRecipients)
	}

	a.RecipientEmails = splitStringList(gluedRecipients)
	return nil
}

func getCFContainsStrings(a *AlertsConfig) error {
	gluedSearchStrings := os.Getenv(EnvKeyCFContainsStrings)
	if gluedSearchStrings == "" {
		return fmt.Errorf("required value missing for environment variable %s", EnvKeyCFContainsStrings)
	}

	a.CFContainsStrings = splitStringList(gluedSearchStrings)
	return nil
}

func getRequiredString(envKey string, configEntry *string) error {
	if *configEntry != "" {
		return nil
	}

	value := os.Getenv(envKey)
	if value == "" {
		return fmt.Errorf("required value missing for environment variable %s", envKey)
	}
	*configEntry = value

	return nil
}

func (a *AlertsConfig) init() error {
	if err := getRequiredString(EnvKeyCFApiToken, &a.CFApiToken); err != nil {
		return err
	}

	if len(a.CFContainsStrings) < 1 {
		if err := getCFContainsStrings(a); err != nil {
			return err
		}
	}

	if len(a.CFZoneNames) < 1 {
		if err := getZoneNames(a); err != nil {
			return err
		}
	}

	if len(a.RecipientEmails) < 1 {
		if err := getRecipientAddresses(a); err != nil {
			return err
		}
	}

	if a.SESCharSet == "" {
		a.SESCharSet = SESCharSet
	}

	if err := getRequiredString(EnvKeySESReturnToAddress, &a.SESReturnToAddr); err != nil {
		return err
	}

	if a.SESAWSRegion == "" {
		a.SESAWSRegion = getSESAWSRegion()
	}

	if err := getRequiredString(EnvKeySESSubject, &a.SESSubjectText); err != nil {
		return err
	}

	return nil
}

type AlertsConfig struct {
	CFApiToken        string   `json:"CFApiToken"`
	CFZoneNames       []string `json:"CFZoneNames"`
	CFContainsStrings []string `json:"CFContainsString"`
	SESAWSRegion      string   `json:"SESAWSRegion"`
	SESCharSet        string   `json:"SESCharSet"`
	SESReturnToAddr   string   `json:"SESReturnToAddr"`
	SESSubjectText    string   `json:"SESSubjectText"`
	RecipientEmails   []string `json:"RecipientEmails"`
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

func getCFRecords(config AlertsConfig) map[string][]string {
	api, err := cloudflare.NewWithAPIToken(config.CFApiToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting scan for %v", config.CFZoneNames)

	results := map[string][]string{}

	for _, zoneName := range config.CFZoneNames {
		zoneID, err := api.ZoneIDByName(zoneName)
		if err != nil {
			err = fmt.Errorf("error getting Cloudflare zone %s ... %v ", zoneName, err.Error())
			sendErrorEmails(config, err)
			continue
		}

		// Fetch all records for a zone
		recs, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID),
			cloudflare.ListDNSRecordsParams{})
		if err != nil {
			err = fmt.Errorf("error getting Cloudflare dns records for zone %s ... %v ", zoneName, err.Error())
			sendErrorEmails(config, err)
			continue
		}

		for _, ss := range config.CFContainsStrings {
			ss = strings.Trim(ss, " ")
			getCFRecordsWithSubstring(ss, zoneName, recs, results)
		}
	}

	return results
}

func sendAnEmail(emailMsg ses.Message, recipient *string, config AlertsConfig) error {
	recipients := []*string{recipient}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: recipients,
		},
		Message: &emailMsg,
		Source:  aws.String(config.SESReturnToAddr),
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.SESAWSRegion),
	},
	)

	// Create an SES session.
	svc := ses.New(sess)
	result, err := svc.SendEmail(input)
	log.Println(result)
	return err
}

func sendEmails(config AlertsConfig, cfRecords map[string][]string) {
	msg := fmt.Sprintf("%s\n", config.SESSubjectText)
	for zone, ps := range cfRecords {
		msg = fmt.Sprintf("%s\n Those found in %s", msg, zone)
		for _, p := range ps {
			msg = fmt.Sprintf("%s\n%s", msg, p)
		}
	}

	subject := config.SESSubjectText

	emailMsg := getSESMessage(config, subject, msg)

	// Only report the last email error
	lastError := ""
	badRecipients := []string{}

	// Send emails to one recipient at a time to avoid one bad email sabotaging it all
	for _, address := range config.RecipientEmails {
		err := sendAnEmail(emailMsg, &address, config)
		if err != nil {
			lastError = err.Error()
			badRecipients = append(badRecipients, address)
		}
	}

	logLastError(config, lastError, badRecipients)
}

func sendErrorEmails(config AlertsConfig, err error) {
	subject := "error attempting to scan Cloudflare."
	msg := fmt.Sprintf("The Cloudflare scanner failed with the following error. \n%s", err)

	emailMsg := getSESMessage(config, subject, msg)

	// Only report the last email error
	lastError := ""
	badRecipients := []string{}

	// Send emails to one recipient at a time to avoid one bad email sabotaging it all
	for _, address := range config.RecipientEmails {
		err := sendAnEmail(emailMsg, &address, config)
		if err != nil {
			lastError = err.Error()
			badRecipients = append(badRecipients, address)
		}
	}

	logLastError(config, lastError, badRecipients)
}

func getSESMessage(config AlertsConfig, subject, msg string) ses.Message {
	charSet := config.SESCharSet

	subjContent := ses.Content{
		Charset: &charSet,
		Data:    &subject,
	}

	msgContent := ses.Content{
		Charset: &charSet,
		Data:    &msg,
	}

	msgBody := ses.Body{
		Text: &msgContent,
	}

	emailMsg := ses.Message{}
	emailMsg.SetSubject(&subjContent)
	emailMsg.SetBody(&msgBody)

	return emailMsg
}

func logLastError(config AlertsConfig, lastError string, badRecipients []string) {
	if lastError == "" {
		return
	}

	addresses := strings.Join(badRecipients, ", ")
	msg := fmt.Sprintf(
		"\nError sending Cloudflare scanner email from %s to \n %s: \n %s",
		*aws.String(config.SESReturnToAddr),
		addresses,
		lastError,
	)
	log.Print(msg)
}

func handler(config AlertsConfig) error {
	if err := config.init(); err != nil {
		return err
	}

	cfRecords := getCFRecords(config)

	if len(cfRecords) < 1 {
		log.Printf("\n No records found in Cloudflare containing any of these: %v", config.CFContainsStrings)
		return nil
	}

	sendEmails(config, cfRecords)
	return nil
}

func main() {
	lambda.Start(handler)
}
