// Package db provides an interface to connect to and interact with the Postgres database in our AWS
// environment. It manages the connection pooling, and use of the correct read or write proxy, as
// needed.
package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/auxten/postgresql-parser/pkg/sql/parser"
	"github.com/auxten/postgresql-parser/pkg/sql/sem/tree"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	numeric "github.com/jackc/pgtype/ext/shopspring-numeric"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/credifranco/stori-utils-go/aws"
)

// DBConnector provides methods to connect to and execute SQL operations
type DBConnector interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginFunc(ctx context.Context, f func(pgx.Tx) error) error
	Close()
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// DBProxyError implements the pgx.Row interface
type DBProxyError struct {
	err error
}

// Scan just returns an error so that we can return an error value from our QueryRow method
func (d DBProxyError) Scan(dest ...interface{}) error { return d.err }

type dbSecret struct {
	UserName string `json:"username"`
	Engine   string `json:"engine"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
	exp      time.Time
}

// AWSDB implements DBConnector that utilizes separate AWS RDS Postgres read and write database pools
type AWSDB struct {
	reader *pgxpool.Pool
	writer *pgxpool.Pool
}

type DBProxies string

const (
	Read         = DBProxies("read")
	Write        = DBProxies("write")
	ReadAndWrite = DBProxies("readAndWrite")
)

var secretNames = map[DBProxies]string{
	Read:  "core_iam_user_read",
	Write: "core_iam_user_write",
}

var ErrReaderNotCreated = errors.New("reader pool not created")
var ErrWriterNotCreated = errors.New("writer pool not created")

// Begin acquires a connection from the writer Pool and starts a transaction.
func (d *AWSDB) Begin(ctx context.Context) (pgx.Tx, error) {
	if d.writer == nil {
		return nil, ErrWriterNotCreated
	}

	return d.writer.Begin(ctx)
}

// BeginFunc begins a transaction, executes the function `f`, and commits or rolls back the
// transaction, depending on the return value of the function.
func (d *AWSDB) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	if d.writer == nil {
		return ErrWriterNotCreated
	}

	return d.writer.BeginFunc(ctx, f)
}

// Exec acquires a connection from the Pool and executes the given SQL. SQL can be either a prepared
// statement name or an SQL string. Arguments should be referenced positionally from the SQL string
// as $1, $2, etc. The acquired connection is returned to the pool when the Exec function returns.
func (d *AWSDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if d.writer == nil {
		return nil, ErrWriterNotCreated
	}

	return d.writer.Exec(ctx, sql, args...)
}

// Query acquires a connection and executes a query that returns pgx.Rows.
// Arguments should be referenced positionally from the SQL string as $1, $2, etc.
// See pgx.Rows documentation to close the returned Rows and return the acquired connection to the
// Pool.
func (d *AWSDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	proxy, err := getProxyFromSQL(sql)
	if err != nil {
		return nil, fmt.Errorf("could not parse sql string: %v", err)
	}

	switch proxy {
	case Read:
		if d.reader == nil {
			return nil, ErrReaderNotCreated
		}

		return d.reader.Query(ctx, sql, args...)
	case Write:
		if d.writer == nil {
			return nil, ErrWriterNotCreated
		}

		return d.writer.Query(ctx, sql, args...)
	default:
		return nil, errors.New("error getting db proxy")
	}
}

// QueryRow acquires a connection and executes a query that is expected to return at most one
// row (pgx.Row). Errors are deferred until pgx.Row's Scan method is called. If the query selects no
// rows, pgx.Row's Scan will return ErrNoRows. Otherwise, pgx.Row's Scan scans the first selected
// row and discards the rest. The acquired connection is returned to the Pool when pgx.Row's Scan
// method is called.
//
// Arguments should be referenced positionally from the SQL string as $1, $2, etc.
func (d *AWSDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	proxy, err := getProxyFromSQL(sql)
	if err != nil {
		return DBProxyError{fmt.Errorf("could not parse sql string: %v", err)}
	}

	switch proxy {
	case Read:
		if d.reader == nil {
			return DBProxyError{ErrReaderNotCreated}
		}

		return d.reader.QueryRow(ctx, sql, args...)
	case Write:
		if d.writer == nil {
			return DBProxyError{ErrWriterNotCreated}
		}

		return d.writer.QueryRow(ctx, sql, args...)
	default:
		return DBProxyError{errors.New("error getting db proxy")}
	}
}

// getProxyFromSQL parses out a SQL string and determines which db proxy to return based on if a
// write operation is found
func getProxyFromSQL(sql string) (DBProxies, error) {
	stmts, err := parser.Parse(sql)
	if err != nil {
		return "", err
	}

	for _, s := range stmts {
		if tree.CanWriteData(s.AST) || tree.CanModifySchema(s.AST) {
			return Write, nil
		}
	}

	// no statement in the tree needed to write
	return Read, nil
}

// connect opens and validates a connection to the requested db proxy
func (db *AWSDB) connect(ctx context.Context, dp DBProxies, c chan error, wg *sync.WaitGroup) {
	var err error
	var rs dbSecret
	defer wg.Done()

	if err := rs.getValuesAWS(secretNames[dp]); err != nil {
		c <- err
		return
	}

	switch dp {
	case Read:
		if db.reader, err = rs.dbOpen(ctx); err != nil {
			c <- err
			return
		}

		if err := db.reader.Ping(ctx); err != nil {
			c <- err
			return
		}
	case Write:
		if db.writer, err = rs.dbOpen(ctx); err != nil {
			c <- err
			return
		}

		if err := db.writer.Ping(ctx); err != nil {
			c <- err
			return
		}
	}

	log.Printf("database %v proxy connected", dp)
}

// NewConnection sets up the Reader and/or Writer values of `db`, based on `dp`. It will return an
// error if retrieving the relevant AWS Secrets Manager secrets fail, or if the setup of
// the database connection fails, or the database is unreachable by (*sql.DB).Ping()
func (db *AWSDB) NewConnection(ctx context.Context, dp DBProxies) error {
	errs := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Wait()
		close(errs)
	}()

	switch dp {
	case Read:
		go db.connect(ctx, Read, errs, &wg)
	case Write:
		go db.connect(ctx, Write, errs, &wg)
	case ReadAndWrite:
		wg.Add(1)
		go db.connect(ctx, Read, errs, &wg)
		go db.connect(ctx, Write, errs, &wg)
	default:
		wg.Done()
		return fmt.Errorf("%v, %v, or %v not provided", Read, Write, ReadAndWrite)
	}

	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// CloseConnection calls the sql.Close method on the reader and writer. This should be called in
// a defer statement after successfully calling (*AWSDB).NewConnection()
func (d *AWSDB) Close() {
	if d.reader != nil {
		d.reader.Close()
	}

	if d.writer != nil {
		d.writer.Close()
	}
}

func (d *AWSDB) pingProxy(ctx context.Context, dp DBProxies, c chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	switch dp {
	case Read:
		if err := d.reader.Ping(ctx); err != nil {
			c <- err
			return
		}
	case Write:
		if err := d.writer.Ping(ctx); err != nil {
			c <- err
			return
		}
	}
}

func (d *AWSDB) Ping(ctx context.Context) error {
	errs := make(chan error)
	var wg sync.WaitGroup

	go func() {
		wg.Wait()
		close(errs)
	}()

	if d.reader != nil {
		wg.Add(1)
		go d.pingProxy(ctx, Read, errs, &wg)
	}

	if d.writer != nil {
		wg.Add(1)
		go d.pingProxy(ctx, Write, errs, &wg)
	}

	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// getValuesAWS populates `ds` from the values stored in AWS Secrets Manager with Secret ID `name`
func (ds *dbSecret) getValuesAWS(name string) error {
	r, err := aws.GetSecret(name)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(r), &ds); err != nil {
		return err
	}

	return nil
}

// dpOpen wraps sql.Open to work with AWS IAM authentication
func (ds dbSecret) dbOpen(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		ds.Host,
		ds.Port,
		ds.UserName,
		"foo",
		ds.DBName,
	)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// we were getting errors related to prepared statements. This resolves that.
	cfg.ConnConfig.PreferSimpleProtocol = true
	// match the RDS timeout config, minus 1 second
	cfg.MaxConnIdleTime = time.Second * 59

	// Refresh rds token before connect (token expires after 15 min)
	cfg.BeforeConnect = func(c context.Context, cc *pgx.ConnConfig) error {
		now := time.Now()
		if now.After(ds.exp) {
			if cc.Config.Password, err = aws.GetRDSAuthToken(ctx, ds.Host, ds.UserName, ds.Port); err != nil {
				return err
			}
			// Set expire token expire time in 14m 50s
			ds.exp = time.Now().Add(time.Second * 890)
		}
		return nil
	}

	// Register custom types for non-native types
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// postgres `numeric` = decimal.Decimal
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &numeric.Numeric{},
			Name:  "numeric",
			OID:   pgtype.NumericOID,
		})

		// postgres `uuid` = gofrs uuid.UUID
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &pgtype.UUID{},
			Name:  "uuid",
			OID:   pgtype.UUIDOID,
		})

		return nil
	}

	return pgxpool.ConnectConfig(ctx, cfg)
}
