package coursematerial

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

func CreateMaterial(db *sql.DB, name string, uri string, uploaderid, courseID int, local int8) error {

	// Begin transaction with database handle
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Creates File (courseMaterial) struct
	cm := &models.File{Name: name, URI: uri, UploaderID: uploaderid, Local: local}

	err = cm.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		tx.Rollback()
		return err
	}
	// Commit transaction into database
	tx.Commit()
	return nil
}

// Takes id and updates name and uri
func UpdateMaterial(db *sql.DB, id int, name, uri string) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	c, err := models.FindFile(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	c.Name = name
	c.URI = uri

	_, err = c.Update(context.Background(), db, boil.Infer())

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func DeleteFile(db *sql.DB, id int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	c, err := models.FindFile(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = c.Delete(context.Background(), db, true)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
