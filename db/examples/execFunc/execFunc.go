package main

/*
	This example shows a slightly more optimized way of handling database transactions. You should
	read and understand the Exec example before this one.

	The improvements here are (1) using the BeginFunc method to handle the rollback or commit in a
	cleaner and more reliable way, and (2) moving the methods to get, insert, and delete a record
	from the ccm.customer_profile table, to receiver functions on the customerProfile struct type,
	as well as scanning a db row into a struct.
*/

import (
	"context"
	"log"
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

	// BeginFunc runs the function passed into it, then either rolls back or commits the transaction
	// based on whether an error returns from the given function.
	if err := d.BeginFunc(ctx, func(tx pgx.Tx) error {
		// insert the record
		if err := cp.insert(ctx, tx); err != nil {
			// transaction will rollback
			return err
		}

		// get the record and put it into a struct
		newCP := customerProfile{id: 0}
		if err := newCP.getByID(ctx, tx); err != nil {
			// transaction will rollback
			return err
		}

		// delete the record from the db
		if err := newCP.delete(ctx, tx); err != nil {
			// transaction will rollback
			return err
		}

		// transaction will be committed
		return nil
	}); err != nil {
		log.Fatalf("transaction error: %v:", err)
	}
}

func (c *customerProfile) delete(ctx context.Context, tx pgx.Tx) error {
	sql := `
	DELETE FROM ccm.customer_profile
	WHERE id = $1
	`

	if res, err := tx.Exec(ctx, sql, c.id); err != nil {
		return err
	} else {
		log.Printf("%d rows deleted", res.RowsAffected())
	}

	return nil
}

func (c *customerProfile) getByID(ctx context.Context, tx pgx.Tx) error {
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

	row := tx.QueryRow(ctx, sql.String(), c.id)

	if err := c.scan(row); err != nil {
		return err
	}

	log.Printf("got customer profile: %+v", c)

	return nil
}

func (c *customerProfile) insert(ctx context.Context, tx pgx.Tx) error {
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
	if res, err := tx.Exec(
		ctx,
		sql,
		pgx.QuerySimpleProtocol(true),
		c.id,
		c.userID,
		c.curp,
		c.firstName,
		c.lastName1,
		c.birthDate,
		c.email,
		c.createdTMS,
		c.updateTMS,
	); err != nil {
		return err
	} else {
		log.Printf("%d rows inserted\n", res.RowsAffected())
	}

	return nil
}

func (c *customerProfile) scan(row pgx.Row) error {
	if err := row.Scan(
		&c.id,
		&c.userID,
		&c.curp,
		&c.firstName,
		&c.lastName1,
		&c.birthDate,
		&c.email,
		&c.createdTMS,
		&c.updateTMS,
	); err != nil {
		return err
	}

	return nil
}
