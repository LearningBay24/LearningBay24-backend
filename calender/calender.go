package calender

// TODO:
// FA470: Appointment erst ab einem gegebenen Datum anezeigen lassen -> submission.visibleFrom
//	"Checkbox, ob der Kursteilnehmer die Abgabe erst ab einem bestimmten Zeitpunkt sieht"
//	"Wenn vorherige Checkbox checked ist: Zeitfeld und Uhrzeitfeld für Sichtbarkeitsdatum für Kursteilnehmer"
// FA330: Gelangen zu Abgabe über den Kalender:
//	"Endnutzer sieht auf seinem Dashboard und über den Navigationspunkt „Stundenplan“ seinen Kalender"
//	"Durch Drücken auf einer der Deadlines für seine Abgaben,
//		öffnet sich für den Endnutzer ein Formular zu dieser Abgabe mit dazugehörigen Informationen (siehe andere F.A.)"

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

// Returns all appointments the user with the user-ID has
func GetAllAppointments(db *sql.DB, userId int) ([]*models.Appointment, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	user, err := models.FindUser(context.Background(), tx, userId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return nil, err
	}
	courseQuery := user.UserHasCourses()
	courseSlice, err := courseQuery.All(context.Background(), tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", e, e)
		}
		return nil, err
	}
	var allAppointments []*models.Appointment

	// Collect all appointments from all courses of the user
	for _, course := range courseSlice {
		appointmentQuery := course.R.Course.Appointments()
		appointmentSlice, err := appointmentQuery.All(context.Background(), tx)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", e, e)
			}
			return nil, err
		}
		for _, appointment := range appointmentSlice {
			allAppointments = append(allAppointments, appointment)
		}
	}

	// TODO: es werden für den user noch die Submissions und Exams geholt und returned
	// 1. user.exams holen
	// 2. user.course.submissions holen

	if e := tx.Commit(); e != nil {
		return nil, fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
	return allAppointments, nil
}

// Returns all appointments the user with the user-ID has, between the beforeDate and afterDate
func GetAppointments(db *sql.DB, userId int, beforeDate time.Time, afterDate time.Time) ([]*models.Appointment, error) {

	if beforeDate.After(afterDate) {
		return nil, fmt.Errorf("calender: incorrect parameter usage")
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	user, err := models.FindUser(context.Background(), tx, userId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return nil, err
	}
	courseQuery := user.UserHasCourses()
	courseSlice, err := courseQuery.All(context.Background(), tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", e, e)
		}
		return nil, err
	}
	var allAppointments []*models.Appointment

	// Collect all appointments from all courses of the user
	for _, course := range courseSlice {
		appointmentQuery := course.R.Course.Appointments()
		appointmentSlice, err := appointmentQuery.All(context.Background(), tx)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", e, e)
			}
			return nil, err
		}
		for _, appointment := range appointmentSlice {
			if afterDate.After(appointment.Date) && beforeDate.Before(appointment.Date) {
				allAppointments = append(allAppointments, appointment)
			}
		}
	}

	if e := tx.Commit(); e != nil {
		return nil, fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
	return allAppointments, nil
}

// Returns the dates of all submissions the user with the user-ID has
func GetAllSubmissions(db *sql.DB, userId int) ([]*time.Time, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	// TODO - replace following lines with..
	// models.UserHasCourses(models.UserHasCourseWhere.UserID.EQ(userId)).All(context.Background(), tx)
	user, err := models.FindUser(context.Background(), tx, userId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return nil, err
	}
	courseQuery := user.UserHasCourses()
	courseSlice, err := courseQuery.All(context.Background(), tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", e, e)
		}
		return nil, err
	}
	var allSubmissions []*time.Time

	// Collect all appointments from all courses of the user
	for _, course := range courseSlice {
		submissionQuery := course.R.Course.Submissions()
		submissionSlice, err := submissionQuery.All(context.Background(), tx)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", e, e)
			}
			return nil, err
		}
		for j := 0; j < len(submissionSlice); j++ {
			allSubmissions = append(allSubmissions, &submissionSlice[j].Deadline.Time)
		}
	}

	if e := tx.Commit(); e != nil {
		return nil, fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
	return allSubmissions, nil
}

