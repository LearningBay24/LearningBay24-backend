package calender

// TODO - unsure:
// FA470: Appointment erst ab einem gegebenen Datum anezeigen lassen (?) wie umsetzen?
//	• Checkbox, ob der Kursteilnehmer die Abgabe erst ab einem bestimmten Zeitpunkt sieht
//	• Wenn vorherige Checkbox checked ist: Zeitfeld und Uhrzeitfeld für Sichtbarkeitsdatum für Kursteilnehmer
// -> submission.visibleFrom
// FA330: Gelangen zu Abgabe über den Kalender:
//	• Endnutzer sieht auf seinem Dashboard und über den Navigationspunkt „Stundenplan“ seinen Kalender
//	• Durch Drücken auf einer der Deadlines für seine Abgaben,
//		öffnet sich für den Endnutzer ein Formular zu dieser Abgabe mit dazugehörigen Informationen (siehe andere F.A.)

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

func AddCourseToCalender(db *sql.DB, appoId int, date time.Time, location null.String, online int8, courseId int, repeats bool, repeatDistance int, repeatEnd time.Time) error {

	// Begins the transaction (got from course.go)
	transaction, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Creating new appointment for that course
	newAppoint := &models.Appointment{ID: appoId, Date: date, Location: location, Online: online, CourseID: courseId}

	// Inserts into database
	err = newAppoint.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		transaction.Rollback()
		return err
	}

	var nextDate time.Time = date
	if repeats {
		// Go through the calendar at the given interval, insert an appointment; Stop, when the end date is reached
		for !nextDate.After(repeatEnd) {
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
		}
	} else {
		course, err := models.FindCourse(context.Background(), db, courseId)
		if err != nil {
			transaction.Rollback()
			return err
		}
		course.AddAppointments(context.Background(), db, true, newAppoint)
	}

	// End transaction
	transaction.Commit()
	return nil
}

func AddSubmissionToCalender(db *sql.DB, appoId int, submDate time.Time, submName null.String, courseId int) error {
	// uses location as description for the submission-name, used F.A.'s: 300, 470

	// Begins transaction
	transaction, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Creating new appointment for that course
	newAppoint := &models.Appointment{ID: appoId, Date: submDate, Location: submName, CourseID: courseId}

	// Inserts into database
	err = newAppoint.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		transaction.Rollback()
		return err
	}

	// Add to course
	course, err := models.FindCourse(context.Background(), db, courseId)
	if err != nil {
		transaction.Rollback()
		return err
	}
	course.AddAppointments(context.Background(), db, true, newAppoint)

	// End transaction
	transaction.Commit()
	return nil
}

func AddExamToCalender(db *sql.DB, appoId int, examDate time.Time, location null.String, online int8, examId int) error {
	// used FA300, TODO -> store exam name in which appointment parameter?
	// Begins transaction
	transaction, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Creating new appointment for that course
	newAppoint := &models.Appointment{ID: appoId, Date: examDate, Location: location, Online: online}

	// Inserts into database
	err = newAppoint.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		transaction.Rollback()
		return err
	}

	// TODO -> add appointment to exam, as it is possible to add an appointment to a course (examId):
	/*
		course, err := models.FindCourse(context.Background(), db, courseId)
		if err != nil {
			transaction.Rollback()
			return err
		}
		course.AddAppointments(context.Background(), db, true, newAppoint)
	*/

	// End transaction
	transaction.Commit()
	return nil
}

func ChangeSubmissionDate(db *sql.DB, appointmentId int, courseId int, submDate time.Time, submName null.String) error {
	// used F.A.'s: 480
	// Begins transaction
	transaction, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Get appointment element
	appointment, err := models.FindAppointment(context.Background(), db, appointmentId)
	if err != nil {
		transaction.Rollback()
		return err
	}
	// Change the Date -> enough?
	appointment.Date = submDate

	// TODO: If there is conditional visibility: change date for visibility
	// 1. case: Change it automatically depending on the new date
	// 2. case: Change it to a given new visibility

	// End transaction
	transaction.Commit()
	return nil
}

func InactivateAppointment(db *sql.DB, appointmentId int) error {
	// Begins transaction
	transaction, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// Get appointment element
	appointment, err := models.FindAppointment(context.Background(), db, appointmentId)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// Deleted at
	currentDate := time.Now()
	appointment.DeletedAt = null.TimeFrom(currentDate)

	// Delete in Database -> necessary?
	/*
		_, errDel := appointment.Delete(context.Background(), db, false)
		if errDel != nil {
			transaction.Rollback()
			return errDel
		}
	*/

	// End transaction
	transaction.Commit()
	return nil
}
