package dbi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"learningbay24.de/backend/models"

	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

func AddDefaultData(db *sql.DB) error {
	password, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Unable to create password for admin user. Skipping inserting default and dummy data.")
	}
	admin := models.User{ID: 1, Firstname: "Admin", Surname: "Admin", Email: "admin@learningbay24.de", Password: password, RoleID: AdminRoleId, PreferredLanguageID: 9999}
	if err := admin.Insert(context.Background(), db, boil.Infer()); err != nil {
		return errors.New("Unable to insert admin user. Skipping inserting default data.")
	}

	return nil
}

func AddDummyData(db *sql.DB) {
	log.Info("Populating database with dummy data.")
	log.Info("If insertion fails because of duplicate entries, this is normal as the database is already populated with dummy data")
	// populate database with dummy data
	password, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Unable to create password for dummy data. Skipping inserting dummy data.")
		return
	}

	course := models.Course{ID: 9999, Name: "dummy course", Description: null.NewString("dummy course description", true), EnrollKey: "", ForumID: 9999}
	directory := models.Directory{ID: 9999, Name: "dummy directory", CourseID: 9999}
	exam := models.Exam{ID: 9999, Name: "dummy exam", Description: "dummy exam description", Date: time.Date(2022, time.May, 12, 10, 45, 00, 00, time.UTC), Duration: 5400, Online: 0, Location: null.NewString("dummy room 101", true), CourseID: 9999, CreatorID: 9999}
	fos := models.FieldOfStudy{ID: 9999, Name: null.NewString("dummy field of study", true), Semesters: null.NewInt(6, true)}
	forum := models.Forum{ID: 9999, Name: "dummy forum"}
	forum_entry := models.ForumEntry{ID: 9999, Subject: "dummy forum entry", Content: "dummy forum entry content", AuthorID: 9999, ForumID: 9999}
	graduation_level := models.GraduationLevel{ID: 9999, GraduationLevel: "dummy graduation level", Level: 1}
	language := models.Language{ID: 9999, Name: "dummy language"}
	notification := models.Notification{ID: 9999, Title: "dummy notification", Body: null.NewString("dummy notification body", true), UserToID: 9999}
	role := models.Role{ID: 9999, Name: "dummy role", DisplayName: "dummy role"}
	submission := models.Submission{ID: 9999, Name: "dummy submission", CourseID: 9999, VisibleFrom: time.Date(2022, time.May, 12, 10, 45, 00, 00, time.UTC)}
	user := models.User{ID: 9999, Firstname: "dummy firstname", Surname: "dummy surname", Email: "dummy@email.com", Password: password, RoleID: 9999, PreferredLanguageID: 9999}
	user_has_course := models.UserHasCourse{UserID: 9999, CourseID: 9999, RoleID: 9999}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Error("Unable to start transaction. Skipping inserting dummy data.")
		return
	}

	err = forum.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy forum into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = course.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy course into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = directory.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy directory into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = language.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy language into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = role.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy role into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = user.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy user into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = exam.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy exam into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = fos.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy field of study into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = forum_entry.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy forum entry into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = graduation_level.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy graduation level into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = notification.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy notification into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = submission.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy submission into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = user_has_course.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		log.Errorf("Unable to insert dummy user_has_course into db: %s. Skipping inserting dummy data.", err.Error())
		e := tx.Rollback()
		if e != nil {
			log.Error("Unable to rollback changes from database, aborting insertion of dummy data")
		}

		return
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf("Unable to commit changes for dummy data: %s. Aborting insertion of dummy data\n", err.Error())
		return
	}

	log.Info("Inserted dummy data into database.")
}
