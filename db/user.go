package db

import (
	"context"
	"database/sql"
	"fmt"

	"learningbay24.de/backend/models"

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
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
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
