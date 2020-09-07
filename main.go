/*
	This is a service of sending message to the messengers (telegramm, email,)
	TODO:
	1. Take POST queries with message body, head, file etc. (m.b https://golang.org/pkg/net/http/fcgi/, https://uwsgi-docs.readthedocs.io/en/latest/Go.html)
	2. After POST query - fast response, and asynchronously send message to email
	3. Authorize query
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func get(res http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" && req.Method != "POST" {
		res.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		switch method := req.Method; method {
		case "GET":
			res.WriteHeader(http.StatusOK)
			res.Header().Set("Content-Type", "text/html")
			io.WriteString(res, "Hello. Use the POST query for send message.")
		case "POST":
			var result map[string]string
			decoder := json.NewDecoder(req.Body)
			err := decoder.Decode(&result)

			if err != nil {
				http.Error(res, "BAD REQUEST. Detail info:"+err.Error(), http.StatusBadRequest)
				return
			}

			fmt.Println("\n message body=", result["body"])

			res.WriteHeader(http.StatusCreated)
			res.Header().Set("Content-Type", "text/html")
			io.WriteString(res, "Your message has been sent")
		}
	}
}

func main() {
	fmt.Println("service of sending is runnning")
	http.HandleFunc("/", get)
	http.ListenAndServe("127.0.0.1:9999", nil)
}
