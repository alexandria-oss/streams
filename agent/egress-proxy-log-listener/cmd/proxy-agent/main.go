package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	agent "github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener"
	"github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener/forwarder"
	"github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener/listener"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("streams")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	db, err := sql.Open("pgx", viper.GetString("postgres.connection_string"))
	if err != nil {
		agent.DefaultLogger.Fatal().Err(err).Str("process", "database").Msg("fatal failure detected, stopping agent")
	}
	defer func() {
		if errClose := db.Close(); errClose != nil {
			agent.DefaultLogger.Err(err).Msg("failure during database client shutdown")
		}
	}()

	writerDriver := viper.GetString("agent.writer.driver")
	fwd, cleanFwd, err := forwarder.NewForwarder(writerDriver, db)
	if err != nil {
		agent.DefaultLogger.Fatal().Err(err).Str("process", "forwarder").Msg("fatal failure detected, stopping agent")
	}
	defer func() {
		if errClean := cleanFwd(); errClean != nil {
			agent.DefaultLogger.Err(err).Str("process", "forwarder").Msg("failure detected during cleanup")
		}
	}()
	agent.DefaultLogger.Info().
		Str("writer_driver", writerDriver).
		Msg("driver detected")

	listenDriver := viper.GetString("agent.listener.driver")
	l, err := listener.NewListener(listenDriver, fwd)
	if err != nil {
		agent.DefaultLogger.Fatal().Err(err).Str("process", "log_listener").Msg("fatal failure detected, stopping agent")
	}

	agent.DefaultLogger.Info().
		Str("listener_driver", listenDriver).
		Msg("driver detected")

	sysChan := make(chan os.Signal, 4)
	signal.Notify(sysChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	go func() {
		if err = fwd.Start(); err != nil {
			agent.DefaultLogger.Err(err).Str("process", "forwarder").Msg("fatal failure detected, stopping agent")
			sysChan <- os.Interrupt
		}
	}()
	go func() {
		if err = l.Start(); err != nil {
			agent.DefaultLogger.Err(err).Str("process", "log_listener").Msg("fatal failure detected, stopping agent")
			sysChan <- os.Interrupt
		}
	}()
	<-sysChan

	fwd.Shutdown()
	shutdownListener(l)

	agent.DefaultLogger.Info().Msg("agent shutdown succeeded")
}

func shutdownListener(l listener.Listener) {
	shutdownListenerCtx, listenerCancel := context.WithTimeout(context.Background(), time.Second*15)
	defer listenerCancel()
	if err := l.Close(shutdownListenerCtx); err != nil {
		agent.DefaultLogger.Err(err).Msg("listener graceful shutdown process failed")
	}
}
