
#Email your Eventbright events attendees

I use this to send a link to the users asking them to fill up a form with about their health (for
contract tracing).

```shell
go run main.go -config config.json
```

This tool connects to Eventbrite's API, get's the current and upcomming (up to 50) event. Then if 
today's date (can be overridden by `MOCK_DATE` ) is the same as the start day for an event, if get's 
the orders and the attendees and send them an email.

Google Forms lets you prefill a form field, so you can prefill it with the event's name (use
`FORM_PARAMS` for this).

In the config you can specify smtp server, Google form url, your Eventbright organisation ID and
an API private token.

Set the config
```json
{
    "FROM": "Office <office@organisation.org>",
    "EMAIL_TMPL_FILE" : "message-template.html",
    "FORM_URL": "https://docs.google.com/forms/d/e/1FAIpQLSd...HtUhg/viewform",
    "FORM_PARAMS": {"usp":"pp_url", "entry.4701265": "???"},
    "EB_ORG_ID": "2640231310",
    "EB_API_TOKEN": "private-eventbright-api-token",
    "MOCK_DATE" : "2021-02-16",
    "SMTP_SERVER": "localhost",
    "SMTP_PORT": 1025,
    "SMTP_USER": "",
    "SMTP_PASS": "",
    "USE_TLS":   false,
    "USE_LOGIN": false
}
```

The contents of the email are read from a html file set by `EMAIL_TMPL_FILE` in the config. This
template needs a placeholder for the url. A secon placeholder, will be replaced with today's date.

For example:

```html
<p>Please remember to complete the <a href="%s">%s Health Survey</a> by 9 AM.
Submission of the survey is required for your participation today.</p>
```

could be renderedas:

```html
<p>Please remember to complete the <a
href="https://docs.google.com/forms/d/e/1FAIpQLSd...HtUhg/viewform?usp=pp_url&entry.4701265=Event+Name">Feb 21 Health Survey</a> by 9 AM.
Submission of the survey is required for your participation today.</p>
```

The same work for the subject.

You can test this using MailHog or https://mailtrap.io and setting the smtp server accordingly.


```

You can test this using MailHog or https://mailtrap.io and setting the smtp server accordingly.

