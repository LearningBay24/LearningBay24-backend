// Package db implements operations on the db, in the learningbay24 context
package db

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"learningbay24.de/backend/config"
	"learningbay24.de/backend/models"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Location of an Object, like a File. One of User, Course, Directory, Exam, Submission
type LearningLocation int

const (
	// Profile Picture
	User LearningLocation = iota
	Course
	Directory
	Exam
	Submission
)

// Save a File to disk, creating a database entry alongside it.
// The fileName can change, depending on if a file with the same name exists already
func SaveFile(db *sql.DB, fileName string, uploaderID int, isLocal bool, location LearningLocation, locationID int, file io.Reader) (*models.File, error) {
	// TODO: implement non-local files
	// TODO: implement non-course files
	path := config.Conf.Files.Path
	name := fileName

	for num := 0; ; num++ {
		if _, err := os.Stat(filepath.Join(path, name)); err != nil {
			if os.IsNotExist(err) {
				if num != 0 {
					name = fmt.Sprintf("%s-%d", name, num)
				}
				break
			} else {
				return nil, err
			}
		}
	}

	f := models.File{Name: name, URI: filepath.Join(path, name), Local: 1, UploaderID: uploaderID}
	err := f.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &f, nil
}
