package db_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/credifranco/stori-utils-go/db"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
)

// mockAWSDB is used to test db.AWSDB method errors without needing to connect to a live database
type mockAWSDB struct {
	db.AWSDB
	reader pgxmock.PgxPoolIface
	writer pgxmock.PgxPoolIface
}

// NewConnection will create mock connection pools using pgmock. Since we are not testing actual db
// operations, all that is needed is for the reader and writer pools to exist.
func (md *mockAWSDB) NewConnection(ctx context.Context, dp db.DBProxies) error {
	pool, err := pgxmock.NewPool()
	if err != nil {
		return err
	}

	switch dp {
	case db.Read:
		md.reader = pool
	case db.Write:
		md.writer = pool
	case db.ReadAndWrite:
		md.reader = pool
		md.writer = pool
	default:
		return fmt.Errorf("%v, %v, or %v not provided", db.Read, db.Write, db.ReadAndWrite)
	}

	return nil
}

func TestNewConnection(t *testing.T) {
	ctx := context.Background()
	var d db.AWSDB

	expectedErr := "read, write, or readAndWrite not provided"
	err := d.NewConnection(ctx, "not read or write")

	assert.Equal(
		t,
		expectedErr,
		err.Error(),
		"NewConnection should fail if correct db proxy is not named",
	)
}

func TestReadOnly(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	var d mockAWSDB

	if err := d.NewConnection(ctx, db.Read); err != nil {
		t.Fatalf("error making new db connection: %v", err)
	}

	_, err := d.Exec(ctx, "")
	a.ErrorIs(
		db.ErrWriterNotCreated,
		err,
		"should not be able to use Exec method if writer proxy was not created",
	)

	_, err = d.Begin(ctx)
	a.ErrorIs(
		db.ErrWriterNotCreated,
		err,
		"should not be able to use Begin method if writer proxy was not created",
	)

	err = d.BeginFunc(ctx, func(pgx.Tx) error { return nil })
	a.ErrorIs(
		db.ErrWriterNotCreated,
		err,
		"should not be able to use BeginFunc method if writer proxy was not created",
	)

	// Query func proxy selector
	_, err = d.Query(ctx, "CREATE TABLE some_table(id INT);")
	a.ErrorIs(
		db.ErrWriterNotCreated,
		err,
		"should not be able to do CREATE operation if writer proxy was not created",
	)

	_, err = d.Query(ctx, "DELETE FROM some_table WHERE id = 1;")
	a.ErrorIs(
		db.ErrWriterNotCreated,
		err,
		"should not be able to do DELETE operation if writer proxy was not created",
	)

	_, err = d.Query(ctx, "INSERT INTO some_table(id) VALUES (1);")
	a.ErrorIs(
		db.ErrWriterNotCreated,
		err,
		"should not be able to do INSERT operation if writer proxy was not created",
	)

	_, err = d.Query(ctx, "UPDATE some_table SET id = 2 WHERE id = 1;")
	a.ErrorIs(
		db.ErrWriterNotCreated,
		err,
		"should not be able to do UPDATE operation if writer proxy was not created",
	)

	// QueryRow func proxy selector
	err = d.QueryRow(ctx, "CREATE TABLE some_table(id INT);").Scan()
	a.ErrorIs(
		err,
		db.ErrWriterNotCreated,
		"should not be able to do CREATE operation if writer proxy was not created",
	)

	err = d.QueryRow(ctx, "DELETE FROM some_table WHERE id = 1;").Scan()
	a.ErrorIs(
		err,
		db.ErrWriterNotCreated,
		"should not be able to do DELETE operation if writer proxy was not created",
	)

	err = d.QueryRow(ctx, "INSERT INTO some_table(id) VALUES (1);").Scan()
	a.ErrorIs(
		err,
		db.ErrWriterNotCreated,
		"should not be able to do INSERT operation if writer proxy was not created",
	)

	err = d.QueryRow(ctx, "UPDATE some_table SET id = 2 WHERE id = 1;").Scan()
	a.ErrorIs(
		err,
		db.ErrWriterNotCreated,
		"should not be able to do UPDATE operation if writer proxy was not created",
	)
}

func TestWriteOnly(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	var d mockAWSDB

	if err := d.NewConnection(ctx, db.Write); err != nil {
		t.Fatalf("error making new db connection: %v", err)
	}

	// Query func proxy selector
	_, err := d.Query(ctx, "SELECT 1")
	a.Equal(
		db.ErrReaderNotCreated,
		err,
		"should not be able to do SELECT operation if reader proxy was not created",
	)

	// QueryRow func proxy selector
	err = d.QueryRow(ctx, "SELECT 1").Scan()
	a.Equal(
		err,
		db.ErrReaderNotCreated,
		"should not be able to do SELECT operation if reader proxy was not created",
	)
}
