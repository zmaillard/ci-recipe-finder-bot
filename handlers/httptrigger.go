package handlers

import (
	"ci-recipe-finder-bot/config"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)
/*
{
    "@odata.context": "https://signsearchtest.search.windows.net/indexes('recipe-index')/$metadata#docs(*)",
    "value": [
        {
            "@search.score": 7.641527,
            "id": "13439125386506261573",
            "issue": "107",
            "months": [
                "November",
                "December"
            ],
            "year": 2010,
            "recipe": "Classic Pot Roast",
            "article": "Really Good Pot Roast",
            "category": "Main Dish",
            "page": 11,
            "notes": ""
        },
 */

type SearchResult struct {
	Score float64 `json:"@search.score"`
	Id string`json:"id"`
	Issue string `json:"issue"`
	Months []string `json:"months"`
	Year int`json:"year"`
	Recipe string`json:"recipe"`
	Article string`json:"article"`
	Category string`json:"category"`
	Page int`json:"page"`
	Notes string`json:"notes"`
}

func (r SearchResult) String() string {
	return fmt.Sprintf("%s %s, %v", r.Recipe, r.FormatMonths(), r.Year)
}

func (r SearchResult) FormatMonths() string {
	return strings.Join(r.Months, "-")
}

func ReceiveSMSHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
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

	url:= fmt.Sprintf("https://%s.search.windows.net/indexes/%s/docs?api-version=2019-05-06&api-key=%s&search=%s", cfg.SearchService, cfg.SearchIndex, cfg.SearchApiKey, queryVals.Get("Body"))
	resp, err := http.Get(url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Error Getting Response")
		return
	}
	defer resp.Body.Close()

	// Read body from response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithFields(log.Fields{
			"error": err,,
		}).Warn("Error Reading Response")
		return
	}

	searchRes := struct {
		Value []SearchResult
	} { }

	err = json.Unmarshal(body, &searchRes)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithFields(log.Fields{
			"error": err,
			"body": string(body),
		}).Warn("Error Parsing Response")
		return
	}

	var output string
	searchCount := len(searchRes.Value)
	if searchCount == 0 {
		output = "No Recipes Found"
	} else {
		idx := searchCount
		if idx > 5 {
			idx = 5
		}

		items := searchRes.Value[0:idx]
		var outputArr  []string
		for _, v := range items {
			outputArr = append(outputArr, v.String())
		}

		output = strings.Join(outputArr, "\n")
	}


	client := twilio.NewRestClient()
	params := &openapi.CreateMessageParams{}
	params.SetTo(queryVals.Get("From"))
	params.SetFrom(queryVals.Get("To"))
	params.SetBody(output)

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