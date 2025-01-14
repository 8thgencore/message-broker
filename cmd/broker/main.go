package main

import (
	"context"
	"log"

	"github.com/8thgencore/message-broker/internal/app"
)

func main() {
	ctx := context.Background()

	a, err := app.NewApp(ctx)
	if err != nil {
		log.Fatal("failed to init app: ", err)
	}

	err = a.Run()
	if err != nil {
		log.Fatal("failed to run app: ", err)
	}
}
