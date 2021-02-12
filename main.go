package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type Pagination struct {
	ObjectCount  int    `json:"object_count"`
	PageNumber   int    `json:"page_number"`
	PageSize     int    `json:"page_size"`
	PageCount    int    `json:"page_count"`
	Continuation string `json:"continuation"`
	HasMoreItems bool   `json:"has_more_items"`
}

type EventMetaName struct {
	Text string `json:"text"`
	HTML string `json:"html"`
}

type EventDate struct {
	TimeZone string
	Local    string
	UTC      string
}

type Event struct {
	ID          string        `json:"id"`
	Name        EventMetaName `json:"name"`
	Start       EventDate     `json:"start"`
	End         EventDate     `json:"end"`
	Summary     string        `json:"summary"`
	Description EventMetaName `json:"description"`
	Status      string        `json:"status"`
	Listed      bool          `json:"listed"`
	IsFree      bool          `json:"is_free"`
}

type EventListResponse struct {
	Pagination Pagination `json:"pagination"`
	Events     []Event    `json:"events"`
}

type Order struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Status    string     `json:"status"`
	Attendees []Attendee `json:"attendees"`
}

type AttendeeProfile struct {
	Name string `json:"name"`
	/*FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`*/
	Email     string `json:"email"`
	CellPhone string `json:"cell_phone"`
	Status    string `json:"status"`
}

type Attendee struct {
	ID       string          `json:"id"`
	Quantity int             `json:"quantity"`
	Profile  AttendeeProfile `json:"profile"`
}

type OrderListResponse struct {
	Pagination Pagination `json:"pagination"`
	Orders     []Order    `json:"orders"`
}

func GetOrderList(eventID string) []Order {
	url := fmt.Sprintf("https://www.eventbriteapi.com/v3/events/%s/orders/?expand=attendees", eventID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.EBAPIToken))
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Errored when sending request to the server")
		return []Order{}
	}

	defer resp.Body.Close()

	var list OrderListResponse
	//respBody, _ := ioutil.ReadAll(resp.Body)
	//json.Unmarshal(respBody, &list)

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&list)

	if err != nil {
		fmt.Println(err)
		return []Order{}
	}

	return list.Orders

}

func GetEventList() []Event {
	client := &http.Client{}

	url := fmt.Sprintf("https://www.eventbriteapi.com/v3/organizations/%s/events/?time_filter=current_future", cfg.EBOrgID)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.EBAPIToken))

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Errored when sending request to the server")
		return []Event{}
	}

	defer resp.Body.Close()

	var list EventListResponse
	//respBody, _ := ioutil.ReadAll(resp.Body)
	//json.Unmarshal(respBody, &list)

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&list)

	if err != nil {
		fmt.Println(err)
		return []Event{}
	}

	return list.Events
}

func getSMTPClient() (*mail.SMTPClient, error) {
	server := mail.NewSMTPClient()

	// SMTP Server
	server.Host = cfg.SMTPServer
	server.Port = cfg.SMTPPort
	fmt.Println("cfg.UseTLS:", cfg.UseTLS)
	if cfg.UseTLS {
		server.Encryption = mail.EncryptionTLS
		// Set TLSConfig to provide custom TLS configuration. For example,
		// to skip TLS verification (useful for testing):
		server.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if cfg.UseLogin {
		// Since v2.3.0 you can specified authentication type:
		// - PLAIN (default)
		// - LOGIN
		// - CRAM-MD5
		server.Authentication = mail.AuthLogin
		server.Username = cfg.SMTPUser
		server.Password = cfg.SMTPPassword
	}

	//Set your smtpClient struct to keep alive connection
	server.KeepAlive = true

	// Timeout for connect to SMTP Server
	server.ConnectTimeout = 5 * time.Second

	// Timeout for send the data and wait respond
	server.SendTimeout = 5 * time.Second

	// SMTP client
	return server.Connect()
}

func sendEmail(smtpClient *mail.SMTPClient, today time.Time, formURL, to string, ccs ...string) {
	email := mail.NewMSG()
	email.SetFrom(cfg.From).
		AddTo(to).
		SetSubject(fmt.Sprintf(cfg.SubjectTmpl, today.Format("Jan 2")))
	for _, cc := range ccs {
		fmt.Println("adding cc ", cc)
		email.AddCc(cc)
	}

	email.SetBody(mail.TextHTML, fmt.Sprintf(htmlBody, formURL, today.Format("Jan 2")))

	// Call Send and pass the client
	err := email.Send(smtpClient)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email Sent")
	}
}

