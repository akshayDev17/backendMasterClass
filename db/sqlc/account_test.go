package db
import (
	"testing"
	"context"
	"github.com/stretchr/testify/require"
	util "my/backendMasterclass/util"
	"database/sql"
	"time"
)

func createRandomAccount(t *testing.T) Account{
	arg := CreateAccountParams{
		Owner: util.RandomOwner(),
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account, err := testQueries.CreateAccount(context.Background(), arg) // testQueries defined in main_test.go, also CreateAccount is implemented by the DBTX type in account.sql.go
	require.NoError(t, err) // checks if error is nil, fails the test if its not
	require.NotEmpty(t, account)// account should not be empty

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Currency, account.Currency)
	require.Equal(t, arg.Balance, account.Balance)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}
func TestCreateAccount(t *testing.T){
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T){
	account_created := createRandomAccount(t)
	account_fetched, err := testQueries.GetAccount(context.Background(), account_created.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account_fetched)

	require.Equal(t, account_created.Owner, account_fetched.Owner)
	require.Equal(t, account_created.Balance, account_fetched.Balance)
	require.Equal(t, account_created.Currency, account_fetched.Currency)
	require.WithinDuration(t, account_created.CreatedAt, account_fetched.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T){
	account_created := createRandomAccount(t)
	arg := UpdateAccountParams{
		ID: account_created.ID,
		Balance: util.RandomMoney(),
	}

	updated_account, err := testQueries.UpdateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, updated_account)

	require.Equal(t, account_created.Owner, updated_account.Owner)
	require.Equal(t, arg.Balance, updated_account.Balance)
	require.Equal(t, account_created.Currency, updated_account.Currency)
	require.WithinDuration(t, account_created.CreatedAt, updated_account.CreatedAt, time.Second)

}

func TestDeleteAccount(t *testing.T){
	account_created := createRandomAccount(t)
	err := testQueries.DeleteAccount(context.Background(), account_created.ID)
	require.NoError(t, err)

	deleted_account, err := testQueries.GetAccount(context.Background(), account_created.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, deleted_account)
}
