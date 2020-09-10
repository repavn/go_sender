/*

	telegram get updates

	{
		"ok": true,
		"result": [
			{
			"update_id": 44732148,
			"message": {
				"message_id": 2,
				"from": {
				"id": 543712595,
				"is_bot": false,
				"first_name": "Temitch",
				"last_name": "Loverman",
				"username": "temitch",
				"language_code": "ru"
				},
				"chat": {
				"id": -393115454,
				"title": "BotSendGroup",
				"type": "group",
				"all_members_are_administrators": true
				},
				"date": 1599661394,
				"new_chat_participant": {
				"id": 1244918083,
				"is_bot": true,
				"first_name": "go_sender",
				"username": "repavnnnnnnnnnnnnn_bot"
				},
				"new_chat_member": {
				"id": 1244918083,
				"is_bot": true,
				"first_name": "go_sender",
				"username": "repavnnnnnnnnnnnnn_bot"
				},
				"new_chat_members": [
				{
					"id": 1244918083,
					"is_bot": true,
					"first_name": "go_sender",
					"username": "repavnnnnnnnnnnnnn_bot"
				}
				]
			}
			}
		]
	}


	example client query:
	curl --location --request POST 'http://127.0.0.1:9999/' \
	--header 'Content-Type: application/json' \
	--data-raw '{"text": "Hello, test recipient 9!"}'

	This is a service of sending message to the messengers (telegram, email,)
	TODO:
	1. Take POST queries with message subject, file etc. (m.b https://golang.org/pkg/net/http/fcgi/, https://uwsgi-docs.readthedocs.io/en/latest/Go.html)
	2. After POST query - fast response, and asynchronously send message to email
	3. Authorize query
	4. Print via log package
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/smtp"
	"os"
	"path/filepath"
)

// Chat ...
type Chat struct {
	ID int `json:"id"`
}

// TeleMessage ...
type TeleMessage struct {
	Chat Chat `json:"chat"`
}

// TeleResult ...
type TeleResult struct {
	Message TeleMessage `json:"message"`
}

// Update ...
type Update struct {
	Ok     bool         `json:"ok"`
	Result []TeleResult `json:"result"`
}

const mail = "mail"
const telegram = "telegram"

// TeleAPIURL ...
var TeleAPIURL = "https://api.telegram.org/"

// TeleBotToken ...
var TeleBotToken = os.Getenv("SEND_BOT_TOKEN")

// LogRawHTTP prints post query body
func LogRawHTTP(request *http.Request, response http.ResponseWriter) {

	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		http.Error(response, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	fmt.Println(fmt.Sprintf("***LOG income http query, BEGIN: \n %s \n ***END\n", string(dump)))
}

// Message - is the base interface with method send()
type Message interface {
	send() string
}

// BaseMessage - represents email message with data for sending
type BaseMessage struct {
	Text string `json:"text"`
}

// Send - sends message to ...
func (message *BaseMessage) send() string {
	fmt.Println("Base sending")
	return ""
}

// EmailMessage - represents email message with data for sending
type EmailMessage struct {
	BaseMessage
	Subject string `json:"subject"`
	To      string `json:"to"`
}

// TelegramMessage ...
type TelegramMessage struct {
	BaseMessage
}

// sendTelegram - sends message to telegram bot group
func sendTelegram(text string) {
	url := fmt.Sprintf("%sbot%s/getUpdates", TeleAPIURL, TeleBotToken)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error getUpdates:", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	var update Update
	json.Unmarshal(body, &update)

	var ChatID int
	ChatID = update.Result[0].Message.Chat.ID
	type Body struct {
		ChatID int    `json:"chat_id"`
		Text   string `json:"text"`
	}

	body, err = json.Marshal(Body{ChatID, text})
	r := bytes.NewReader(body)

	if err != nil {
		fmt.Println("error json.Marshal (chat):", err)
		return
	}

	sendURL := fmt.Sprintf("%sbot%s/sendMessage", TeleAPIURL, TeleBotToken)
	// Send message to bot group
	resp, err = http.Post(sendURL, "application/json", r)
	if err != nil {
		fmt.Println("error sendMessage:", err)
		return
	}
}

// Send - extracts send options and send to email, return message text
func (message *TelegramMessage) send() string {
	go sendTelegram(message.Text)
	return message.Text
}

// SendMail - sends message to email address
func SendMail(text string, subject string, from string, to []string, pass string, smtpHost string, smtpPort string) {
	auth := smtp.PlainAuth("", from, pass, smtpHost)
	msg := []byte(text)
	if subject != "" {
		msg = []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, text))
	}
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		fmt.Println("Error of sending to email:", err)
		return
	}
	fmt.Println(fmt.Sprintf("OK - message '%s' is sent from %s to %s ", text, from, to))
}

// Send - extracts send options and send to email, return message text
func (message *EmailMessage) send() string {
	from := os.Getenv("FROM_MAIL")
	to := []string{message.To}
	pass := os.Getenv("MAIL_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	go SendMail(message.Text, message.Subject, from, to, pass, smtpHost, smtpPort)
	return message.Text
}

// Index - single view, dispatch sending to messangers
func index(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		res.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		LogRawHTTP(req, res)
		var message Message
		var text string

		switch messanger := filepath.Base(req.URL.Path); messanger {
		case mail:
			message = &EmailMessage{}
		case telegram:
			message = &TelegramMessage{}
		default:
			message = &BaseMessage{}
		}

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return
		}
		err = json.Unmarshal(body, &message)
		if err != nil {
			return
		}

		// send message
		text = message.send()

		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "text/html")
		io.WriteString(res, fmt.Sprintf("Your message text: '%s' has been sent", text))
	}
}

func main() {
	fmt.Println("service of sending is runnning")
	http.HandleFunc("/send/mail", index)
	http.HandleFunc("/send/telegram", index)
	http.ListenAndServe("127.0.0.1:9999", nil)
}
