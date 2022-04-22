package course

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

// GetCourse takes a ID and returns a struct of the course with this ID
func GetCourse(db *sql.DB, id int) (*models.Course, error) {

	c, err := models.FindCourse(context.Background(), db, id)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// CreateCourse takes a name,enrollkey and description and adds a course and forum with that Name in the Database while userid is an array of IDs that is used to assign the role of the creator
// and the roles for tutor
func CreateCourse(db *sql.DB, name string, description null.String, enrollkey string, usersid []int) error {
	// TODO: implement check for certificates

	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	// Creates a Forum struct (Forum has to be created first because of Foreign Key)
	f := &models.Forum{Name: name}
	// Inserts into database
	err = f.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		tx.Rollback()
		return err
	}

	// Creates a Course struct
	c := &models.Course{Name: name, Description: description, EnrollKey: enrollkey, ForumID: f.ID}
	// Inserts into database
	err = c.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		tx.Rollback()
		return err
	} else {
		// TODO: Implement roles assigment for tutors
		// Gives the user with the ID in the 0 place in the array the role of the creator
		shasc := models.UserHasCourse{UserID: usersid[0], CourseID: c.ID, RoleID: 1}
		err = shasc.Insert(context.Background(), db, boil.Infer())
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

// UpdateCourse takes the ID of a existing course and the already existing fields for name,enrollkey and description and overwrites the corespoding course and forum with the new Strings(name,enrollkey and description)
func UpdateCourse(db *sql.DB, id int, name string, description null.String, enrollkey string) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	c, err := models.FindCourse(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	c.EnrollKey = enrollkey
	c.Description = description
	c.Name = name

	_, err = c.Update(context.Background(), db, boil.Infer())

	if err != nil {
		tx.Rollback()
		return err
	}

	f, err := models.FindForum(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	f.Name = name

	_, err = f.Update(context.Background(), db, boil.Infer())

	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// DeleteCourse takes a ID and deletes the course and the forum associated with it
func DeleteCourse(db *sql.DB, id int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	c, err := models.FindCourse(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Checks if more than 10 Minutes have passed wont delete if thats the case
	curTime := time.Now()
	diff := curTime.Sub(c.CreatedAt.Time)
	if diff.Minutes() < 10 {

		_, err = c.Delete(context.Background(), db, true)
		if err != nil {
			tx.Rollback()
			return err
		}

		f, err := models.FindForum(context.Background(), db, id)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = f.Delete(context.Background(), db, true)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		return errors.New("More than 10 Minutes have passed")
	}
	tx.Commit()
	return nil
}

// DeactivateCourse takes a ID and deactivates the course and the forum associated with it
func DeactivateCourse(db *sql.DB, id int) error {

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	c, err := models.FindCourse(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	c.DeletedAt = null.NewTime(time.Now(), true)
	_, err = c.Update(context.Background(), db, boil.Infer())

	if err != nil {
		tx.Rollback()
		return err
	}

	f, err := models.FindForum(context.Background(), db, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	f.DeletedAt = null.NewTime(time.Now(), true)
	_, err = f.Update(context.Background(), db, boil.Infer())

	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetUserCourses takes the ID of a User and returns a slice of Courses in which he is enrolled
func GetUserCourses(db *sql.DB, uid int) (models.CourseSlice, error) {

	usrhascourse, err := models.UserHasCourses(models.UserHasCourseWhere.UserID.EQ(uid)).All(context.Background(), db)
	if err != nil {

		return nil, err
	}

	var courses models.CourseSlice

	for num := 0; len(usrhascourse) > num; num++ {

		course, err := models.FindCourse(context.Background(), db, usrhascourse[num].CourseID)
		courses = append(courses, course)
		if err != nil {

			return nil, err
		}

	}

	return courses, nil
}

// GetUserCourses takes the ID of a Course and returns a slice of Users which are enrolled in it
func GetUsersinCourse(db *sql.DB, cid int) (models.UserSlice, error) {

	usrhascourse, err := models.UserHasCourses(models.UserHasCourseWhere.CourseID.EQ(cid)).All(context.Background(), db)
	if err != nil {

		return nil, err
	}

	var users models.UserSlice

	for num := 0; len(usrhascourse) > num; num++ {

		user, err := models.FindUser(context.Background(), db, usrhascourse[num].UserID)
		users = append(users, user)
		if err != nil {
			return nil, err
		}

	}

	return users, nil
}

// DeleteUserFromCourse takes a UserID and a CourseID and deletes the corresponding entry in the table "user_has_course"
func DeleteUserFromCourse(db *sql.DB, uid int, cid int) error {

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {

		return err
	}

	usrhascourse, err := models.FindUserHasCourse(context.Background(), db, uid, cid)
	if err != nil {
		tx.Rollback()
		return err
	}
	if usrhascourse.RoleID != 1 {
		_, err = usrhascourse.Delete(context.Background(), db, true)
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	} else {
		return errors.New("Trying to delete the creator of the course")
	}
}

// EnrollUser takes a UserID, CourseID and Enrollkey and adds the User to the course if the enrollkey is correct
func EnrollUser(db *sql.DB, uid int, cid int, enrollkey string) error {

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {

		return err
	}

	c, err := models.FindCourse(context.Background(), db, cid)
	if err != nil {
		tx.Rollback()
		return err
	}
	if c.EnrollKey == enrollkey {
		usrhascourse := models.UserHasCourse{UserID: uid, CourseID: cid, RoleID: 3}
		err = usrhascourse.Insert(context.Background(), db, boil.Infer())
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil

	} else {
		tx.Rollback()
		return errors.New("Wrong Enrollkey")
	}

}
