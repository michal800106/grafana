package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/grafana/grafana/pkg/services/ngalert/store"
	"github.com/grafana/grafana/pkg/services/ngalert/tests"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/stretchr/testify/require"
)

func TestUnitOfWork(t *testing.T) {
	_, dbstore := tests.SetupTestEnv(t, testAlertingIntervalSeconds)

	t.Run("Transaction failure is returned from Execute", func(t *testing.T) {
		xact := store.NewTransaction(dbstore.SQLStore)
		xact = xact.Do(successfulStage)
		xact = xact.Do(failingStage)

		err := xact.Execute(context.Background())

		require.EqualError(t, err, "Something bad happened.")
	})
}

func successfulStage(s *sqlstore.DBSession) error {
	return nil
}

func failingStage(s *sqlstore.DBSession) error {
	return fmt.Errorf("Something bad happened.")
}

func TestUnitOfWorkIntegration(t *testing.T) {
	_, dbstore := tests.SetupTestEnv(t, testAlertingIntervalSeconds)
	setupTestRecordTable(t, dbstore)
	xact := store.NewTransaction(dbstore.SQLStore)

	xact = xact.Do(insertTestRecord)
	xact = xact.Do(insertTestRecord)
	require.Equal(t, int64(0), countRecords(t, dbstore))

	err := xact.Execute(context.Background())

	require.NoError(t, err)
	require.Equal(t, int64(2), countRecords(t, dbstore))
}

type testRecord struct {
	Id    int `xorm:"pk autoincr 'id'"`
	Value int
}

func setupTestRecordTable(t *testing.T, dbstore *store.DBstore) {
	err := dbstore.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		err := sess.DropTable(testRecord{})
		if err != nil {
			return err
		}
		err = sess.CreateTable(testRecord{})
		if err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)
}

func insertTestRecord(sess *sqlstore.DBSession) error {
	_, err := sess.Insert(testRecord{})
	return err
}

func countRecords(t *testing.T, dbstore *store.DBstore) int64 {
	result := int64(0)
	err := dbstore.SQLStore.WithDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		count, err := sess.Count(testRecord{})
		result = count
		return err
	})
	require.NoError(t, err)
	return result
}
