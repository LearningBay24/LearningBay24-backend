package coursematerial

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"learningbay24.de/backend/dbi"
	"learningbay24.de/backend/models"
)

// GetMaterialFromCourse takes an ID and returns a struct of the file with the corresponding ID
func GetMaterialFromCourse(db *sql.DB, courseId int, fileId int) (*models.File, error) {
	// TODO: use courseId to verify that the file is indeed in that course
	cm, err := models.FindFile(context.Background(), db, fileId)
	if err != nil {
		return nil, err
	}
	return cm, err
}

// GetAllMaterialsFromCourse takes a courseID and returns a slice of files associated with it
func GetAllMaterialsFromCourse(db *sql.DB, courseId int) (models.FileSlice, error) {
	var files []*models.File
	// NOTE: raw query is used because sqlboiler seems to not be able to query the database properly in this case when used with query building
	err := queries.Raw("select * from file, course_has_files where course_has_files.course_id=? AND course_has_files.file_id=file.id", courseId).Bind(context.Background(), db, &files)

	if err != nil {
		return nil, err
	}

	return files, nil
}

// CreateMaterial takes a fileName, URI, associated uploader-id, course, id and indicator if file is local or remote
// Created struct gets inserted into database
func CreateMaterial(dbHandle *sql.DB, fileName string, uri string, uploaderId, courseId int, local bool, file io.Reader) error {
	// TODO: max upload size

	fileId, err := dbi.SaveFile(dbHandle, fileName, uri, uploaderId, local, &file)
	if err != nil {
		return err
	}

	chf := models.CourseHasFile{
		CourseID: courseId, FileID: fileId,
	}

	err = chf.Insert(context.Background(), dbHandle, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

// DeleteMaterialFromCourse takes both course-ID and file-ID and soft-deletes the chosen material
func DeleteMaterialFromCourse(db *sql.DB, courseId, fileId int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	cm, err := models.FindFile(context.Background(), tx, fileId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
	}

	_, err = cm.Delete(context.Background(), tx, false)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
	}

	chf, err := models.FindCourseHasFile(context.Background(), tx, courseId, fileId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
	}

	_, err = chf.Delete(context.Background(), tx, false)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}
	}

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}

	return nil
}

// DeleteAllMaterialsFromCourse takes the ID of a course and deactivates all files associated with it
// TODO: compability with transactions
func DeleteAllMaterialsFromCourse(db *sql.DB, courseId int, hardDelete bool) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	materials, err := GetAllMaterialsFromCourse(db, courseId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}

		return err
	}

	_, err = materials.DeleteAll(context.Background(), tx, hardDelete)
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

// RenameMaterialFromCourse takes the ID of an existing file and name and overwrites the corresponding string with the new one
func RenameMaterialFromCourse(db *sql.DB, fileId int, fileName string) error {
	if fileName == "" {
		return fmt.Errorf("Invalid value for variable fileName: String can't be empty!")
	}

	cm, err := models.FindFile(context.Background(), db, fileId)
	if err != nil {
		return err
	}

	cm.Name = fileName

	_, err = cm.Update(context.Background(), db, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}
