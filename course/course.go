package course

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	coursematerial "learningbay24.de/backend/courseMaterial"
	"learningbay24.de/backend/dbi"
	"learningbay24.de/backend/errs"
	"learningbay24.de/backend/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
		return 0, errs.ErrEmptyName
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
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return 0, err
	}

	// Creates a Course struct
	c := &models.Course{Name: name, Description: description, EnrollKey: enrollkey, ForumID: f.ID}
	// Inserts into database
	err = c.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return 0, err
	} else {
		// TODO: Implement roles assigment for tutors
		// Gives the user with the ID in the 0 place in the array the role of the creator
		shasc := models.UserHasCourse{UserID: usersid, CourseID: c.ID, RoleID: dbi.CourseAdminRoleId}
		err = shasc.Insert(context.Background(), tx, boil.Infer())
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
			}

			return 0, err
		}
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
	}
	return c.ID, nil
}

// UpdateCourse takes the ID of a existing course and the already existing fields for name,enrollkey and description and overwrites the corespoding course and forum with the new Strings(name,enrollkey and description)
func EditCourse(db *sql.DB, id int, name string, description null.String, enrollkey string) (int, error) {
	// Validation
	if name == "" {
		return 0, errs.ErrEmptyName
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	c, err := models.FindCourse(context.Background(), tx, id)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return 0, err
	}
	c.EnrollKey = enrollkey
	c.Description = description
	c.Name = name

	_, err = c.Update(context.Background(), tx, boil.Infer())

	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return 0, err
	}

	f, err := models.FindForum(context.Background(), tx, id)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return 0, err
	}
	f.Name = name

	_, err = f.Update(context.Background(), tx, boil.Infer())

	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
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
		return 0, errs.ErrCourseNotEmpty
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
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return 0, err
	}
	f, err := models.FindForum(context.Background(), tx, id)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
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
				return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
			}

			return 0, err
		}
		_, err = f.Delete(context.Background(), tx, false)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
			}

			return 0, err
		}

		err = coursematerial.DeleteAllMaterialsFromCourse(db, id, false)
		if e := tx.Commit(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}
		return c.ID, nil

	}

	_, err = c.Delete(context.Background(), tx, true)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}
		return 0, err
	}

	_, err = f.Delete(context.Background(), tx, true)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return 0, err
	}
	if e := tx.Commit(); e != nil {
		return 0, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
	}
	return c.ID, nil
}

// GetEnrolledCoursesFromUser takes the ID of a User and returns a slice of Courses in which he is enrolled
func GetEnrolledCoursesFromUser(db *sql.DB, uid int) ([]*models.Course, error) {

	courses, err := models.Courses(
		qm.Select(models.CourseColumns.ID, models.CourseColumns.Name, models.CourseColumns.Description, models.CourseColumns.ForumID, "course.created_at", "course.updated_at"),
		qm.From(models.TableNames.UserHasCourse),
		qm.Where("user_has_course.user_id=?", uid),
		qm.And("user_has_course.course_id = course.id"),
		qm.And("user_has_course.role_id = ?", dbi.CourseUserRoleId),
		qm.Or("user_has_course.user_id=?", uid),
		qm.And("user_has_course.course_id = course.id"),
		qm.And("user_has_course.role_id = ?", dbi.CourseModeratorRoleId),
	).All(context.Background(), db)
	if err != nil {
		return nil, err
	}

	return courses, nil
}

// GetCoursesFromUser takes the ID of a User and returns a slice of Courses in which he is enrolled
func GetCreatedCoursesFromUser(db *sql.DB, uid int) ([]*models.Course, error) {

	courses, err := models.Courses(
		qm.Select(models.CourseColumns.ID, models.CourseColumns.Name, models.CourseColumns.Description, models.CourseColumns.ForumID, "course.created_at", "course.updated_at"),
		qm.From(models.TableNames.UserHasCourse),
		qm.Where("user_has_course.user_id=?", uid),
		qm.And("user_has_course.course_id = course.id"),
		qm.And("user_has_course.role_id = ?", dbi.CourseAdminRoleId),
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
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return err
	}

	_, err = userhascourse.Delete(context.Background(), tx, false)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return err
	}

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
	}
	return nil
}

// EnrollUser takes a UserID, CourseID and Enrollkey and adds the User to the course if the enrollkey is correct
func EnrollUser(db *sql.DB, uid int, cid int, enrollkey string) (*models.User, error) {

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {

		return nil, err
	}

	c, err := models.FindCourse(context.Background(), db, cid)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return nil, err
	}
	if c.EnrollKey != enrollkey {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}
		fmt.Println("enrollkey:", enrollkey, "expected:", c.EnrollKey)
		return nil, errs.ErrWrongEnrollkey

	}
	u, err := models.FindUser(context.Background(), db, uid)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return nil, err
	}

	var uhex models.UserHasCourseSlice
	// first check if relation already exists in the database and either insert a new row or reset deleted_at
	err = queries.Raw("select * from user_has_course where course_id=? AND user_id=?", cid, uid).Bind(context.Background(), db, &uhex)
	if err != nil {
		return nil, err
	}
	if len(uhex) > 0 {
		uhex[0].DeletedAt = null.TimeFromPtr(nil)
		_, err = uhex[0].Update(context.Background(), db, boil.Infer())
		if err != nil {
			return nil, err
		}
		return u, nil
	}

	userhascourse := models.UserHasCourse{UserID: uid, CourseID: cid, RoleID: dbi.CourseUserRoleId}
	err = userhascourse.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return nil, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return nil, err
	}

	if e := tx.Commit(); e != nil {
		return nil, fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
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
		qm.Select(models.CourseColumns.ID, models.CourseColumns.Name, models.CourseColumns.Description, models.CourseColumns.ForumID, models.CourseColumns.CreatedAt, models.CourseColumns.UpdatedAt),
		qm.Where(models.CourseColumns.Name+" LIKE ?", searchterm),
		qm.Or(models.CourseColumns.Description+" LIKE ?", searchterm),
	).All(context.Background(), db)
	if err != nil {
		// avoid no courses available being a `404`
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.Course{}, nil
		}

		return nil, err
	}
	return courses, nil
}
