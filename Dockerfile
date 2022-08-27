FROM golang:1.17 as build

WORKDIR /go/src/recipebot

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .

FROM alpine:latest

WORKDIR /app

RUN mkdir ./static
COPY ./static ./static

COPY --from=build /go/src/recipebot/app .

EXPOSE 3000

CMD ["./app"]