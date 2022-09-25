package db

import "fmt"

func (account *Account) print() {
	fmt.Printf("\tAccount ID = %d\n\tOwner = %s\n\tAmount = %d %s", account.ID, account.Owner, account.Balance, account.Currency)
}
