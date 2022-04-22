// Package db implements operations on the db, in the learningbay24 context
package db

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"net/url"

	"learningbay24.de/backend/config"
	"learningbay24.de/backend/models"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Save a File to disk, creating a database entry alongside it.
// The fileName can change, depending on if a file with the same name exists already
// The file represents either a local file or a remote one
func SaveFile(db *sql.DB, fileName string, uploaderID int, isLocal bool, file *io.Reader) (*models.File, error) {
	// TODO: implement non-local files
	// TODO: implement non-course files
	path := config.Conf.Files.Path

	var f *models.File
	var err error

	if isLocal {
		f, err = saveLocalFile(db, path, fileName, uploaderID, file)
	} else {
		var u *url.URL
		u, err = url.ParseRequestURI(fileName)
		if err != nil {
			return nil, err
		}

		f, err = saveRemoteFile(db, path, u, uploaderID, file)
	}

	if err != nil {
		return nil, err
	}

	return f, nil
}

func saveLocalFile(db *sql.DB, path string, fileName string, uploaderID int, file *io.Reader) (*models.File, error) {
	// possibly changed name due to a file with the same name already existing
	name := fileName

	// check if file type is valid
	for num := 0; ; num++ {
		if _, err := os.Stat(filepath.Join(path, name)); err != nil {
			if os.IsNotExist(err) {
				if num != 0 {
					// TODO: extension
					name = fmt.Sprintf("%s-%d", name, num)
				}
				break
			} else {
				return nil, err
			}
		}
	}

	// TODO: actually save file locally
	f := models.File{Name: name, URI: filepath.Join(path, name), Local: 1, UploaderID: uploaderID}
	// TODO: db transaction
	err := f.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &f, nil
}

func saveRemoteFile(db *sql.DB, url string, u *url.URL, uploaderID int, file *io.Reader) (*models.File, error) {
	// TODO
	return nil, nil
}
