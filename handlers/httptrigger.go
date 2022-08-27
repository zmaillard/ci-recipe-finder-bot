package handlers

import (
	"ci-recipe-finder-bot/config"
	"ci-recipe-finder-bot/index"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type searchResult struct {
	index.RecipeIndex
	Score float64 `json:"@search.score"`
}

func HelpHandler(c *fiber.Ctx) error {
	cfg := config.GetConfig()

	return c.Render("layout", fiber.Map{
		"PublicUrl":   cfg.PublicUrl,
		"PhoneNumber": cfg.PhoneNumber,
	})

}

func ReceiveSMSHandler(c *fiber.Ctx) error {
	cfg := config.GetConfig()

	queryVals, err := url.ParseQuery(string(c.Body()))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("invalid payload")
		return c.SendStatus(http.StatusBadRequest)
	}

	searchTerm := url.QueryEscape(queryVals.Get("Body"))

	client := twilio.NewRestClient()
	params := &openapi.CreateMessageParams{}
	params.SetTo(queryVals.Get("From"))
	params.SetFrom(queryVals.Get("To"))

	if strings.ToLower(searchTerm) == url.QueryEscape("show help") {
		params.SetBody(cfg.HelpPage)
	} else if strings.ToLower(searchTerm) == "web" {
		params.SetBody(cfg.SearchUIBaseUrl)
	} else {

		searchUrl := fmt.Sprintf("https://%s.search.windows.net/indexes/%s/docs?api-version=2019-05-06&api-key=%s&search=%s", cfg.SearchService, cfg.SearchIndex, cfg.SearchApiKey, searchTerm)
		log.WithFields(log.Fields{
			"url": searchUrl,
		}).Warn("Building Url")
		resp, err := http.Get(searchUrl)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warn("Error Getting Response")
			return c.SendStatus(http.StatusBadRequest)
		}
		defer resp.Body.Close()

		// Read body from response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warn("Error Reading Response")
			return c.SendStatus(http.StatusBadRequest)
		}

		log.WithFields(log.Fields{
			"body": string(body),
		}).Warn("Results From Search Service")

		searchRes := struct {
			Value []searchResult `json:"value"`
		}{}

		err = json.Unmarshal(body, &searchRes)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"body":  string(body),
			}).Warn("Error Parsing Response")
			return c.SendStatus(http.StatusBadRequest)
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
			var outputArr []string
			for _, v := range items {
				outputArr = append(outputArr, v.String())
			}

			output = strings.Join(outputArr, "\n")

		}

		if searchCount > 5 {
			u, _ := url.Parse(cfg.SearchUIBaseUrl)
			u.Path = path.Join(u.Path, "search")
			output = output + "\nView More Results Here: " + u.String() + "?q=" + searchTerm
		}

		params.SetBody(output)
	}
	_, err = client.ApiV2010.CreateMessage(params)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Could Not Send Message")
		return c.SendStatus(http.StatusBadRequest)
	} else {
		log.Debug("Success")
		return c.SendStatus(http.StatusOK)
	}

}
