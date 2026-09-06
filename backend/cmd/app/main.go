package main

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/infrastructure/api/controller"
	"backend/internal/infrastructure/oidcclient"
	"backend/internal/infrastructure/repository"
	"backend/internal/logger"
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	logger.Init(config.String("logging.level"))

	repo, err := repository.NewRepository()
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	if err := domain.EnsureBootstrapAdmin(repo); err != nil {
		zap.L().Fatal(err.Error())
	}

	var oidcClient *oidcclient.Client
	if strings.EqualFold(config.String("authentication.type"), "OIDC") {
		oidcClient, err = oidcclient.New(context.Background(), repo.OidcStateRepository)
		if err != nil {
			zap.L().Fatal(err.Error())
		}
	}

	svc := domain.NewService(repo, oidcClient)

	router, err := controller.Router(svc)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	err = router.Run(fmt.Sprintf(":%s", config.String("server.port")))
	if err != nil {
		zap.L().Fatal(err.Error())
	}
}
