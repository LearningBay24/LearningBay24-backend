package coursematerial

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/null/v8"
)

func CreateMaterial(db *sql.DB, name, enrollkey string, description null.String, usersid []int) error {

	//Begin transaction with database handle
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	//Commit transaction into database
	tx.Commit()
	return nil
}
