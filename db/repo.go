package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"strings"
)

type Recipe struct {
	Issue
	RecipeId        int
	FrontCoverTitle *string
	BackCoverTitle  *string
	Category        string
	Page            int
	Notes           *string
}

type Issue struct {
	IssueNumber int
	StartMonth  string
	EndMonth    string
	Year        int
}

func (i Issue) Title() string {
	return fmt.Sprintf("%s and %s %v", i.StartMonth, i.EndMonth, i.Year)
}

type Repository interface {
	GetIssueNumber(issueNumber int, startMonth string, endMonth string, year int) (int, error)
	GetCategory(categoryName string) (int, error)
	AddRecord(issueId int, categoryId int, page int, backCover string, frontCover string) error
	AllRecipes() ([]Recipe, error)
	GetLatestIssue() (Issue, error)
}

type db struct {
	Tx pgx.Tx
}

func (d db) GetLatestIssue() (Issue, error) {
	var issue Issue
	rows, err := d.Tx.Query(context.Background(), `select "number", "startMonth", "endMonth", "year" 
	from issue 
	order by number desc 
	limit 1`)

	defer rows.Close()
	if err != nil {
		return issue, err
	}

	if rows.Next() {
		err = rows.Scan(&issue.IssueNumber, &issue.StartMonth, &issue.EndMonth, &issue.Year)
		return issue, err
	}

	return issue, fmt.Errorf("no issue found")
}

func (d db) AllRecipes() ([]Recipe, error) {
	var recipes []Recipe
	rows, err := d.Tx.Query(context.Background(),
		`SELECT r."recipeId",r."titleFrontCover",
			 r."titleBackCover",r."page",r."notes",
			 i."number", i."startMonth", i."endMonth", 
			i."year", c."categoryName"
			FROM recipe r INNER JOIN issue i ON r."issueId" = i."issueId"
			INNER JOIN category c ON r."categoryId" = c."categoryId"`)
	defer rows.Close()
	if err != nil {
		return recipes, err
	}
	for rows.Next() {
		var recipe Recipe
		err = rows.Scan(&recipe.RecipeId, &recipe.FrontCoverTitle, &recipe.BackCoverTitle, &recipe.Page, &recipe.Notes, &recipe.IssueNumber, &recipe.StartMonth, &recipe.EndMonth, &recipe.Year, &recipe.Category)
		if err != nil {
			return recipes, err
		}
		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

func (d db) AddRecord(issueId int, categoryId int, page int, backCover string, frontCover string) error {
	_, err := d.Tx.Exec(context.Background(), `INSERT INTO recipe ("titleFrontCover", "titleBackCover", page, "categoryId", "issueId") VALUES ($1,$2,$3,$4,$5)`, frontCover, backCover, page, categoryId, issueId)
	if err != nil {
		return err
	}

	return nil
}

func (d db) GetIssueNumber(issueNumber int, startMonth string, endMonth string, year int) (int, error) {
	var issueId int
	rows, err := d.Tx.Query(context.Background(), `SELECT "issueId" FROM issue WHERE number = $1`, issueNumber)
	defer rows.Close()
	if err != nil {
		return issueId, err
	}
	if rows.Next() {
		err = rows.Scan(&issueId)
		if err != nil {
			return issueId, err
		}

	} else {
		rows, err := d.Tx.Query(context.Background(), `INSERT INTO issue (number, "startMonth", "endMonth", year) VALUES ($1,$2,$3,$4) RETURNING "issueId"`, issueNumber, startMonth, endMonth, year)
		defer rows.Close()
		if err != nil {
			return issueId, err
		}

		if rows.Next() {
			rows.Scan(&issueId)
		} else {
			return issueId, fmt.Errorf("no Records Found")
		}

	}

	return issueId, nil
}

func (d db) GetCategory(categoryName string) (int, error) {
	var categoryId int
	rows, err := d.Tx.Query(context.Background(), `SELECT "categoryId" FROM category WHERE lower("categoryName") = $1`, strings.ToLower(categoryName))
	defer rows.Close()
	if err != nil {
		return categoryId, err
	}
	if rows.Next() {
		err = rows.Scan(&categoryId)
		if err != nil {
			return categoryId, err
		}

	} else {
		rows, err := d.Tx.Query(context.Background(), `INSERT INTO category ("categoryName") VALUES ($1) RETURNING "categoryId"`, strings.ToTitle(categoryName))
		defer rows.Close()
		if err != nil {
			return categoryId, err
		}

		if rows.Next() {
			rows.Scan(&categoryId)
		} else {
			return categoryId, fmt.Errorf("no Records Found")
		}

	}
	return categoryId, nil
}

func CreateRepository(tx pgx.Tx) Repository {
	return db{Tx: tx}
}
