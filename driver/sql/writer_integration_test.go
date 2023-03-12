//go:build integration

package sql_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alexandria-oss/streams"
	streamsql "github.com/alexandria-oss/streams/driver/sql"
	"github.com/alexandria-oss/streams/persistence"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type writerSuite struct {
	suite.Suite

	writer        streams.Writer
	db            *sql.DB
	tableFlushMap sync.Map
}

func TestWriter_Integration(t *testing.T) {
	suite.Run(t, &writerSuite{})
}

func (s *writerSuite) SetupSuite() {
	s.writer = streamsql.NewWriter()
	db, err := sql.Open("postgres", "user=postgres password=root sslmode=disable port=6432 host=localhost database=sample_database")
	s.Require().NoError(err)
	s.db = db
}

func (s *writerSuite) TearDownSuite() {
	defer s.db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	s.tableFlushMap.Range(func(key, value any) bool {
		table := key.(string)
		_, _ = s.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table))
		return true
	})
}

func (s *writerSuite) TearDownTest() {
}

func (s *writerSuite) TestWrite() {
	conn, err := s.db.Conn(context.TODO())
	s.Require().NoError(err)
	defer conn.Close()

	requestContext := context.TODO()

	tx, err := conn.BeginTx(requestContext, &sql.TxOptions{})
	s.Require().NoError(err)

	requestContext = persistence.SetTransactionContext(requestContext, tx)

	_, err = tx.ExecContext(context.TODO(), "INSERT INTO payments(payment_id,user_id,amount) VALUES ($1,$2,$3)",
		"123", "aruiz", 89.99)
	s.Require().NoError(err)

	err = s.writer.Write(requestContext, []streams.Message{
		{
			StreamName:  "org.alexandria.foo",
			ContentType: "application/text",
			Data:        []byte("foo"),
		},
		{
			StreamName:  "org.alexandria.foo",
			ContentType: "application/text",
			Data:        []byte("foo"),
		},
		{
			StreamName:  "org.alexandria.foo",
			ContentType: "application/text",
			Data:        []byte("foo"),
		},
	})
	s.Assert().NoError(err)
	s.Assert().NoError(tx.Commit())

	s.tableFlushMap.Store("payments", struct{}{})
	s.tableFlushMap.Store("streams_egress", struct{}{})
}
