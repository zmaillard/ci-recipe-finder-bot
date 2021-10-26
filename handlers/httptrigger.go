package handlers

import (
	"ci-recipe-finder-bot/config"
	"ci-recipe-finder-bot/index"
	"embed"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

//go:embed layout.html.tmpl
var templateFS embed.FS



type searchResult struct {
	index.RecipeIndex
	Score float64 `json:"@search.score"`
}

type HelpTemplateModel struct {
	PhoneNumber string
	PublicUrl string
}

func HelpHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	tmpl := template.Must(template.ParseFS(templateFS, "layout.html.tmpl"))


	err := tmpl.Execute(w, HelpTemplateModel{
		PublicUrl: cfg.PublicUrl,
		PhoneNumber: cfg.PhoneNumber,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("invalid Template")
		return
	}

}

func ReceiveSMSHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()

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

	searchTerm := url.QueryEscape( queryVals.Get("Body"))
	searchUrl:= fmt.Sprintf("https://%s.search.windows.net/indexes/%s/docs?api-version=2019-05-06&api-key=%s&search=%s", cfg.SearchService, cfg.SearchIndex, cfg.SearchApiKey, searchTerm)
	log.WithFields(log.Fields{
		"url": searchUrl,
	}).Warn("Building Url")
	resp, err := http.Get(searchUrl)
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
			"error": err,
		}).Warn("Error Reading Response")
		return
	}

	log.WithFields(log.Fields{
		"body": string(body),
	}).Warn("Results From Search Service")

	searchRes := struct {
		Value []searchResult `json:"value"`
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

	if searchCount > 5 {
		u, _ := url.Parse(cfg.SearchUIBaseUrl)
		u.Path = path.Join(u.Path, "search")
		output = output +  "\nView More Results Here: " + u.String() + "?q=" + searchTerm
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