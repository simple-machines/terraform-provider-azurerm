package azurerm

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// a successful case
func TestShouldUpdateStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE products(name varchar(20))").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
