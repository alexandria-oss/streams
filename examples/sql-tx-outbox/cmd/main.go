package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"sample/domain"
	"sample/domain/payment"
	"sample/storage"
	"syscall"

	streamskafka "github.com/alexandria-oss/streams/driver/kafka"
	streamsql "github.com/alexandria-oss/streams/driver/sql"
	"github.com/alexandria-oss/streams/proxy/egress"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

func main() {
	sysChan := make(chan os.Signal, 2)
	signal.Notify(sysChan, os.Interrupt, syscall.SIGTERM)

	db, err := sql.Open("postgres", "user=postgres password=root sslmode=disable port=6432 host=localhost database=sample_database")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	kafkaWriter := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092"),
		Topic:                  "",
		Balancer:               kafka.CRC32Balancer{},
		MaxAttempts:            0,
		WriteBackoffMin:        0,
		WriteBackoffMax:        0,
		BatchSize:              0,
		BatchBytes:             0,
		BatchTimeout:           0,
		ReadTimeout:            0,
		WriteTimeout:           0,
		RequiredAcks:           0,
		Async:                  false,
		Completion:             nil,
		Compression:            kafka.Snappy,
		Logger:                 nil,
		ErrorLogger:            nil,
		Transport:              nil,
		AllowAutoTopicCreation: true,
	}
	fwdCfg := egress.NewForwarderDefaultConfig()
	fwdCfg.Writer = streamskafka.NewWriter(kafkaWriter)
	fwdCfg.Storage = streamsql.NewEgressStorage(db)
	fwd := egress.NewForwarder(fwdCfg)
	go fwd.Start()
	defer fwd.Shutdown()

	var notifier egress.Notifier = egress.EmbeddedNotifier{
		Forwarder: fwd,
	}

	repo := storage.PaymentSQL{}
	var cmdBus domain.SyncCommandBus = domain.NewMemoryCommandBus()
	_ = cmdBus.Register(payment.CreateCommand{},
		storage.WithEmbeddedEgressAgent(notifier, storage.WithSQLTransaction(db, payment.HandleCreateCommand)))
	egressWriter := streamsql.NewWriter(streamsql.WithEgressTable("streams_egress"))
	payment.DefaultService = payment.NewService(repo, egressWriter)

	err = cmdBus.Exec(context.TODO(), payment.CreateCommand{
		UserID: "aruiz",
		Amount: 99.99,
	})
	if err != nil {
		panic(err)
	}
	<-sysChan
}
