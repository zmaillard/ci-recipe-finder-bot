PROJECT_NAME ?= ci-recipe-finder-bot
ENV ?= dev

buildfn:
	export GO111MODULE=on
	 GOOS=windows GOARCH=amd64 go build -o ./azure-function/main main.go

build:
	export GO111MODULE=on
	go build -o ./azure-function/main main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

default: build

.PHONY: build clean buildfn lambda default
