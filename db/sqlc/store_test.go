package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	fmt.Println(">> before:", account1.Balance, account2.Balance)

	// run with concurrent go routines
	n := 4
	amount := int64(10)

	// form channels
	errs := make(chan error)
	results := make(chan TransferTxResult)

	// timeString := "2022-09-24 05:10:40"
	// theTime, _ := time.Parse("2006-01-02 03:04:05", timeString)
	// fmt.Println("Comparison time = ", theTime)
	// compareTime := theTime.UnixMilli()

	// the code below launches 10 go routines that execute 5 transactions simultaneously
	// , and for implementing
	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
		// fmt.Println(strings.Repeat("*", 30), "Comitting transaction from id = ", goid(), "at time = ", time.Now().UnixMilli()-compareTime, strings.Repeat("*", 30))
	}

	// check results
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer

		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID) // check if the entry record is actually created
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// check account balance of FromAccountID and ToAccountID
		fmt.Println(">> tx: ", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance

		// example of how to print messages when a test fails(require function)
		msgs := make([]interface{}, 1)
		msgs[0] = fmt.Sprintf("Iteration number: %d\n", i)

		require.Equal(t, diff1, diff2, msgs)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)
		/*
			1. after 1st transaction, remaninig blc = account1.balance - 10
			2. after 2nd transaction, remaninig blc = account1.balance - 20
			3. after 3rd transaction, remaninig blc = account1.balance - 30...
		*/
		/*
			Note: We cannot use require.Equal(t, -amount*(i+1), diff1) or require.Equal(t, amount*(i+1), diff2)
			since we don't know which go routine will be executed first.
			it could be that the i'th go routine is executed AFTER the i+1'th go routine;
			making the require statement false.
		*/

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}
	// check final updated balances of both accounts
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)
}

func TestTransferDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	fmt.Printf("before tx: balance(%d) = %d, balance(%d) = %d\n", account1.ID, account1.Balance, account2.ID, account2.Balance)

	n := 6
	amount := 10

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID
		if (i & 1) == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}
		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        int64(amount),
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.Balance, updatedAccount1.Balance)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)

	fmt.Printf("after tx: balance(%d) = %d, balance(%d) = %d\n", updatedAccount1.ID, updatedAccount1.Balance, updatedAccount2.ID, updatedAccount2.Balance)
}
