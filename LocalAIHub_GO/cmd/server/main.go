package main

import (
	"log"
	"os"
	"time"

	"localaihub/localaihub_go/internal/app/bootstrap"
	"localaihub/localaihub_go/internal/pkg/logger"
)

func main() {
	logFileName := "logs/localaihub-" + time.Now().Format("2006-01-02") + ".log"
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("create log dir failed: %v", err)
	}

	if err := logger.InitWithFile(logFileName); err != nil {
		log.Fatalf("init logger failed: %v", err)
	}

	app, err := bootstrap.New()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("bootstrap failed")
	}
	defer func() {
		_ = app.DB.Close()
	}()

	logger.Log.Info().Str("address", app.Config.Server.Address()).Msg("server starting")
	if err := app.Server.ListenAndServe(); err != nil {
		logger.Log.Fatal().Err(err).Msg("server stopped")
	}
}
