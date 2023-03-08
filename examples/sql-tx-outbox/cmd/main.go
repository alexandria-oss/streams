package main

import (
	"context"
	"database/sql"
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
	w := streamsql.NewWriter(streamsql.Config{
		IdentifierFactory: streams.NewUUID,
		WriterEgressTable: "streams_egress",
	})
	var svc payment.Service
	svc = payment.NewService(repo, w)
	svc = payment.NewServiceTransactionContext(db, svc)
	if _, err = svc.Create(context.TODO(), "aruiz", 99.99); err != nil {
		panic(err)
	}
}
