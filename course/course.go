package course

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	coursematerial "learningbay24.de/backend/courseMaterial"
	"learningbay24.de/backend/dbi"
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
func CreateCourse(db *sql.DB, name string, description null.String, enrollkey string, usersid int) (int, error) {
	// Validation
	if name == "" {
		return 0, errors.New("name cant be empty")
	}
	// Begins the transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	// Creates a Forum struct (Forum has to be created first because of Foreign Key)
	f := &models.Forum{Name: name}
	// Inserts into database
	err = f.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	// Creates a Course struct
	c := &models.Course{Name: name, Description: description, EnrollKey: enrollkey, ForumID: f.ID}
	// Inserts into database
	err = c.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	} else {
		// TODO: Implement roles assigment for tutors
		// TODO: remove hard coded role
		// Gives the user with the ID in the 0 place in the array the role of the creator
		shasc := models.UserHasCourse{UserID: usersid, CourseID: c.ID, RoleID: dbi.CourseAdminRoleId}
		err = shasc.Insert(context.Background(), tx, boil.Infer())
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
			}

			return 0, err
		}
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
	}
	return c.ID, nil
}

// UpdateCourse takes the ID of a existing course and the already existing fields for name,enrollkey and description and overwrites the corespoding course and forum with the new Strings(name,enrollkey and description)
func EditCourse(db *sql.DB, id int, name string, description null.String, enrollkey string) (int, error) {
	// Validation
	if name == "" {
		return 0, errors.New("name cant be empty")
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	c, err := models.FindCourse(context.Background(), tx, id)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}
	c.EnrollKey = enrollkey
	c.Description = description
	c.Name = name

	_, err = c.Update(context.Background(), tx, boil.Infer())

	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	f, err := models.FindForum(context.Background(), tx, id)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}
	f.Name = name

	_, err = f.Update(context.Background(), tx, boil.Infer())

	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return c.ID, nil
}

// DeleteCourse takes a ID and deletes the course and the forum associated with it
func DeleteCourse(db *sql.DB, id int) (int, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	// Check if there are more users in the course besides the creator
	userhascourse, err := models.UserHasCourses(models.UserHasCourseWhere.CourseID.EQ(id)).Count(context.Background(), db)
	if err != nil {
		return 0, err
	}
	if userhascourse > 1 {
		return 0, errors.New("there are still people enrolled in the course besides the creator")
	}
	// Get the creator of the course
	userinc, err := GetUsersInCourse(db, id)
	if err != nil {
		return 0, err
	}
	// Its just creator in the course so delete him
	err = DeleteUserFromCourse(db, userinc[0].ID, id)
	if err != nil {
		return 0, err
	}
	c, err := models.FindCourse(context.Background(), tx, id)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}
	f, err := models.FindForum(context.Background(), tx, id)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	// Checks if more than 10 Minutes have passed will softdelete if thats the case
	curTime := time.Now()
	diff := curTime.Sub(c.CreatedAt.Time)
	if diff.Minutes() < 10 {
		_, err = c.Delete(context.Background(), tx, false)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
			}

			return 0, err
		}
		_, err = f.Delete(context.Background(), tx, false)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
			}

			return 0, err
		}

		err = coursematerial.DeleteAllMaterialsFromCourse(db, id, false)
		if e := tx.Commit(); e != nil {
			return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
		}
		return c.ID, nil

	}

	_, err = c.Delete(context.Background(), tx, true)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
		return 0, err
	}

	_, err = f.Delete(context.Background(), tx, true)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return c.ID, nil
}

// GetCoursesFromUser takes the ID of a User and returns a slice of Courses in which he is enrolled
func GetCoursesFromUser(db *sql.DB, uid int) ([]*models.Course, error) {

	courses, err := models.Courses(
		qm.From(models.TableNames.UserHasCourse),
		qm.Where("user_has_course.user_id=?", uid),
		qm.And("user_has_course.course_id = course.id"),
	).All(context.Background(), db)
	if err != nil {
		return nil, err
	}

	return courses, nil
}

// GetUserCourses takes the ID of a Course and returns a slice of Users which are enrolled in it
func GetUsersInCourse(db *sql.DB, cid int) ([]*models.User, error) {

	users, err := models.Users(
		qm.From(models.TableNames.UserHasCourse),
		qm.Where("user_has_course.course_id=?", cid),
		qm.And("user_has_course.user_id = user.id"),
	).All(context.Background(), db)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteUserFromCourse takes a UserID and a CourseID and deletes the corresponding entry in the table "user_has_course"
func DeleteUserFromCourse(db *sql.DB, uid int, cid int) error {

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {

		return err
	}

	userhascourse, err := models.FindUserHasCourse(context.Background(), tx, uid, cid)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}

		return err
	}

	_, err = userhascourse.Delete(context.Background(), tx, false)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}

		return err
	}

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return nil
}

// EnrollUser takes a UserID, CourseID and Enrollkey and adds the User to the course if the enrollkey is correct
func EnrollUser(db *sql.DB, uid int, cid int, enrollkey string) (*models.User, error) {

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {

		return nil, err
	}

	c, err := models.FindCourse(context.Background(), tx, cid)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}

		return nil, err
	}
	if c.EnrollKey != enrollkey {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}
		fmt.Println("enrollkey:", enrollkey, "expected:", c.EnrollKey)
		return nil, errors.New("wrong Enrollkey")

	}
	userhascourse := models.UserHasCourse{UserID: uid, CourseID: cid, RoleID: dbi.CourseUserRoleId}
	err = userhascourse.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}

		return nil, err
	}
	u, err := models.FindUser(context.Background(), tx, uid)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}

		return nil, err
	}
	if e := tx.Commit(); e != nil {
		return nil, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return u, nil
}

func GetCourseRole(db *sql.DB, user_id int, course_id int) (int, error) {

	userhascourse, err := models.FindUserHasCourse(context.Background(), db, user_id, course_id)
	if err != nil {
		return 0, err
	}
	return userhascourse.RoleID, nil
}

// Search in course_name and course_descripton for the searchterm
func SearchCourse(db *sql.DB, searchterm string) ([]*models.Course, error) {
	searchterm = "%" + searchterm + "%"
	courses, err := models.Courses(
		qm.Where(models.CourseColumns.Name+" LIKE ?", searchterm),
		qm.Or(models.CourseColumns.Description+" LIKE ?", searchterm),
	).All(context.Background(), db)
	if err != nil {
		return nil, err
	}
	return courses, nil
}
