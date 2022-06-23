package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"

	"github.com/credifranco/stori-utils-go/db"
)

/*
	Notes about types:

	decimal.Decimal - Go doesn't have a built-in decimal type, and we don't want to cast the
		Postgres numeric values to float32 or float64. This third-party type works well with
		Postgres numeric values

	uuid.UUID - Same as decimal, there is no native UUID type.

	Why all the pointers? - Nullable fields in Postgress will return a NULL where there is no value.
		The pgx.Rows.Scan method casts NULL values to nil. This creates a runtime error for types
		where nil is an invalid value. So pointers need to be used for the struct fields that match
		up to a nullable field.
*/

// accountProfile is a struct representation of the rows in the ccm.account_profile table
type accountProfile struct {
	id                   int
	acctUUID             uuid.UUID
	externalAcctID       string
	custID               *int
	userID               string
	creditLimit          decimal.Decimal
	creditBalance        decimal.Decimal
	authCreditBal        decimal.Decimal
	availCredit          decimal.Decimal
	instllmntLimit       decimal.Decimal
	authBalInInstllmnts  decimal.Decimal
	instllmntOstndgBal   decimal.Decimal
	avlblBalInInstllmnts decimal.Decimal
	intrnlTrkngAcctType  *string
	acctStatusCode       *int
	acctStatus           *string
	acctType             string
	acctCode             *string
	acctSinceDate        *time.Time
	acctEndDate          *time.Time
	productCode          *string
	subProductCode       *string
	taxCode              *string
	countryCode          *string
	baseCurrency         *string
	nextDueDate          *time.Time
	lastDueDate          *time.Time
	cntrctResDate        *time.Time
	lastCntrctUpdateDate *time.Time
	lastPmtDate          *time.Time
	bsnsAcctInd          *int
	bsnsShortName        *string
	bsnsFullName         *string
	blockCode            *int
	blockCodeDesc        *string
	blockFlag            *string
	blockDesc            *string
	allowDebitsInd       *int
	allowCreditsInd      *int
	acctgClxn            *string
	acctgClxnDesc        *string
	covenantCode         *string
	covenantDesc         *string
	cntrctSitFlag        *int
	createdTms           time.Time
	updatedTms           time.Time
	clabe                *string
}

func main() {
	ctx := context.Background()
	var d db.AWSDB

	// Open connection to the database. The AWSDB type will manage both read and write connections.
	// The underlying pgxpool package creates a connection pool for each.
	if err := d.NewConnection(ctx, db.Read); err != nil {
		log.Fatalf("error establishing db connection: %v", err)
	}
	// Defered function calls are run when we exit this scope. In this case, when `main` returns.
	// This guarantees we won't leave the connection hanging. Like with `NewConnection`, both read
	// and write connections will be closed for us.
	defer d.Close()

	// Declare the columns we want to get from the table
	cols := []string{
		"id",
		"acct_uuid",
		"external_acct_id",
		"cust_id",
		"user_id",
		"credit_limit",
		"credit_balance",
		"auth_credit_bal",
		"available_credit",
		"installment_limit",
		"auth_bal_in_installments",
		"installment_outstanding_bal",
		"available_bal_in_installments",
		"internal_tracking_acct_type",
		"acct_status_code",
		"acct_status",
		"acct_type",
		"acct_code",
		"acct_since_date",
		"acct_end_date",
		"product_code",
		"sub_product_code",
		"tax_code",
		"country_code",
		"base_currency",
		"next_due_date",
		"last_due_date",
		"contract_resolution_date",
		"last_contract_update_date",
		"last_pmt_date",
		"business_acct_ind",
		"business_short_name",
		"business_full_name",
		"block_code",
		"block_code_desc",
		"block_flag",
		"block_desc",
		"allow_debits_ind",
		"allow_credits_ind",
		"accounting_classification",
		"accounting_classification_desc",
		"covenant_code",
		"covenant_desc",
		"contract_situation_flag",
		"created_tms",
		"updated_tms",
		"clabe",
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
	sql.WriteString("FROM ccm.account_profile ")
	sql.WriteString("LIMIT $1")

	// Make the database call
	rows, err := d.Query(ctx, sql.String(), 10)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	var aps []accountProfile
	for rows.Next() {
		var ap accountProfile
		// Scan copies the values from the current row into the destination struct fields.
		if err := rows.Scan(
			&ap.id,
			&ap.acctUUID,
			&ap.externalAcctID,
			&ap.custID,
			&ap.userID,
			&ap.creditLimit,
			&ap.creditBalance,
			&ap.authCreditBal,
			&ap.availCredit,
			&ap.instllmntLimit,
			&ap.authBalInInstllmnts,
			&ap.instllmntOstndgBal,
			&ap.avlblBalInInstllmnts,
			&ap.intrnlTrkngAcctType,
			&ap.acctStatusCode,
			&ap.acctStatus,
			&ap.acctType,
			&ap.acctCode,
			&ap.acctSinceDate,
			&ap.acctEndDate,
			&ap.productCode,
			&ap.subProductCode,
			&ap.taxCode,
			&ap.countryCode,
			&ap.baseCurrency,
			&ap.nextDueDate,
			&ap.lastDueDate,
			&ap.cntrctResDate,
			&ap.lastCntrctUpdateDate,
			&ap.lastPmtDate,
			&ap.bsnsAcctInd,
			&ap.bsnsShortName,
			&ap.bsnsFullName,
			&ap.blockCode,
			&ap.blockCodeDesc,
			&ap.blockFlag,
			&ap.blockDesc,
			&ap.allowDebitsInd,
			&ap.allowCreditsInd,
			&ap.acctgClxn,
			&ap.acctgClxnDesc,
			&ap.covenantCode,
			&ap.covenantDesc,
			&ap.cntrctSitFlag,
			&ap.createdTms,
			&ap.updatedTms,
			&ap.clabe,
		); err != nil {
			log.Printf("scan error: %v\n", err.Error())
		}
		aps = append(aps, ap)
	}

	// Now we have a properly formatted slice of accountProfile structs ready to work with
	for i, ap := range aps {
		fmt.Printf("%d: %v\n", i, ap)
	}
}
