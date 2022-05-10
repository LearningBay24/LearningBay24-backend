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

// GetMaterials takes a courseID and returns a slice of files associated with it
func GetMaterials(db *sql.DB, courseId int) (models.FileSlice, error) {
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

// DeactivateMaterial takes both course-ID and file-ID and deactivates the chosen material
// Sets deactivation-timer and updates database
func DeactivateMaterial(db *sql.DB, courseId, fileId int) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	cm, err := models.FindFile(context.Background(), tx, fileId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = cm.Delete(context.Background(), tx, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	chf, err := models.FindCourseHasFile(context.Background(), tx, courseId, fileId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = chf.Delete(context.Background(), tx, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
