package notify

import (
	"ci-recipe-finder-bot/config"
	"ci-recipe-finder-bot/db"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
)

func buildBodyString(i db.Issue) string {
	return fmt.Sprintf("Greetings From The Recipe Bot!  The %s Issue Of Cooks Illustrated is now ready to search!", i.Title())
}

func Send() error {
	config.Init()

	cfg := config.GetConfig()

	conn, err := pgx.Connect(context.Background(), cfg.BuildDatabaseUrl())
	defer conn.Close(context.Background())

	tx, err := conn.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}

	repo := db.CreateRepository(tx)
	issue, err := repo.GetLatestIssue()

	client := &http.Client{}

	for _, toNumber := range cfg.GetTwilioToNumbers() {
		vals := url.Values{}
		vals.Set("Body", buildBodyString(issue))
		vals.Set("From", cfg.TwilioFromNumber)
		vals.Set("To", toNumber)
		req, err := http.NewRequest("POST", cfg.BuildTwilioUrl(), strings.NewReader(vals.Encode()))
		if err != nil {
			return err
		}
		req.SetBasicAuth(cfg.TwilioUser, cfg.TwilioApiKey)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusCreated {
			log.Warn(fmt.Printf("Message Returned Status Code %v", resp.StatusCode))
			return fmt.Errorf("an Error Ocurred Creating Message")
		}
	}

	return nil

}
