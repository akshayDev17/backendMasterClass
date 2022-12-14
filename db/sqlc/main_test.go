// houses the connection object to DB
package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://postgres:mysecretpassword@localhost:5432/simple_bank?sslmode=disable"
) // not the correct way of declaring things, we do this in an environment file instead
var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	// convention: main entry point for all unit testing inside a specific Golang package
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to database")
	}
	testQueries = New(testDB)
	// os.Exit(m.Run()) // m.Run() = returns an exit code telling us passing/failing of a test.
	exitVal := m.Run()
	log.Println("Do stuff AFTER the tests!")

	os.Exit(exitVal)
}
