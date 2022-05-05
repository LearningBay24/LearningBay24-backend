package exam

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

func CreateExam(db *sql.DB, name, description string, location null.String, creatorId, courseID, duration int, online, graded int8,
	date time.Time, registerDeadline, deregisterDeadline null.Time) error {

	// Begin transaction with database handle
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Creates File (courseMaterial) struct
	ex := &models.Exam{Name: name, Description: description, Date: date, Duration: duration, Online: online, Location: location,
		CourseID: courseID, CreatorID: creatorId, Graded: graded, RegisterDeadline: registerDeadline, DeregisterDeadline: deregisterDeadline}

	err = ex.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction into database
	tx.Commit()
	return nil
}

// Takes id and updates name and uri
func UpdateExam(db *sql.DB, description string, location null.String, id, duration int, online, graded int8,
	date time.Time, registerDeadline, deregisterDeadline null.Time) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	ex, err := models.FindExam(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	ex.Description = description
	ex.Location = location
	ex.Duration = duration
	ex.Online = online
	ex.Graded = graded
	ex.Date = date
	ex.RegisterDeadline = registerDeadline
	ex.DeregisterDeadline = deregisterDeadline

	_, err = ex.Update(context.Background(), db, boil.Infer())

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func DeleteExam(db *sql.DB, id int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	ex, err := models.FindFile(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = ex.Delete(context.Background(), db, true)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
