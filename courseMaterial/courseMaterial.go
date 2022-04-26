package coursematerial

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

func CreateMaterial(db *sql.DB, name, uri string, uploaderid, courseID int, local int8) error {

	// Begin transaction with database handle
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Creates File (courseMaterial) struct
	cm := &models.File{Name: name, URI: uri, UploaderID: uploaderid, Local: local}

	err = cm.Insert(context.Background(), db, boil.Infer())
	/*
		if err != nil {
			tx.Rollback()
			return err
		} else {

		}
	*/
	// Commit transaction into database
	tx.Commit()
	return nil
}
