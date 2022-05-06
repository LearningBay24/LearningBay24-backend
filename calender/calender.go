package calender

// TODO - unsure:
// FA470: Appointment erst ab einem gegebenen Datum anezeigen lassen -> submission.visibleFrom
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
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

func AddCourseToCalender(db *sql.DB, appointmentId int, date time.Time, location null.String, online int8, courseId int, repeats bool, repeatDistance int, repeatEnd time.Time) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	newAppoint := &models.Appointment{ID: appointmentId, Date: date, Location: location, Online: online, CourseID: courseId}
	err = newAppoint.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}

	var nextDate time.Time = date
	if repeats {
		// Go through the calendar at the given interval, insert an appointment; Stop, when the end date is reached
		for !nextDate.After(repeatEnd) {
			// Get Course object, insert appointment with AddAppointments:
			course, err := models.FindCourse(context.Background(), tx, courseId)
			if err != nil {
				if e := tx.Rollback(); e != nil {
					return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
				}
				return err
			}
			course.AddAppointments(context.Background(), tx, true, newAppoint)

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
		course, err := models.FindCourse(context.Background(), tx, courseId)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
			}
			return err
		}
		course.AddAppointments(context.Background(), tx, true, newAppoint)
	}

	tx.Commit()
	return nil
}

func DeactivateCourseInCalender(db *sql.DB, appointmentId int, courseId int, repeats bool) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// TODO
	// repeats: if repeats: delete it more than once? All repeating appointments have same id -> change?
	appointment, err := models.FindAppointment(context.Background(), tx, appointmentId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	currentDate := time.Now()
	appointment.DeletedAt = null.TimeFrom(currentDate)

	/* TODO - deactivate Appointment in course?
	course, err := models.FindCourse(context.Background(), tx, courseId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	course.Appointments()...
	*/

	tx.Commit()
	return nil
}

func AddSubmissionToCalender(db *sql.DB, appointmentId int, submDate time.Time, submName null.String, courseId int) error {
	// uses location from Appointment as description for the submission-name, used F.A.'s: 300, 470

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	newAppoint := &models.Appointment{ID: appointmentId, Date: submDate, Location: submName, CourseID: courseId}
	err = newAppoint.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}

	course, err := models.FindCourse(context.Background(), tx, courseId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	course.AddAppointments(context.Background(), tx, true, newAppoint)

	tx.Commit()
	return nil
}

func DeactivateSubmissionInCalender(db *sql.DB, appointmentId int, courseId int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	appointment, err := models.FindAppointment(context.Background(), tx, appointmentId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	currentDate := time.Now()
	appointment.DeletedAt = null.TimeFrom(currentDate)

	/* TODO - deactivate Appointment in course?
	course, err := models.FindCourse(context.Background(), tx, courseId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	course.Appointments()...
	*/

	tx.Commit()
	return nil
}

func AddExamToCalender(db *sql.DB, appointmentId int, examDate time.Time, location null.String, online int8, examId int) error {
	// used FA300, TODO -> store exam name in which appointment parameter?

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	newAppoint := &models.Appointment{ID: appointmentId, Date: examDate, Location: location, Online: online}
	err = newAppoint.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}

	// TODO -> add appointment to exam, as it is possible to add an appointment to a course (examId):
	/*
		course, err := models.FindCourse(context.Background(), tx, courseId)
		if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
			return err
		}
		course.AddAppointments(context.Background(), tx, true, newAppoint)
	*/

	tx.Commit()
	return nil
}

func DeactivateExamInCalender(db *sql.DB, appointmentId int, examId int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	appointment, err := models.FindAppointment(context.Background(), tx, appointmentId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	currentDate := time.Now()
	appointment.DeletedAt = null.TimeFrom(currentDate)

	/* TODO - deactivate Appointment in exam, as it is possible in course (examId)?
	course, err := models.FindCourse(context.Background(), tx, courseId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	course.Appointments()...
	*/

	tx.Commit()
	return nil
}

func ChangeSubmissionDate(db *sql.DB, appointmentId int, courseId int, submDate time.Time, submName null.String) error {
	// used F.A.'s: 480

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	appointment, err := models.FindAppointment(context.Background(), tx, appointmentId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	// Change the Date -> enough?
	appointment.Date = submDate

	// TODO: If there is conditional visibility: change date for visibility
	// 1. case: Change it automatically depending on the new date
	// 2. case: Change it to a given new visibility

	tx.Commit()
	return nil
}

func DeactivateAppointment(db *sql.DB, appointmentId int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	appointment, err := models.FindAppointment(context.Background(), tx, appointmentId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	currentDate := time.Now()
	appointment.DeletedAt = null.TimeFrom(currentDate)

	tx.Commit()
	return nil
}
