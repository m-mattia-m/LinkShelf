package main

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/controller"
	"backend/internal/infrastructure/repository"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	loadLogger()

	repo, err := repository.NewRepository()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	svc := domain.NewService(repo)

	router, err := controller.Router(svc)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = router.Run(fmt.Sprintf(":%s", viper.GetString("server.port")))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func loadLogger() {

	fmt.Println(viper.GetString("database.name"))
	fmt.Println(viper.GetString("logging.level"))
	fmt.Println("---")
	level := config.ParseLevel(viper.GetString("logging.level"))

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)

	slog.Info("Logger initialized", slog.String("level", level.String()))
}
