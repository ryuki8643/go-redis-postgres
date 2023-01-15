FROM golang:1.19.2-alpine3.16

ENV ROOT=/go/src/app
WORKDIR ${ROOT}

RUN apk update && apk add git

COPY . ${ROOT}

RUN go mod download
RUN go build -o ./app main.go


EXPOSE 8080


CMD ["./app"]