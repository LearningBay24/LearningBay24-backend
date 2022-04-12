package course

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"learningbay24.de/backend/models"
)

func CreateCourse(name string) {
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
	c := &models.Course{Name: name, ForumID: f.ID}
	err = c.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		panic(err.Error())
	}
	db.Close()
}
