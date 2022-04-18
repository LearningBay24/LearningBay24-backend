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
	"learningbay24.de/backend/models"
)

// CreateCourse takes a name,enrollkey and description and adds a course and forum with that Name in the Database while userid is an array of IDs that is used to assign the role of the creator
// and the roles for tutor
func CreateCourse(db *sql.DB, name, enrollkey string, description null.String, usersid []int) error {
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
	c := &models.Course{Name: name, ForumID: f.ID, Description: description}
	// Inserts into database
	err = c.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		tx.Rollback()
		return err
	} else {
		// TODO: Implement roles assigment for tutors
		shasc := models.UserHasCourse{UserID: usersid[0], CourseID: c.ID, RoleID: 1}
		err = shasc.Insert(context.Background(), db, boil.Infer())
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return err
}

// UpdateCourse takes the ID of a existing course and the already existing fields for name,enrollkey and description and overwrites the corespoding course and forum with the new Strings(name,enrollkey and description)
func UpdateCourse(db *sql.DB, id int, name, enrollkey string, description null.String) error {
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

	cupdate, err := c.Update(context.Background(), db, boil.Infer())
	fmt.Println(cupdate)
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

	fupdate, err := f.Update(context.Background(), db, boil.Infer())
	fmt.Println(fupdate)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
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
	return err
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
	return err
}

//UpdateCourse takes the ID of a existing course and overwrites the corespoding course and forum with the new Strings(name,enrollkey and description)
func UpdateCourse(id int, name, enrollkey string, description null.String) {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/learningbay24?parseTime=true")

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Connected")
	}

	c, err := models.FindCourse(context.Background(), db, id)
	if err != nil {
		panic(err.Error())
	}

	c.EnrollKey = enrollkey
	c.Description = description
	c.Name = name

	cupdate, err := c.Update(context.Background(), db, boil.Infer())
	fmt.Println(cupdate)
	if err != nil {
		panic(err.Error())
	}

	f, err := models.FindForum(context.Background(), db, id)
	if err != nil {
		panic(err.Error())
	}
	f.Name = name

	fupdate, err := f.Update(context.Background(), db, boil.Infer())
	fmt.Println(fupdate)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()
}

//DeleteCourse takes a ID and deletes the course and the forum associated with it
func DeleteCourse(id int) {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/learningbay24?parseTime=true")

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Connected")
	}
	c, err := models.FindCourse(context.Background(), db, id)
	if err != nil {
		panic(err.Error())
	}

	//Checks if more than 10 Minutes have passed wont delete if thats the case
	curTime := time.Now()
	diff := curTime.Sub(c.CreatedAt.Time)
	if diff.Minutes() < 10 {
		fmt.Println("Only ", diff.Minutes(), "Have passed")
		cdel, err := c.Delete(context.Background(), db, true)
		fmt.Println(cdel)
		if err != nil {
			panic(err.Error())
		}

		f, err := models.FindForum(context.Background(), db, id)
		if err != nil {
			panic(err.Error())
		}

		fdel, err := f.Delete(context.Background(), db, true)
		fmt.Println(fdel)
		if err != nil {
			panic(err.Error())
		}
	} else {
		fmt.Println("More than 10 Minutes have passed")
	}
	defer db.Close()
}

//DeactivateCourse takes a ID and deactivates the course and the forum associated with it
func DeactivateCourse(id int) {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/learningbay24?parseTime=true")

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Connected")
	}
	c, err := models.FindCourse(context.Background(), db, id)
	if err != nil {
		panic(err.Error())
	}
	c.DeletedAt = null.NewTime(time.Now(), true)
	cupdate, err := c.Update(context.Background(), db, boil.Infer())
	fmt.Println(cupdate)
	if err != nil {
		panic(err.Error())
	}

	f, err := models.FindForum(context.Background(), db, id)
	if err != nil {
		panic(err.Error())
	}
	f.DeletedAt = null.NewTime(time.Now(), true)
	fudate, err := f.Update(context.Background(), db, boil.Infer())
	fmt.Println(fudate)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()
}
