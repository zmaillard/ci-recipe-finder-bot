package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"os"
)

type TwilioPayload struct {
	MessageSid string
	SmsSid string
	AccountSid string
	MessagingServiceSid string
	From string
	To string
	Body string
	NumMedia int
}


func ReceiveSMSHandler(w http.ResponseWriter, r *http.Request) {
	// set the response header as JSON
	//w.Header().Set("Content-Type", "application/json")

	// read request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid payload")
		return
	}

	var twilioPayload TwilioPayload
	err = json.Unmarshal(reqBody, &twilioPayload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid payload")
		return
	}

	client := twilio.NewRestClient()
	params := &openapi.CreateMessageParams{}
	params.SetTo(os.Getenv(twilioPayload.From))
	params.SetFrom(os.Getenv(twilioPayload.To))
	params.SetBody("Hello from Golang!")

	_, err = client.ApiV2010.CreateMessage(params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}

}