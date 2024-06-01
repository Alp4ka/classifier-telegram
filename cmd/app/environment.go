package main

import (
	"context"
	"log"

	"github.com/Alp4ka/classifier-telegram/internal/app"
	"github.com/Alp4ka/classifier-telegram/internal/manager"

	"github.com/Alp4ka/classifier-telegram/internal/interactions/core"
	coregrpc "github.com/Alp4ka/classifier-telegram/internal/interactions/core/grpc"

	classifieraas "github.com/Alp4ka/classifier-telegram"
	"github.com/Alp4ka/classifier-telegram/internal/config"
	"github.com/Alp4ka/mlogger"
	"github.com/Alp4ka/mlogger/field"
	"github.com/Alp4ka/mlogger/misc"
)

type environment struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	cfg        *config.Config
	coreClient core.Client

	app *app.App
}

func setup() *environment {
	var env environment
	setupContext(&env)
	setupLogging(&env)
	setupConfig(&env)
	setupCoreClient(&env)
	setupApp(&env)
	return &env
}

func setupContext(env *environment) {
	// Env setup.
	env.ctx, env.cancelFunc = context.WithCancel(context.Background())
}

func setupLogging(env *environment) {
	env.ctx = field.WithContextFields(env.ctx, field.String("appName", classifieraas.AppName))
	logger, err := mlogger.NewProduction(
		env.ctx,
		mlogger.Config{
			Level: misc.LevelDebug,
		},
	)
	if err != nil {
		log.Fatal("Could not create logger!", err)
	}
	mlogger.ReplaceGlobals(logger)
}

func setupConfig(env *environment) {
	cfg, err := config.FromEnv()
	if err != nil {
		mlogger.L().Fatal("Failed to load config", field.Error(err))
	}

	// Env setup.
	env.cfg = cfg
}

func setupCoreClient(env *environment) {
	client, err := coregrpc.NewClient(coregrpc.Config{GRPCAddr: env.cfg.CoreGRPCAddress})
	if err != nil {
		mlogger.L().Fatal("Failed to connect to core gRPC server!", field.Error(err))
	}

	// Env setup.
	env.coreClient = manager.NewManager(client)
}

func setupApp(env *environment) {
	a, err := app.New(
		app.Config{
			APIKey:     env.cfg.APIKey,
			CoreClient: env.coreClient,
		},
	)
	if err != nil {
		mlogger.L().Fatal("Failed to setup app!", field.Error(err))
	}

	// Env setup.
	env.app = a
}
