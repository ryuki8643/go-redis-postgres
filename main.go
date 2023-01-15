package main

import (
	"github.com/geolocket/batch_redis/internal/gin_server"
)

func main() {
	gin_server.NewHTTPServer()
}
