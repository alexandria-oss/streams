package main

import (
	"context"
	"database/sql"
	"sample/domain"
	"sample/domain/payment"
	"sample/storage"

	"github.com/alexandria-oss/streams"
	streamsql "github.com/alexandria-oss/streams/driver/sql"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "user=postgres password=root sslmode=disable port=6432 host=localhost database=sample_database")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	repo := storage.PaymentSQL{}
	w := streamsql.
		NewWriter(
			streamsql.WithEgressTable("streams_egress"),
		).
		WithParentConfig(
			streams.WithIdentifierFactory(streams.NewUUID),
		)
	var cmdBus domain.SyncCommandBus = domain.NewMemoryCommandBus()
	_ = cmdBus.Register(payment.CreateCommand{}, storage.WithSQLTransaction(db, payment.HandleCreateCommand))
	payment.DefaultService = payment.NewService(repo, w)

	err = cmdBus.Exec(context.TODO(), payment.CreateCommand{
		UserID: "aruiz",
		Amount: 99.99,
	})
	if err != nil {
		panic(err)
	}
}
