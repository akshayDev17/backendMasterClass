package db
import (
	"testing"
	"context"
	"github.com/stretchr/testify/require"
)
func TestCreateAccount(t *testing.T){
	arg := CreateAccountParams{
		Owner: "Tom",
		Balance: 500,
		Currency: "USD",
	}
	account, err := testQueries.CreateAccount(context.Background(), arg) // testQueries defined in main_test.go, also CreateAccount is implemented by the DBTX type in account.sql.go
	require.NoError(t, err) // checks if error is nil, fails the test if its not
	require.NotEmpty(t, account)// account should not be empty

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Currency, account.Currency)
	require.Equal(t, arg.Balance, account.Balance)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

}