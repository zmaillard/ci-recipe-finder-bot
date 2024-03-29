package main

import (
	"ci-recipe-finder-bot/config"
	"ci-recipe-finder-bot/index"
)

func main() {
	config.Init()
	err := index.RefreshIndex()
	if err != nil {
		panic(err)
	}
}
