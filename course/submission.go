package course

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

func GetSubmission(db *sql.DB, sid int) (*models.Submission, error) {
	s, err := models.FindSubmission(context.Background(), db, sid)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateCourse takes a name,enrollkey and description and adds a course and forum with that Name in the Database while userid is an array of IDs that is used to assign the role of the creator
// and the roles for tutor
func CreateSubmission(db *sql.DB, name string, deadline string, cid int, maxfilesize int, visiblefrom string) (int, error) {

	var dtime null.Time
	var parseddtime time.Time
	// Get current Time
	curTime := time.Now()
	//Check if name ist emtpy
	if name == "" {
		return 0, errors.New("name cant be empty")
	}
	// Parse visiblefrom String to time
	vtime, err := time.Parse(time.RFC3339, visiblefrom)
	if err != nil {
		return 0, err
	}

	// Check if visiblefrom time is in the past
	if vtime.Sub(curTime) < 0 {

		return 0, errors.New("visiblefrom time cant be in the past")
	}
	// Check if deadline is empty
	if deadline == "" {
		// null
		dtime = null.NewTime(parseddtime, false)
	} else {
		// not null
		parseddtime, err := time.Parse(time.RFC3339, deadline)
		if err != nil {
			return 0, err
		}

		dtime = null.NewTime(parseddtime, true)
		// Check if deadline time is in the past
		if dtime.Time.Sub(curTime) < 0 {

			return 0, errors.New("deadline time cant be in the past")
		}
		if dtime.Time.Sub(vtime) < 0 {
			return 0, errors.New("visible from time cant be after deadline time")
		}
	}

	// Begins the transaction

	s := &models.Submission{Name: name, Deadline: dtime, CourseID: cid, MaxFilesize: maxfilesize, VisibleFrom: vtime}

	// Inserts into database
	err = s.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		return 0, err
	}

	return s.ID, nil
}

func EditSubmission(db *sql.DB, sid int, name string, deadline string, maxfilesize int, visiblefrom string) (int, error) {

	var dtime null.Time
	var parseddtime time.Time
	// Get current Time
	curTime := time.Now()
	//Check if name ist emtpy
	if name == "" {
		return 0, errors.New("name cant be empty")
	}
	// Parse visiblefrom String to time
	vtime, err := time.Parse(time.RFC3339, visiblefrom)
	if err != nil {
		return 0, err
	}

	// Check if visiblefrom time is in the past
	if vtime.Sub(curTime) < 0 {

		return 0, errors.New("visiblefrom time cant be in the past")
	}
	// Check if deadline is empty
	if deadline == "" {
		// null
		dtime = null.NewTime(parseddtime, false)
	} else {
		// not null
		parseddtime, err := time.Parse(time.RFC3339, deadline)
		if err != nil {
			return 0, err
		}

		dtime = null.NewTime(parseddtime, true)
		// Check if deadline time is in the past
		if dtime.Time.Sub(curTime) < 0 {

			return 0, errors.New("deadline time cant be in the past")
		}
		if dtime.Time.Sub(vtime) < 0 {
			return 0, errors.New("visible from time cant be after deadline time")
		}
	}

	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	s, err := GetSubmission(db, sid)
	if err != nil {
		return 0, err
	}
	// New Values
	s.Name = name
	s.Deadline = dtime
	s.VisibleFrom = vtime
	s.MaxFilesize = maxfilesize

	_, err = s.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return s.ID, nil
}
func DeleteSubmission(db *sql.DB, sid int) (int, error) {

	s, err := GetSubmission(db, sid)
	if err != nil {
		return 0, err
	}
	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	_, err = s.Delete(context.Background(), tx, false)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return s.ID, nil
}

func CreateSubmissionHasFiles(db *sql.DB, sid int, fid int) (int, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec("INSERT INTO submission_has_files(submission_id,file_id) VALUES (?,?);", sid, fid)
	if err != nil {
		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return 0, err
}
func DeleteSubmissionHasFiles(db *sql.DB, sid int, fid int) (int, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec("DELETE FROM submission_has_files WHERE submission_id = ? AND file_id = ? ;", sid, fid)
	if err != nil {
		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return 0, err
}