func AddCourseToCalender(db *sql.DB, date time.Time, location null.String, online int8, courseId int, repeats bool, repeatDistance int, repeatEnd time.Time) (int, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	newAppoint := &models.Appointment{Date: date, Location: location, Online: online, CourseID: courseId}

	err = newAppoint.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return 0, err
	}
	course, err := models.FindCourse(context.Background(), tx, courseId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return 0, err
	}
	course.AddAppointments(context.Background(), tx, true, newAppoint)

	var nextDate time.Time = date
	var checkDate time.Time
	if repeats {
		// Go through the calendar at the given interval, insert an appointment; Stop, when the end date is reached
		for !nextDate.After(repeatEnd) {
			newAppointAfter := &models.Appointment{Date: nextDate, Location: location, Online: online, CourseID: courseId}
			err = newAppointAfter.Insert(context.Background(), tx, boil.Infer())
			if err != nil {
				if e := tx.Rollback(); e != nil {
					return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
				}
				return 0, err
			}
			course.AddAppointments(context.Background(), tx, true, newAppointAfter)

			// go to next appointment
			switch repeatDistance {
			case 1:
				checkDate = nextDate.AddDate(0, 0, 7) // add seven days
			case 2:
				checkDate = nextDate.AddDate(0, 1, 0) // add one month
			case 3:
				checkDate = nextDate.AddDate(1, 0, 0) // add a year
			default:
				if e := tx.Rollback(); e != nil {
					return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
				}
			}

			if checkDate.Before(date) {
				return 0, fmt.Errorf("error when trying to create new appointments")
			}
		}
	} else {
		course, err := models.FindCourse(context.Background(), tx, courseId)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
			}
			return 0, err
		}
		course.AddAppointments(context.Background(), tx, true, newAppoint)
	}

	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
	return newAppoint.ID, nil
}

func DeactivateCourseInCalender(db *sql.DB, appointmentId int, courseId int, repeats bool) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// TODO
	// if repeats: how to get the other appointments?
	appointment, err := models.FindAppointment(context.Background(), tx, appointmentId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	appointment.Delete(context.Background(), db, false)

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

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
	return nil
}

func AddSubmissionToCalender(db *sql.DB, submDate time.Time, submName null.String, courseId int) (int, error) {
	// uses location from Appointment as description for the submission-name, used F.A.'s: 300, 470

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	// TODO -> get submission object with submission.id as parameter, set submission.date to given Date, don't create a new appointment
	// newer TODO: Don't set Date, read Date from submission and add to calender (calender API)

	newAppoint := &models.Appointment{Date: submDate, Location: submName, CourseID: courseId}
	err = newAppoint.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return 0, err
	}
	course, err := models.FindCourse(context.Background(), tx, courseId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return 0, err
	}
	course.AddAppointments(context.Background(), tx, true, newAppoint)

	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
	return newAppoint.ID, nil
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
	appointment.Delete(context.Background(), db, false)

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
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
	appointment.Delete(context.Background(), db, false)

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
	return nil
}

/*
func ChangeSubmissionDate(db *sql.DB, appointmentId int, courseId int, submDate time.Time, submName null.String, submId int) error {
	// used F.A.'s: 480

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// TODO: if realy no appointment created, delete this {
	appointment, err := models.FindAppointment(context.Background(), tx, appointmentId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	appointment.Date = submDate
	// }

	submission, err := models.FindSubmission(context.Background(), tx, submId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		return err
	}
	submission.Deadline = null.TimeFrom(submDate)

	// TODO: If there is conditional visibility: change date for visibility
	// 1. case: Change it automatically depending on the new date
	// 2. case: Change it to a given new visibility

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
	return nil
}
*/
