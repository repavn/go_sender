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
	"strconv"
)

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

	var chatID int
	var err error
	if os.Getenv("GROUP_CHAT_ID") != "" {
		chatID, err = strconv.Atoi(os.Getenv("GROUP_CHAT_ID"))
		if err != nil {
			fmt.Println("error get from env (chat):", err)
			return
		}
	} else {
		return
	}

	type Body struct {
		ChatID int    `json:"chat_id"`
		Text   string `json:"text"`
	}

	body, err := json.Marshal(Body{chatID, text})
	r := bytes.NewReader(body)
	if err != nil {
		fmt.Println("error json.Marshal (chat):", err)
		return
	}

	// Send message to bot group
	resp, err := http.Post(fmt.Sprintf("%sbot%s/sendMessage", TeleAPIURL, TeleBotToken), "application/json", r)
	if err != nil {
		fmt.Println("error sendMessage: err, response", err, resp)
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
	http.ListenAndServe(":9999", nil)
}
