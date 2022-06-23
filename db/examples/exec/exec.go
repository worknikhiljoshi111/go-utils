package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/credifranco/stori-utils-go/db"
	"github.com/jackc/pgx/v4"
)

type customerProfile struct {
	id         int
	userID     string
	curp       string
	firstName  string
	lastName1  string
	birthDate  time.Time
	email      string
	createdTMS time.Time
	updateTMS  time.Time
}

func main() {
	ctx := context.Background()
	var d db.AWSDB

	if err := d.NewConnection(ctx, db.Write); err != nil {
		log.Fatalf("error establishing db connection: %v", err)
	}
	defer d.Close()

	// begin transaction
	tx, err := d.Begin(ctx)
	if err != nil {
		log.Fatalf("error creating db transaction: %v", err)
	}

	// get some vaules to put into the struct
	bd, err := time.Parse("2006-01-02", "1986-05-27")
	if err != nil {
		log.Fatalf("error creating birthdate: %v", err)
	}
	now := time.Now()

	cp := customerProfile{
		id:         0,
		userID:     "fake-user",
		curp:       "fake-curp",
		firstName:  "Stori",
		lastName1:  "Dev",
		birthDate:  bd,
		email:      "email@example.com",
		createdTMS: now,
		updateTMS:  now,
	}

	// attempt to insert the record into ccm.customer_profile
	if err := insert(ctx, tx, cp); err != nil {
		log.Printf("error inserting record: %v", err)

		if err := tx.Rollback(ctx); err != nil {
			log.Printf("error rolling back transaction")
		}
		log.Println("rolled back transaction")
		os.Exit(1)
	}

	// get the record we just inserted
	if foundCP, err := get(ctx, tx, cp.id); err != nil {
		log.Printf("error getting record: %v", err)

		if err := tx.Rollback(ctx); err != nil {
			log.Printf("error rolling back transaction")
		}
		log.Println("rolled back transaction")
		os.Exit(1)
	} else {
		log.Printf("found customer profile: %+v\n", foundCP)
	}

	// remove the record, let's not clog up the table with arbitrary records
	if err := delete(ctx, tx, cp.id); err != nil {
		log.Printf("error deleting record: %v", err)

		if err := tx.Rollback(ctx); err != nil {
			log.Printf("error rolling back transaction")
		}
		log.Println("rolled back transaction")
		os.Exit(1)
	}

	// commit the transaction, all things have succeeded
	if err := tx.Commit(ctx); err != nil {
		// if er
		log.Fatalf("error commiting transaction: %v", err)
	}

	log.Println("transaction successfully committed")
}

// delete deletes row from ccm.customer_profile with given id
func delete(ctx context.Context, tx pgx.Tx, id int) error {
	sql := `
	DELETE FROM ccm.customer_profile
	WHERE id = $1
	`

	res, err := tx.Exec(ctx, sql, pgx.QuerySimpleProtocol(true), id)
	if err != nil {
		return err
	}

	log.Printf("%d rows deleted", res.RowsAffected())

	return nil
}

// get returns a single row from ccm.customer_profile with given id
func get(ctx context.Context, tx pgx.Tx, id int) (customerProfile, error) {
	cols := []string{
		"id",
		"user_id",
		"curp",
		"first_name",
		"last_name_1",
		"birth_date",
		"email",
		"created_tms",
		"update_tms",
	}

	// Build the SQL query string. strings.Builder isn't strictly necessary here with the amount
	// of columns we have, but it's a very efficient way to concatenate a string repeatedly, and
	// it's good practice to use it, and I just wanted to show an example of it.
	var sql strings.Builder
	sql.WriteString("SELECT ")
	for i, c := range cols {
		sql.WriteString(c)
		if i < len(cols)-1 {
			sql.WriteString(", ")
		} else {
			sql.WriteString(" ")
		}
	}
	sql.WriteString("FROM ccm.customer_profile ")
	sql.WriteString("WHERE id = $1")

	row := tx.QueryRow(ctx, sql.String(), pgx.QuerySimpleProtocol(true), id)

	var cp customerProfile
	if err := row.Scan(
		&cp.id,
		&cp.userID,
		&cp.curp,
		&cp.firstName,
		&cp.lastName1,
		&cp.birthDate,
		&cp.email,
		&cp.createdTMS,
		&cp.updateTMS,
	); err != nil {
		return customerProfile{}, err
	}

	return cp, nil
}

// insert inserts into ccm.customer_profile mapping fields from cp into the table columns
func insert(ctx context.Context, tx pgx.Tx, cp customerProfile) error {
	sql := `
	INSERT INTO
		ccm.customer_profile (
			id,
			user_id,
			curp,
			first_name,
			last_name_1,
			birth_date,
			email,
			created_tms,
			update_tms
		)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	res, err := tx.Exec(
		ctx,
		sql,
		pgx.QuerySimpleProtocol(true),
		cp.id,
		cp.userID,
		cp.curp,
		cp.firstName,
		cp.lastName1,
		cp.birthDate,
		cp.email,
		cp.createdTMS,
		cp.updateTMS,
	)
	if err != nil {
		return err
	}

	log.Printf("%d rows inserted\n", res.RowsAffected())

	return nil
}
