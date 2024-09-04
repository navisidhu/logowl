package main

import (
	"github.com/navisidhu/logowl/internal/server"
)

func main() {
	server := server.CreateInstance()

	server.Start()
}
