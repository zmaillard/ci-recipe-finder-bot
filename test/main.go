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


/*
http://localhost:8000
http://localhost:5500
http://localhost:3000
https://ci.sagebrushgis.com
 */