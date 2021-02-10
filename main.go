package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

var orgID string = "164022391512"
var token string = "X476Z2L3DRYUMT2ILGMH"

func GetOrderList(eventID string) []Order {
	url := fmt.Sprintf("https://www.eventbriteapi.com/v3/events/%s/orders/?expand=attendees", eventID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
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

	url := fmt.Sprintf("https://www.eventbriteapi.com/v3/organizations/%s/events/?time_filter=current_future", orgID)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

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

// if changing this, also change the params bellow
const formURL = "https://docs.google.com/forms/d/e/1FAIpQLSdTBOocGsTiozct34uU3NuGGUFFXhXoN0D_OlksfLWCrHtUhg/viewform"
const from = "dnalc@cshl.edu"
const htmlBody = `<html>      
<style type="text/css"> 
body{           
  background:#FFFFFF;
  font-family: Verdana, Arial, Helvetica, sans-serif;
  text-align:left;
  font-size: 1em;
  color:#055596;
  margin:10;
  padding:0;
} 
.redtext { color: #990000;}
</style>
<body>
<p>Dear Parent,</p>

<p>Please remember to complete the <a href="%s">%s DNALC Health Survey</a> by 9 AM.
Submission of the survey is required for your child's participation today.</p>

<p>If you are not able to certify  the  information on the health survey, please email: dnalc@cshl.edu and DO NOT attend class.</p>
<p>If for some reason you cannot submit the form electronically, paper copies will be available at the DNALC.</p>

<p>
Have a Great Day,<br>
The DNALC Team</p>
</body>
</html>`

func getSMTPClient() (*mail.SMTPClient, error) {
	server := mail.NewSMTPClient()

	// SMTP Server
	server.Host = "smtp.cshl.edu"
	server.Port = 25
	//server.Username = "guest"
	//server.Encryption = mail.EncryptionTLS

	// Since v2.3.0 you can specified authentication type:
	// - PLAIN (default)
	// - LOGIN
	// - CRAM-MD5
	//server.Authentication = mail.AuthLogin

	//Set your smtpClient struct to keep alive connection
	server.KeepAlive = true

	// Timeout for connect to SMTP Server
	server.ConnectTimeout = 5 * time.Second

	// Timeout for send the data and wait respond
	server.SendTimeout = 5 * time.Second

	// Set TLSConfig to provide custom TLS configuration. For example,
	// to skip TLS verification (useful for testing):
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// SMTP client
	return server.Connect()
}

func sendEmail(smtpClient *mail.SMTPClient, today time.Time, formURL, to string, ccs ...string) {
	email := mail.NewMSG()
	email.SetFrom(from).
		AddTo(to).
		SetSubject(fmt.Sprintf("%s DNALC Health Survey", today.Format("Jan 2")))
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

func main() {
	var smtpClient *mail.SMTPClient

	today := time.Now()
	todayStr := today.Format("2006-01-02")
	//todayStr = "2021-02-16"
	fmt.Printf("today = %+v\n", today)

	events := GetEventList()
	//fmt.Printf("found %d events\n", len(events))
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

		// this is oonly for the form defined above..
		params := url.Values{}
		params.Add("usp", "pp_url")
		params.Add("entry.479301265", e.Name.Text)
		fullFormURL := formURL + "?" + params.Encode()
		fmt.Println("url: ", fullFormURL)
		if smtpClient == nil {
			smtpClient, err = getSMTPClient()
			if err != nil {
				log.Fatalln(err)
			}
		}
		//sendEmail(smtpClient, today, fullFormURL, "ghiban@cshl.edu", "oo@mm.ll")

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
			//uniqAddresses = []string{"ghiban@cshl.edu"} //, "xx@zz.yy", "user@example.com"}
			sendEmail(smtpClient, today, fullFormURL, uniqAddresses[0], uniqAddresses[1:]...)
			//break
		}
		//break
	}

}
