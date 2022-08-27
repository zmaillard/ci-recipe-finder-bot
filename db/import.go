package db

import (
	"ci-recipe-finder-bot/config"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
	"os"
	"strconv"
	"strings"
)

func ImportFile(filePath string, issueNumber int) error {
	config.Init()
	records, err := readCsv(filePath)

	if err != nil {
		log.Fatalln(err)
	}

	cfg := config.GetConfig()

	url := cfg.BuildDatabaseUrl()
	conn, err := pgx.Connect(context.Background(), url)
	defer conn.Close(context.Background())
	startMonth, endMonth, err := calculateMonthRange(issueNumber)
	year := calculateYear(issueNumber)
	tx, err := conn.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}

	repo := CreateRepository(tx)
	issueId, err := repo.GetIssueNumber(issueNumber, startMonth, endMonth, year)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	insertCount := 0

	for _, record := range records {

		category := strings.TrimSpace(record[0])
		categoryId, err := repo.GetCategory(category)
		if err != nil {
			tx.Rollback(context.Background())
			return err
		}

		backCover := strings.TrimSpace(record[1])
		page := strings.TrimSpace(record[2])
		frontCover := strings.TrimSpace(record[3])

		iPage, err := strconv.Atoi(page)
		if err != nil {
			tx.Rollback(context.Background())
			return err
		}

		err = repo.AddRecord(issueId, categoryId, iPage, backCover, frontCover)
		if err != nil {
			tx.Rollback(context.Background())
			return err
		}

		insertCount++
	}

	tx.Commit(context.Background())
	fmt.Println("Inserted %i Records", insertCount)
	return nil
}

func calculateYear(issueNumber int) int {
	return (issueNumber / 6) + 1993
}

func calculateMonthRange(issueNumber int) (string, string, error) {
	switch issueNumber % 6 {
	case 0:
		return "January", "February", nil
	case 1:
		return "March", "April", nil
	case 2:
		return "May", "June", nil
	case 3:
		return "July", "August", nil
	case 4:
		return "September", "October", nil
	case 5:
		return "November", "December", nil
	}

	return "", "", fmt.Errorf("invalid Issue Number")
}

func readCsv(fileName string) ([][]string, error) {
	f, err := os.Open(fileName)

	if err != nil {
		return [][]string{}, err
	}

	defer f.Close()

	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}
