package db

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"strconv"
	"strings"
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

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
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
	Transfer    Transfer `json:"transfer"`
	FromEntry   Entry    `json:"entry_from"` // post transaction entries
	ToEntry     Entry    `json:"entry_to"`   // post transaction entries
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID int64,
	amount int64,
) (account_ Account, err error) {
	account_, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID,
		Amount: amount,
	})
	if err != nil {
		return
	}
	return
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	// money transfer transaction
	// timeString := "2022-09-24 05:10:40"
	// theTime, _ := time.Parse("2006-01-02 03:04:05", timeString)
	// compareTime := theTime.UnixMilli()

	// fmt.Println("For transaction belonging to goroutines ", goid())

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
		// var accountFrom, accountTo Account
		// fmt.Printf("Starting fetching From Account = %d w.r.t. goroutine id = %d\n", time.Now().UnixMilli()-compareTime, goid())
		if arg.FromAccountID < arg.ToAccountID {
			// update FromAccountID first, since its smaller
			result.FromAccount, err = addMoney(ctx, store.Queries, arg.FromAccountID, -arg.Amount)
			if err != nil {
				return err
			}
			// fmt.Printf("Completed updating From Account = %d w.r.t. goroutine id = %d\n", time.Now().UnixMilli()-compareTime, goid())

			result.ToAccount, err = addMoney(ctx, store.Queries, arg.ToAccountID, arg.Amount)
			if err != nil {
				return err
			}
			// fmt.Printf("Completed updating To Account = %d w.r.t. goroutine id = %d\n", time.Now().UnixMilli()-compareTime, goid())
		} else {
			// update ToAccountID first, since its smaller
			result.ToAccount, err = addMoney(ctx, store.Queries, arg.ToAccountID, arg.Amount)
			if err != nil {
				return err
			}
			// fmt.Printf("Completed updating To Account = %d w.r.t. goroutine id = %d\n", time.Now().UnixMilli()-compareTime, goid())

			result.FromAccount, err = addMoney(ctx, store.Queries, arg.FromAccountID, -arg.Amount)
			if err != nil {
				return err
			}
			// fmt.Printf("Completed updating From Account = %d w.r.t. goroutine id = %d\n", time.Now().UnixMilli()-compareTime, goid())
		}

		// fmt.Printf("Completed updating To Account = %d w.r.t. goroutine id = %d\n", time.Now().UnixMilli()-compareTime, goid())

		return nil
	}) // this makes the callback function a closure, golang supports generics type
	// if debug{
	// 	fmt.Printf("Transferring from(go routine id = %d):\nInitial State:\n", goid())
	// 	accountFrom.print()
	// 	fmt.Printf("\nFinal State(go routine id = %d):\n", goid())
	// 	result.FromAccount.print()
	// 	fmt.Println("\nTransferring To:")
	// 	result.ToAccount.print()
	// 	fmt.Println()
	// 	fmt.Println(strings.Repeat("-", 50))
	// }
	return result, err
}
