package calender

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

func AppointmentInCalender(db *sql.DB, id int, date time.Time, location null.String, online int8, courseId int, repeats bool, repeatDistance int, repeatEnd time.Time, usersid []int) error {

	// Begins the transaction (got from course.go)
	transaction, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Creating new appointment for that course
	newAppoint := &models.Appointment{ID: id, Date: date, Location: location, Online: online, CourseID: courseId}

	// Inserts into database
	err = newAppoint.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		transaction.Rollback()
		return err
	}

	// add new appointment to each person in the course
	// TODO -> per userId, oder muss aus der CourseId der Kurs und die dazugehörigen User gezogen werden?
	for i := 0; i < len(usersid); i++ {

		var nextDate time.Time = date
		if repeats {

			// Im gegebenen Abstand den jeweiligen Kalender durchgehen, Termin wie bei else einfügen
			for {
				// courseMembers[i]->calender.insert(appointment)

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
			// courseMembers[i]->calender.insert(appointment)
		}
	}
	transaction.Commit()
	return nil
}
