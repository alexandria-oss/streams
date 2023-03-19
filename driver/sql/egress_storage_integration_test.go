//go:build integration

package sql_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	streamsql "github.com/alexandria-oss/streams/driver/sql"
	"github.com/alexandria-oss/streams/proxy/egress"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type egressStorageIntegrationTestSuite struct {
	suite.Suite

	db      *sql.DB
	storage streamsql.EgressStorage
}

func TestEgressStorage(t *testing.T) {
	suite.Run(t, &egressStorageIntegrationTestSuite{})
}

func (s *egressStorageIntegrationTestSuite) SetupSuite() {
	db, err := sql.Open("postgres", "user=postgres password=root sslmode=disable port=6432 host=localhost database=sample_database")
	s.Require().NoError(err)
	s.db = db

	s.storage = streamsql.NewEgressStorage(db, egress.WithEgressTable(egress.DefaultEgressTableName))

	conn, err := db.Conn(context.Background())
	require.NoError(s.T(), err)
	defer conn.Close()
	query := fmt.Sprintf("INSERT INTO %s(batch_id,message_count,raw_data) VALUES ($1,$2,$3)",
		egress.DefaultEgressTableName)
	_, err = conn.ExecContext(context.Background(), query, "123", 1, []byte("the quick brown fox"))
	require.NoError(s.T(), err)
	_, err = conn.ExecContext(context.Background(), query, "456", 1, []byte("the quick brown fox 2"))
	require.NoError(s.T(), err)
}

func (s *egressStorageIntegrationTestSuite) TearDownSuite() {
	defer s.db.Close()
	conn, err := s.db.Conn(context.Background())
	require.NoError(s.T(), err)
	defer conn.Close()
	query := fmt.Sprintf("DELETE FROM %s WHERE batch_id = $1", egress.DefaultEgressTableName)
	_, err = conn.ExecContext(context.Background(), query, "123")
	require.NoError(s.T(), err)
}

func (s *egressStorageIntegrationTestSuite) TestGetBatch() {
	b, err := s.storage.GetBatch(context.TODO(), "FAKE_ID")
	assert.Error(s.T(), err)
	assert.Empty(s.T(), b)

	b, err = s.storage.GetBatch(context.TODO(), "123")
	assert.NoError(s.T(), err)
	assert.EqualValues(s.T(), egress.Batch{
		BatchID:           "123",
		TransportBatchRaw: []byte("the quick brown fox"),
	}, b)
}

func (s *egressStorageIntegrationTestSuite) TestCommit() {
	err := s.storage.Commit(context.TODO(), "456")
	assert.NoError(s.T(), err)
}
