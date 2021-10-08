package handlers

import (
	log "github.com/sirupsen/logrus"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"io/ioutil"
	"net/http"
	"net/url"
)


func ReceiveSMSHandler(w http.ResponseWriter, r *http.Request) {
	// set the response header as JSON
	//w.Header().Set("Content-Type", "application/json")
	log.Debug("In Handler")
	// read request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("invalid payload")
		return
	}

	queryVals, err := url.ParseQuery(string(reqBody))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("invalid payload")
		return
	}

	client := twilio.NewRestClient()
	params := &openapi.CreateMessageParams{}
	params.SetTo(queryVals.Get("From"))
	params.SetFrom(queryVals.Get("To"))
	params.SetBody("Hello " + queryVals.Get("Body"))

	_, err = client.ApiV2010.CreateMessage(params)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Could Not Send Message")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		log.Debug("Success")
		w.WriteHeader(http.StatusOK)
	}

}