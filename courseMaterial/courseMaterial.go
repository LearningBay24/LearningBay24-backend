package coursematerial

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func CreateMaterial(db *sql.DB, name, url string, usersid int, courseID int) error {

	// Begin transaction with database handle
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// TODO: local boolean

	// Commit transaction into database
	tx.Commit()
	return nil
}
