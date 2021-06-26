package main

import (
	"os"
)

import (
	"DGUT-yqfkgo/cmd"
	"DGUT-yqfkgo/internal/log"
)

func main() {
	app := cmd.App()
	log.InitLogger()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Msgf("Run App Failed: %v", err)
	}
}
