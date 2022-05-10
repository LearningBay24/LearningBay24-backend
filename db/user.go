package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"learningbay24.de/backend/models"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

func validateUserFields(user *models.User) error {
	if user.Title.Valid && len(user.Title.String) > 64 {
		return errors.New("Field \"title\" is too long, only 64 characters allowed")
	}

	if len(user.Firstname) > 32 {
		return errors.New("Field \"firstname\" is too long, only 32 characters allowed")
	}

	if len(user.Surname) > 32 {
		return errors.New("Field \"surname\" is too long, only 32 characters allowed")
	}

	if len(user.Email) > 256 {
		return errors.New("Field \"email\" is too long, only 256 characters allowed")
	}

	if user.PhoneNumber.Valid && len(user.PhoneNumber.String) > 45 {
		return errors.New("Field \"phone_number\" is too long, only 45 characters allowed")
	}

	if user.Residence.Valid && len(user.Residence.String) > 256 {
		return errors.New("Field \"residence\" is too long, only 256 characters allowed")
	}

	if user.Biography.Valid && len(user.Biography.String) > 512 {
		return errors.New("Field \"biography\" is too long, only 512 characters allowed")
	}

	return nil
}

// Create a user with a given password as []byte.
// the cleartext password received will be hashed in this function.
func CreateUser(db *sql.DB, user models.User) (int, error) {
	err := validateUserFields(&user)
	if err != nil {
		return 0, err
	}

	password, err := bcrypt.GenerateFromPassword(user.Password, bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	user.Password = password

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	// TODO: logic?
	// TODO: role_id, etc.? (foreign keys)
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

	return 0, nil
}

// Verify if the given cleartext password matches the saved password in the database for the user with the given email.
// Passwords in the database are always saved as a hash.
// Returns nil on success, or an error on failure.
func VerifyCredentials(db *sql.DB, email string, password []byte) error {
	user, err := models.Users(qm.From(models.TableNames.User), qm.Where("email = ?", email)).One(context.Background(), db)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("Unable to find user with E-Mail: %s", email)
	}

	err = bcrypt.CompareHashAndPassword(user.Password, password)
	if err != nil {
		return err
	}

	return nil
}
