package coursematerial

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

// CreateMaterial takes a name, URI, associated uploader-id, course, id and indicator if file is local or remote
// Created struct gets inserted into database
func CreateMaterial(db *sql.DB, name string, uri string, uploaderid, courseID int, local int8) error {

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	cm := &models.File{Name: name, URI: uri, UploaderID: uploaderid, Local: local}

	err = cm.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		tx.Rollback()
		return err
	} else {

	}
	tx.Commit()
	return nil
}

// DeactivateMaterial takes an ID and deactivates the chosen material
// Sets deactivation-timer and updates database
func DeactivateMaterial(db *sql.DB, id int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	cm, err := models.FindFile(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	cm.DeletedAt = null.NewTime(time.Now(), true)
	_, err = cm.Update(context.Background(), db, boil.Infer())
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

/*
 Switches out file if creator wants to change it
 Takes id and updates name and uri
*/
/*
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
*/
