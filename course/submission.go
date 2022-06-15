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
	// TODO: ADD FILE TO SUBMISSION_HAS_FILES
	_, err = tx.Exec("INSERT INTO submission_has_files(submission_id,file_id) VALUES (?,?);", sid, fid)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return 0, err
}
func DeleteSubmissionHasFiles(db *sql.DB, sid int, fid int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM submission_has_files WHERE submission_id = ? AND file_id = ? ;", sid, fid)
	if err != nil {
		return err
	}
	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}

	return nil
}

func CreateUserSubmission(db *sql.DB, name string, submitter_id int, submission_id int, ignores_submission_deadline int8) (int, error) {
	var nullname null.String
	curtime := time.Now()
	if name == "" {
		nullname = null.NewString(name, false)
	} else {
		nullname = null.NewString(name, true)
	}
	if ignores_submission_deadline == 0 {
		subm, err := models.FindSubmission(context.Background(), db, submission_id)
		if err != nil {
			return 0, err
		}

		if subm.Deadline.Time.Sub(curtime) < 0 {
			return 0, errors.New("past Deadline time")
		}
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	uhassubmission := models.UserSubmission{Name: nullname, SubmitterID: submitter_id, SubmissionID: submission_id, IgnoresSubmissionDeadline: ignores_submission_deadline, SubmissionTime: null.NewTime(curtime, true)}

	err = uhassubmission.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}

	return uhassubmission.ID, nil
}
func EditUserSubmission(db *sql.DB, user_submission_id int, name string) (int, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	uhassubmission, err := models.FindUserSubmission(context.Background(), db, user_submission_id)
	if err != nil {
		return 0, err
	}
	if name == "" {
		uhassubmission.Name = null.NewString(name, false)
	} else {
		uhassubmission.Name = null.NewString(name, true)
	}

	_, err = uhassubmission.Update(context.Background(), tx, boil.Infer())
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}

	return uhassubmission.ID, nil
}
func DeleteUserSubmission(db *sql.DB, user_submission_id int) (int, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	uhassubmission, err := models.FindUserSubmission(context.Background(), db, user_submission_id)
	if err != nil {
		return 0, err
	}
	_, err = uhassubmission.Delete(context.Background(), db, false)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}

	return uhassubmission.ID, nil
}
