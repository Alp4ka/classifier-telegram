package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alp4ka/mlogger"
	"github.com/Alp4ka/mlogger/field"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	_defaultGracefulShutdownTimeout = 15 * time.Second
)

func main() {
	env := setup()
	go awaitGracefulShutdown(env.cancelFunc)

	mlogger.L().Info("Starting app")
	err := env.app.Run(env.ctx)
	mlogger.L().Info("App stopped!", field.Error(err))

	err = env.app.Close()
	if err != nil {
		mlogger.L().Error("Failed to close app", field.Error(err))
	}
	mlogger.L().Info("App closed successfully!")
}

func awaitGracefulShutdown(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(ch)

	select {
	case sig := <-ch:
		mlogger.L().Info("graceful shutdown: reason: received signal", field.Any("signal", sig.String()))
	}
	cancel()

	go func() {
		time.Sleep(_defaultGracefulShutdownTimeout)
		mlogger.L().Fatal("force quit, graceful shutdown timeout expired, force exit", field.String("timeout", _defaultGracefulShutdownTimeout.String()))
	}()
}
