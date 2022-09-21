package db

import (
	"context"
	"database/sql"
	"fmt"
)

// run DB queries individually, as well as in combination within a transaction.
// each *.sql.go file only performs a single operation(C/R/U/D) on a single table(Account/Entry/Transfer)
type Store struct {
	*Queries // composition, extends functionality of Queries struct into Store struct.
	// all functions provided by Queries are now applicable on a Store instance.
	db *sql.DB // used to create a new DB transaction
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db), // from db.go
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx) //  Queries struct object
	/*
		DBTX is an interface in db.go but *sql.Tx(datatype of tx) is a struct
		which does implement all the 4 methods specified in this
		interface(line 2486 in database/sql/sql.go)
	*/
	err = fn(q) /* perform the required function:
	1. for instance a moeny transfer from account1 to account2,
	2. or a deposit/withdrawal into/from an account.*/

	if err != nil {
		// rollback transaction
		rbErr := tx.Rollback() // rollback error
		if rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr) // print multiple errors together
		}
		return err
	}
	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer      Transfer `json:"transfer"`
	FromAccountID int64    `json:"from_account_id"`
	ToAccountID   int64    `json:"to_account_id"`
	FromEntry     Entry    `json:"entry_from"` // post transaction entries
	ToEntry       Entry    `json:"entry_to"`   // post transaction entries
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	// money transfer transaction
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}
		// TODO: update account balance.
		return nil
	}) // this makes the callback function a closure, golang supports generics type
	return result, err
}
