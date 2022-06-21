package calender

// TODO:
// FA470: Appointment erst ab einem gegebenen Datum anezeigen lassen -> submission.visibleFrom
//	"Checkbox, ob der Kursteilnehmer die Abgabe erst ab einem bestimmten Zeitpunkt sieht"
//	"Wenn vorherige Checkbox checked ist: Zeitfeld und Uhrzeitfeld für Sichtbarkeitsdatum für Kursteilnehmer"
// FA330: Gelangen zu Abgabe über den Kalender:
//	"Endnutzer sieht auf seinem Dashboard und über den Navigationspunkt „Stundenplan“ seinen Kalender"
//	"Durch Drücken auf einer der Deadlines für seine Abgaben,
//		öffnet sich für den Endnutzer ein Formular zu dieser Abgabe mit dazugehörigen Informationen (siehe andere F.A.)"
//
// json.Unmarshal() in api.go -> switch to BindJSON() (see AddCourseToCalender)

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/course"
	"learningbay24.de/backend/models"
)

type Calender interface {
	// TODO
}

type PublicController struct {
	Database *sql.DB
}

type RepeatDistance int

type AppointmentWithCourse struct {
	*models.Appointment `boil:",bind"`
	Name                string `boil:"name" json:"name" toml:"name" yaml:"name"`
}

const (
	None RepeatDistance = iota
	Week
	Month
	Year
)

// Returns all appointments the user with the user-ID has
func (p *PublicController) GetAllAppointments(userId int) ([]AppointmentWithCourse, error) {

	courses, err := course.GetCoursesFromUser(p.Database, userId)
	if err != nil {
		return nil, err
	}
	var fullSlice []AppointmentWithCourse
	for _, course := range courses {
		app, err := models.Appointments(models.AppointmentWhere.CourseID.EQ(course.ID)).All(context.Background(), p.Database)
		var currentAppointment AppointmentWithCourse
		for index := range app {
			currentAppointment.Appointment = app[index]
			currentAppointment.Name = course.Name
			fullSlice = append(fullSlice, currentAppointment)
		}

		if err != nil {
			return nil, err
		}
	}

	return fullSlice, nil
}

// adds appointment/s to the course; appointments may repeat
func (p *PublicController) AddCourseToCalender(date time.Time, duration int, location null.String, online int8, courseId int) (int, error) {
	tx, err := p.Database.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	newAppoint := &models.Appointment{Date: date, Location: location, Online: online, CourseID: courseId, Duration: duration}

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

func (p *PublicController) DeactivateCourseInCalender(appointmentId int) error {
	tx, err := p.Database.BeginTx(context.Background(), nil)
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
	appointment.Delete(context.Background(), p.Database, false)

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to Commit transaction on error: %s; %s", err.Error(), e.Error())
	}
	return nil
}
