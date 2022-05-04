package calender

// TODO - unsure:
// FA470: Appointment erst ab einem gegebenen Datum anezeigen lassen (?) wie umsetzen:
//	• Checkbox, ob der Kursteilnehmer die Abgabe erst ab einem bestimmten Zeitpunkt sieht
//	• Wenn vorherige Checkbox checked ist: Zeitfeld und Uhrzeitfeld für Sichtbarkeitsdatum für Kursteilnehmer
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

func AddAppointmentToCalender(db *sql.DB, appoId int, date time.Time, location null.String, online int8, courseId int, repeats bool, repeatDistance int, repeatEnd time.Time) error {

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

func ChangeSubmissionDate(db *sql.DB, courseId int, submDate time.Time, submName null.String) error {

	// Begins transaction
	transaction, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	// TODO
	// used F.A.'s: 480

	/*
		/FA480/ Nutzer: Abgaben bearbeiten > Fälligkeitsdatum-/Uhrzeit anpassen:
			• Der Ersteller des Kurses kann in der Kursansicht über den Toggle „bearbeiten“ in den Bearbeitungs-Zustand wechseln, oder sieht in der Abgabenübersicht neben allen seinen Abgaben den Button „Bearbeiten“
			• Buttons „Bearbeiten“ neben jeder der Abgaben in der Kursansicht: Erscheinen eines Formulars zum Bearbeiten der Abgabe
			• Buttons „Bearbeiten“ in der Abgabenübersicht: Erscheinen eines Formulars zum Bearbeiten der Abgabe
			• Zeitfeld für Abgabendatum, default = bisher eingestelltes Datum
			• Uhrzeitfeld für Abgabenuhrzeit, default = bisher eingestellte Uhrzeit
			• Checkbox, ob der Abstand zwischen Abgabezeitpunkt und Sichtbarkeitszeitpunkt für Kursteilnehmer gleich bleiben soll,
				default = checked
			• Wenn vorherige Checkbox unchecked ist: Erscheinen von Zeitfeld und Uhrzeitfeld für den Zeitpunkt der Sichtbarkeit für Kursteilnehmer
			• Zeitfeld für den Zeitpunkt der Sichtbarkeit für Kursteilnehmer: default = bisheriges Datum,
				wenn bisheriges Datum nach dem neuen Abgabedatum liegt: default = Abgabedatum
			• Uhrzeitfeld für den Zeitpunkt der Sichtbarkeit für Kursteilnehmer: default = bisherige Uhrzeit,
				wenn bisheriges Datum nach dem neuen Abgabedatum liegt: default = Abgabeuhrzeit
			• Zeitfeld und Uhrzeitfeld für den Zeitpunkt der Sichtbarkeit für Kursteilnehmer können nicht
				zeitlich nach dem Abgabezeitpunkt eingestellt werden
			• Button „OK“: Übernimmt die Informationen, schließt das Formular
			• Button „Abbrechen“: Schließt das Formular
	*/

	// End transaction
	transaction.Commit()
	return nil
}
