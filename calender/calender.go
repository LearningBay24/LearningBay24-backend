package calender

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

func AppointmentInCalender(db *sql.DB, apId int, date time.Time, location null.String, online int8, courseId int, repeats bool, repeatDistance int, repeatEnd time.Time) error {

	// Begins the transaction (got from course.go)
	transaction, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Creating new appointment for that course
	newAppoint := &models.Appointment{ID: apId, Date: date, Location: location, Online: online, CourseID: courseId}

	// Inserts into database
	err = newAppoint.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		transaction.Rollback()
		return err
	}

	var nextDate time.Time = date
	if repeats {
		// Go through the calendar at the given interval, insert an appointment
		for {
			// Get Course object, insert appointment with AddAppointments:
			course, err := models.FindCourse(context.Background(), db, courseId)
			if err != nil {
				transaction.Rollback()
				return err
			}
			course.AddAppointments(context.Background(), db, true, newAppoint)

			// go to next appointment
			switch repeatDistance {
			case 1:
				nextDate.AddDate(0, 0, 7) // add seven days
			case 2:
				nextDate.AddDate(0, 1, 0) // add one month
			case 3:
				nextDate.AddDate(1, 0, 0) // add a year
			default:
				nextDate.AddDate(0, 0, 0)
			}
			if nextDate.After(repeatEnd) {
				break // stop, when the end date is reached
			}
		}
	} else {
		course, err := models.FindCourse(context.Background(), db, courseId)
		if err != nil {
			transaction.Rollback()
			return err
		}
		course.AddAppointments(context.Background(), db, true, newAppoint)
	}
	transaction.Commit()
	return nil
}

func SubmissionInCalender(db *sql.DB) error {

	// Begins transaction
	transaction, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// (submission = erstellte Abgabe)
	// TODO
	// 1. F.A.'s dazu raus ziehen
	// 2. PSeudocode
	// 3. Go-Code

	// End transaction
	transaction.Commit()
	return nil
}
