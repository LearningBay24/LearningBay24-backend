package course

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

//CreateCourse takes a name,enrollkey and description and adds a course and forum with that Name in the Database
func CreateCourse(name, enrollkey string, description null.String) {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/learningbay24")

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Connected")
	}
	f := &models.Forum{Name: name}
	err = f.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		panic(err.Error())
	}
	c := &models.Course{Name: name, ForumID: f.ID, Description: description}
	err = c.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		panic(err.Error())
	}
	db.Close()
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

	db.Close()
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

	db.Close()
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

	db.Close()
}
