package index

import (
	"bytes"
	"ci-recipe-finder-bot/config"
	"encoding/json"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var bufferSize = 100

type recipeModel struct {
	RecipeId   int
	Issue      int
	StartMonth string
	EndMonth   string
	Year       int
	MainTitle  *string
	CoverTitle *string
	Category   string
	Page       int
	Notes      *string
}

func (m recipeModel) ToIndex() recipeIndexRequest {
	r := recipeIndexRequest{
		Action: "mergeOrUpload",
	}

	r.RecipeId = strconv.Itoa(m.RecipeId)
	r.Issue = m.Issue
	r.Months = strings.Trim(m.StartMonth, " ") + " - " + strings.Trim(m.EndMonth, " ")
	r.Year = m.Year
	r.MainTitle = m.MainTitle
	r.CoverTitle = m.CoverTitle
	r.Categories = []string{m.Category}
	r.Page = m.Page
	r.Notes = m.Notes

	return r
}

type RecipeIndex struct {
	RecipeId   string   `json:"recipeId"`
	Issue      int      `json:"issue"`
	Months     string   `json:"months"`
	Year       int      `json:"year"`
	MainTitle  *string  `json:"mainTitle"`
	CoverTitle *string  `json:"coverTitle"`
	Categories []string `json:"categories"`
	Page       int      `json:"page"`
	Notes      *string  `json:"notes"`
}

func (r RecipeIndex) Title() string {
	if r.MainTitle != nil {
		return *r.MainTitle
	}

	return *r.CoverTitle
}

func (r RecipeIndex) String() string {
	return fmt.Sprintf("%s %s, %v", r.Title(), r.Months, r.Year)
}

type recipeIndexRequest struct {
	RecipeIndex
	Action string `json:"@search.action"`
}

type RecipePost struct {
	Value []recipeIndexRequest `json:"value"`
}

func saveIndex(body *RecipePost) error {
	config := config.GetConfig()
	apiVersion := "2020-06-30"

	searchUrl := fmt.Sprintf("https://%s.search.windows.net/indexes/%s/docs/index?api-version=%s", config.SearchService, config.SearchIndex, apiVersion)

	content, err := json.Marshal(body)
	if err != nil {
		log.WithFields(log.Fields{
			"searchUrl": searchUrl,
			"error":     err,
		}).Warn("Failed to generate index post")
		return err
	}
	client := &http.Client{}

	req, err := http.NewRequest("POST", searchUrl, bytes.NewBuffer(content))
	if err != nil {
		log.WithFields(log.Fields{
			"searchUrl": searchUrl,
			"error":     err,
		}).Warn("Error creating index")
		return err
	}

	req.Header.Add("api-key", config.SearchApiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"searchUrl": searchUrl,
			"error":     err,
		}).Warn("Error creating index")
		return err
	} else if resp.StatusCode >= 300 {
		c := string(content)
		fmt.Println(c)
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		s := buf.String()

		if err != nil {
			log.WithFields(log.Fields{
				"searchUrl":  searchUrl,
				"error":      err,
				"statusCode": resp.StatusCode,
			}).Warn("Cannot Parse Error Response")
			return err
		}

		log.WithFields(log.Fields{
			"searchUrl":  searchUrl,
			"response":   s,
			"statusCode": resp.StatusCode,
		}).Warn("Error creating index")

		return err
	}

	return nil
}

func buildRecipeQuery() (string, []interface{}) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	mainQuery := psql.Select(
		`r."recipeId" as "recipeId"`,
		`i."number" as "issue"`,
		`i."startMonth" as "startMonth"`,
		`i."endMonth" as "endMonth"`,
		`i."year" as "year"`,
		`r."titleFrontCover" as "coverTitle"`,
		`r."titleBackCover" as "mainTitle"`,
		`r."page" as "page"`,
		`r."notes" as "notes"`,
		`c."categoryName" as "category"`)

	mainQuery = mainQuery.From(`recipe r`).
		Join(`issue i on "r"."issueId" = i."issueId"`).
		Join(`category c on "r"."categoryId" = "c"."categoryId"`)

	a, b, _ := mainQuery.ToSql()
	return a, b
}

func RefreshIndex() error {
	conn := config.GetDB()
	query, args := buildRecipeQuery()
	rows, err := conn.Queryx(query, args...)
	var recordsToIndex []recipeIndexRequest

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		model := recipeModel{}
		err = rows.Scan(&model.RecipeId, &model.Issue, &model.StartMonth, &model.EndMonth, &model.Year, &model.CoverTitle, &model.MainTitle, &model.Page, &model.Notes, &model.Category)
		if err != nil {
			log.Fatal(err)
			return err
		}

		indexRec := model.ToIndex()

		recordsToIndex = append(recordsToIndex, indexRec)
	}

	for i := 0; i < len(recordsToIndex); i += bufferSize {
		j := i + bufferSize
		if j > len(recordsToIndex) {
			j = len(recordsToIndex)
		}

		err = saveIndex(&RecipePost{
			Value: recordsToIndex[i:j],
		})
		if err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
	}

	return nil
}
