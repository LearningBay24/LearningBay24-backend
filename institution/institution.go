package institution

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

// GetAllUsers returns the number of registered users
func GetUserCount(db *sql.DB) (int, error) {
	// Check if there are more users in the course besides the creator
	users, err := models.Users().Count(context.Background(), db)
	if err != nil {
		return 0, err
	}
	var usercount int = int(users)

	return usercount, nil
}

// GetAllFieldsOfStudies returns a array of all field of studies available in the institution
func GetAllFieldsOfStudies(db *sql.DB) ([]*models.FieldOfStudy, error) {
	fieldofstudies, err := models.FieldOfStudies().All(context.Background(), db)

	if err != nil {
		return nil, err
	}

	return fieldofstudies, nil
}

// GetFieldOfStudy takes a id and returns a single field of study
func GetFieldOfStudy(db *sql.DB, fid int) (*models.FieldOfStudy, error) {
	fieldofstudy, err := models.FindFieldOfStudy(context.Background(), db, fid)

	if err != nil {
		return nil, err
	}

	return fieldofstudy, nil
}

// CreateFieldOfStudy takes a name and number of semesters and creates a field of study
func CreateFieldOfStudy(db *sql.DB, name null.String, semester null.Int) (int, error) {
	if name.String == "" || semester.Int <= 0 {
		return 0, errors.New("name cant be empty and semester has to be higher than 0")
	}
	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	fieldOfStudy := &models.FieldOfStudy{Name: name, Semesters: semester}

	err = fieldOfStudy.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return fieldOfStudy.ID, nil
}

// EditFieldOfStudyHasCourse takes the id,the new name and the new semester
func EditFieldOfStudy(db *sql.DB, fid int, name string, semester int) (int, error) {
	if name == "" || semester <= 0 {
		return 0, errors.New("name cant be empty and semester has to be higher than 0")
	}
	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	// Get the correponding FieldOfStudy
	fos, err := models.FindFieldOfStudy(context.Background(), db, fid)
	if err != nil {
		return 0, err
	}

	fos.Semesters.Int = semester
	fos.Name.String = name

	// Update
	_, err = fos.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return fos.ID, nil
}

// DeleteFieldOfStudy takes a name and number of semesters and deletes a field of study
func DeleteFieldOfStudy(db *sql.DB, fid int) (int, error) {

	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	fieldOfStudy, err := models.FindFieldOfStudy(context.Background(), db, fid)
	if err != nil {
		return 0, err
	}
	_, err = fieldOfStudy.Delete(context.Background(), tx, false)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return fieldOfStudy.ID, nil
}

// AddFieldOfStudyHasCourse takes a FieldOfStudy ID, Course ID, and a Semester and adds them in the Database
func AddFieldOfStudyHasCourse(db *sql.DB, fid int, cid int, semester int) error {

	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Get the correponding FieldOfStudy
	fos, err := models.FindFieldOfStudy(context.Background(), db, fid)
	if err != nil {
		return err
	}

	// Check if the new semester greater then 0 and not bigger then the max semesters
	if semester <= 0 || fos.Semesters.Int < semester {
		return fmt.Errorf("fatal: semester %d is not a valid semester", fos.Semesters.Int)
	}

	fosHasCourse := &models.FieldOfStudyHasCourse{FieldOfStudyID: fid, CourseID: cid, Semester: semester}

	err = fosHasCourse.Insert(context.Background(), tx, boil.Infer())

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return nil
}

// EditFieldOfStudyHasCourse takes a the old FieldOfStudy ID, old Course ID, new FieldOfStudy ID and the new semester and updates the database with the new FieldOfStudy and new semester
func EditFieldOfStudyHasCourse(db *sql.DB, fid int, cid int, Newsemester int) error {

	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Get the struct of the FieldOfStudyHasCourse
	fosHasCourse, err := models.FindFieldOfStudyHasCourse(context.Background(), db, fid, cid)
	if err != nil {
		return err
	}

	// Get the correponding FieldOfStudy
	fos, err := models.FindFieldOfStudy(context.Background(), db, fid)
	if err != nil {
		return err
	}

	// Check if the new semester greater then 0 and not bigger then the max semesters
	if Newsemester <= 0 || fos.Semesters.Int < Newsemester {
		return errors.New("this semester doesnt exist")
	}
	fosHasCourse.Semester = Newsemester
	// Update
	_, err = fosHasCourse.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return err
	}

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return nil
}

// DeleteFieldOfStudyHasCourse takes a field of study ID, course ID and a semester and removes relation to those two
func DeleteFieldOfStudyHasCourse(db *sql.DB, fid int, cid int) error {

	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	fosHasCourse, err := models.FindFieldOfStudyHasCourse(context.Background(), db, fid, cid)
	if err != nil {
		return err
	}

	_, err = fosHasCourse.Delete(context.Background(), tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return err
	}

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return nil
}
