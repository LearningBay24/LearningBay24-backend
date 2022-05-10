package coursematerial

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"learningbay24.de/backend/db"
	"learningbay24.de/backend/models"
)

// GetMaterial takes an ID and returns a struct of the file with the corresponding ID
func GetMaterialFromCourse(db *sql.DB, fileId int) (*models.File, error) {
	cm, err := models.FindFile(context.Background(), db, fileId)
	if err != nil {
		return nil, err
	}
	return cm, err
}

// GetMaterials takes a courseID and returns a slice of files associated with it
func GetMaterialSliceFromCourse(db *sql.DB, courseId int) (models.FileSlice, error) {
	files, err := models.Files(
		qm.From(models.TableNames.CourseHasFiles),
		qm.Where("course_has_files.course_id=?", courseId),
		qm.And("course_has_files.file_id = file.id"),
	).All(context.Background(), db)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// CreateMaterial takes a fileName, URI, associated uploader-id, course, id and indicator if file is local or remote
// Created struct gets inserted into database
func CreateMaterial(dbHandle *sql.DB, fileName string, uri string, uploaderId, courseId int, local int8, file *io.Reader) error {

	var isLocal bool
	switch local {
	case 0:
		isLocal = false
	case 1:
		isLocal = true
	default:
		return fmt.Errorf("Invalid value for variable local: %d", local)
	}

	fileId, err := db.SaveFile(dbHandle, fileName, uploaderId, isLocal, file)
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

// DeactivateMaterial takes both course-ID and file-ID and deactivates the chosen material
// Sets deactivation-timer and updates database
func DeactivateMaterialFromCourse(db *sql.DB, courseId, fileId int) error {
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

// DeactivateAllMaterials takes the ID of a course and deactivates all files associated with it
func DeactivateAllMaterialsFromCourse(db *sql.DB, courseId int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	materials, err := GetMaterialSliceFromCourse(db, courseId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err, e)
		}

		return err
	}
	materials.DeleteAll(context.Background(), tx, false)

	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
	}
	return nil
}

// UpdateMaterial takes the ID of an existing file and name and overwrites the corresponding string with the new one
func RenameMaterialFromCourse(db *sql.DB, fileId int, name string) error {
	cm, err := models.FindFile(context.Background(), db, fileId)
	if err != nil {
		return err
	}

	cm.Name = name

	_, err = cm.Update(context.Background(), db, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}
