package dbi

import (
	"context"
	"database/sql"
	"fmt"

	"learningbay24.de/backend/models"

	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

// Create a user with a given password as []byte.
// the cleartext password received will be hashed in this function.
func CreateUser(db *sql.DB, user models.User) (int, error) {
	// input validation is done on the database level
	// error is being thrown when something cannot be inserted

	password, err := bcrypt.GenerateFromPassword(user.Password, bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	user.Password = password

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	user.ID = 0
	err = user.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	return user.ID, nil
}

// Verify if the given cleartext password matches the saved password in the database for the user with the given email.
// Passwords in the database are always saved as a hash.
// Returns userId and nil on success, or an error on failure.
func VerifyCredentials(db *sql.DB, email string, password []byte) (int, error) {
	user, err := models.Users(qm.Where("email = ?", email)).One(context.Background(), db)
	if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(user.Password, password)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

// Recursively delete a user with their id.
// This doesn't delete forum entries, certificates or exams.
func DeleteUser(db *sql.DB, id int) error {
	flog := log.WithFields(log.Fields{
		"context": "user_deletion",
	})

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		flog.Errorf("Unable to create transaction: %s", err.Error())
		return err
	}

	us, err := models.UserSubmissions(models.UserSubmissionWhere.SubmitterID.EQ(id)).DeleteAll(context.Background(), tx, false)
	if err != nil {
		flog.Errorf("Unable to delete user submissions: %s", err.Error())
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
	}
	flog.Infof("Deleted %d entries from user_submission", us)

	f, err := models.Files(models.FileWhere.UploaderID.EQ(id)).DeleteAll(context.Background(), tx, false)
	if err != nil {
		flog.Errorf("Unable to delete files: %s", err.Error())
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
	}
	flog.Infof("Deleted %d entries from file", f)

	// don't delete forum_entry

	notif, err := models.Notifications(models.NotificationWhere.UserToID.EQ(id)).DeleteAll(context.Background(), tx, false)
	if err != nil {
		flog.Errorf("Unable to delete notifications: %s", err.Error())
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
	}
	flog.Infof("Deleted %d entries from notification", notif)

	uhc, err := models.UserHasCourses(models.UserHasCourseWhere.UserID.EQ(id)).DeleteAll(context.Background(), tx, false)
	if err != nil {
		flog.Errorf("Unable to delete user_has_courses: %s", err.Error())
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
	}
	flog.Infof("Deleted %d entries from user_has_course", uhc)

	// TODO: user_has_field_of_study?

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("unable to commit transaction: %s", err)
	}

	return nil
}