type Configuration struct {
	From             string            `json:"FROM"`
	TemplateFile     string            `json:"EMAIL_TMPL_FILE"`
	SubjectTmpl      string            `json:"EMAIL_SUBJECT_TMPL"`
	SurveyFormURL    string            `json:"FORM_URL"`
	SurveyFormParams map[string]string `json:"FORM_PARAMS"`
	SMTPServer       string            `json:"SMTP_SERVER"`
	SMTPPort         int               `json:"SMTP_PORT"`
	SMTPUser         string            `json:"SMTP_USER"`
	SMTPPassword     string            `json:"SMTP_PASS"`
	UseTLS           bool              `json:"USE_TLS"`
	UseLogin         bool              `json:"USE_LOGIN"`
	EBOrgID          string            `json:"EB_ORG_ID"`
	EBAPIToken       string            `json:"EB_API_TOKEN"`
	MockDate         string            `json:"MOCK_DATE,omitempty"`
}

var cfg Configuration // config object
var htmlBody string   // html template

func init() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to a json config file")
	flag.Parse()

	fmt.Println("configFile:", configPath)
	if configPath == "" {
		log.Println("missing config file (use -config flag)")
		os.Exit(1)
	}

	file, err := os.Open(configPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	cfg = Configuration{}
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("server: ", cfg.SMTPServer)
	//fmt.Printf("config: %+v\n", cfg)

	// read the contents of the html template
	if cfg.TemplateFile == "" {
		log.Fatalln("EMAIL_TMPL_FILE is not set in the config file")
	}
	contents, err := ioutil.ReadFile(cfg.TemplateFile)
	if err != nil {
		log.Fatalln("can't open file: ", cfg.TemplateFile, err)
	}

	htmlBody = string(contents)
}

func main() {
	var smtpClient *mail.SMTPClient

	today := time.Now()
	todayStr := today.Format("2006-01-02")
	if cfg.MockDate != "" {
		todayStr = cfg.MockDate
	}
	fmt.Printf("today = %+v\n", today)

	/*
		smtpClientX, errX := getSMTPClient()
		if errX != nil {
			fmt.Println("err:", errX)
			return
		}

		sendEmail(smtpClientX, today, "http://example.com/", "user@example.com")
		os.Exit(1)
	*/

	events := GetEventList()
	for _, e := range events {
		start, err := time.Parse("2006-01-02T15:04:05", e.Start.Local)
		if err != nil {
			fmt.Printf("Error: Unable to parse date %s for event #%s", e.Start.Local, e.ID)
			fmt.Println(err.Error())
		}
		startDate := start.Format("2006-01-02")
		if todayStr != startDate {
			//fmt.Println("** skipping ", today, " != ", startDate)
			continue
		}
		fmt.Println(e.Name.Text, "\t", e.Start.Local, "\t", e.Start.UTC)
		fmt.Printf("start = %+v\n", start)

		params := url.Values{}
		for k, v := range cfg.SurveyFormParams {
			if v == "???" {
				v = e.Name.Text
			}
			params.Add(k, v)
			//fmt.Printf("*** %s: %s\n", k, v)
		}
		fullFormURL := cfg.SurveyFormURL + "?" + params.Encode()
		if smtpClient == nil {
			smtpClient, err = getSMTPClient()
			if err != nil {
				log.Fatalln(err)
			}
		}
		fmt.Println("URL: ", fullFormURL)
		//sendEmail(smtpClient, today, fullFormURL, "user@domain.com", "oo@mm.ll")

		//continue
		//break
		orders := GetOrderList(e.ID)
		for _, o := range orders {
			//fmt.Println("\t O: ", o.Name, "\t", o.Email)
			addresses := map[string]bool{o.Email: true}
			for _, a := range o.Attendees {
				//fmt.Println("\t  -", a)
				//fmt.Println("\t A: ", a.Profile.Name, "\t", a.Profile.Email)
				addresses[a.Profile.Email] = true
			}
			uniqAddresses := make([]string, 0, len(addresses))
			for k := range addresses {
				uniqAddresses = append(uniqAddresses, k)
			}
			//fmt.Println(o.Email, " <> ", uniqAddresses)
			fmt.Println("\t*", uniqAddresses[0], " <> ", uniqAddresses[1:])
			sendEmail(smtpClient, today, fullFormURL, uniqAddresses[0], uniqAddresses[1:]...)
			time.Sleep(1 * time.Second)
			//break
		}
		time.Sleep(10 * time.Second)
		//break
	}

}
