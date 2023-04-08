package forwarder

import (
	"database/sql"
	"errors"
	stdlog "log"

	agent "github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener"
	"github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener/typeutils"
	"github.com/alexandria-oss/streams/codec"
	streamsql "github.com/alexandria-oss/streams/driver/sql"
	"github.com/alexandria-oss/streams/proxy/egress"
	"github.com/spf13/viper"
)

const (
	KafkaDriver = "kafka"
)

var ErrUnknownDriver = errors.New("forwarder: unknown driver")

func newDefaultConfig() {
	viper.SetDefault("forwarder.egress_table", egress.DefaultEgressTableName)
}

func newConfig(db *sql.DB) egress.ForwarderConfig {
	defCfg := egress.NewForwarderDefaultConfig()
	newDefaultConfig()
	return egress.ForwarderConfig{
		Storage: streamsql.NewEgressStorageWithConfig(db, egress.StorageConfig{
			TableName: viper.GetString("forwarder.egress_table"),
		}),
		Writer: nil,
		Codec:  codec.ProtocolBuffers{},
		Logger: stdlog.New(agent.DefaultLogger.With().Str("level", "info").
			Str("process", "forwarder").Logger(), "", 0),
		TableName:                 viper.GetString("forwarder.egress_table"),
		ForwardJobTimeout:         typeutils.Coalesce(viper.GetDuration("forwarder.job_timeout"), defCfg.ForwardJobTimeout),
		ForwardJobTotalRetries:    typeutils.Coalesce(viper.GetInt("forwarder.job_total_retries"), defCfg.ForwardJobTotalRetries),
		ForwardJobRetryBackoff:    typeutils.Coalesce(viper.GetDuration("forwarder.job_retry_backoff"), defCfg.ForwardJobRetryBackoff),
		ForwardJobRetryBackoffMax: typeutils.Coalesce(viper.GetDuration("forwarder.job_retry_backoff_max"), defCfg.ForwardJobRetryBackoffMax),
	}
}

func NewForwarder(driver string, db *sql.DB) (egress.Forwarder, func() error, error) {
	fwdCfg := newConfig(db)
	var cleanup func() error

	switch driver {
	case KafkaDriver:
		fwdCfg.Writer, cleanup = NewKafka()
	default:
		return egress.Forwarder{}, nil, ErrUnknownDriver
	}
	return egress.NewForwarder(fwdCfg), cleanup, nil
}
