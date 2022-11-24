package main

import (
	"time"

	"git.internal.com/wingspan/pkg"
)

func main() {
	server := pkg.NewServer()
	server.Register("Queue", pkg.NewQueue(2))
	server.Register("Matchmaker", pkg.NewMatchmaker(15*time.Second))
	server.Register("Game", pkg.NewGameManager())

	print("Listening on 0.0.0.0:8080\n")
	server.Listen("0.0.0.0:8080")
}
