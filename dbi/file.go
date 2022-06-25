// Package dbi implements operations on the db, in the learningbay24 context
package dbi

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"learningbay24.de/backend/config"
	"learningbay24.de/backend/errs"
	"learningbay24.de/backend/models"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Save a File to disk, creating a database entry alongside it.
// The fileName can change, depending on if a file with the same name exists already. If the file is a web link (non local), the fileName will become the name given to the URL.
// The file represents either a local file or a remote one
func SaveFile(db *sql.DB, fileName string, uri string, uploaderID int, isLocal bool, file *io.Reader, fileSize int) (int, error) {
	filePath := config.Conf.Files.Path

	var id int
	var err error

	if isLocal {
		id, err = saveLocalFile(db, filePath, fileName, uploaderID, file, fileSize)
		if err != nil {
			return 0, err
		}
	} else {
		var u *url.URL
		u, err = url.ParseRequestURI(uri)
		if err != nil {
			return 0, err
		}

		id, err = saveRemoteFile(db, fileName, u, uploaderID, file)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

// Save a file locally by creating a new one, never overwriting an old one. Should a file with the exact same name exist already, append the new one with a suffix of "-<count>"
func saveLocalFile(db *sql.DB, filePath string, fileName string, uploaderID int, file *io.Reader, fileSize int) (int, error) {
	// possibly changed name due to a file with the same name already existing
	name := fileName

	ext := path.Ext(name)
	allowed := false
	if ext != "" {
		// strip leading dot
		ext = ext[1:]
		for _, e := range config.Conf.Files.AllowedFileTypes {
			if ext == e {
				// file type is allowed per config
				allowed = true
				break
			}
		}
	} else {
		return 0, errs.ErrNoFileExtension
	}

	if !allowed {
		return 0, errs.ErrFileExtensionNotAllowed
	}

	// check if file type is valid
	for num := 0; ; num++ {
		if _, err := os.Stat(filepath.Join(filePath, name)); err != nil {
			if !os.IsNotExist(err) {
				return 0, err
			} else {
				break
			}
		} else {
			if num != 0 {
				ext := path.Ext(name)
				f := strings.TrimSuffix(name, ext)
				name = fmt.Sprintf("%s-%d%s", f, num, ext)
			}
		}
	}

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	// verify user exists and whether the user reached the upload cap yet
	user, err := models.FindUser(context.Background(), tx, uploaderID)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	if config.Conf.Files.MaxUploadPerUser != 0 && user.UploadedBytes+fileSize > config.Conf.Files.MaxUploadPerUser {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, fmt.Errorf("user has reached the upload limit of %d bytes", config.Conf.Files.MaxUploadPerUser)
	}

	fullFile := filepath.Join(filePath, name)
	f := models.File{Name: name, URI: fullFile, Local: 1, UploaderID: uploaderID}
	err = f.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	// create file on disk
	fp, err := os.Create(fullFile)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}
	defer fp.Close()
	// write to the file on disk
	bufr := bufio.NewReader(*file)
	_, err = bufr.WriteTo(fp)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	return f.ID, nil
}

// Save a remote file, a.k.a. a web link, to the database.
func saveRemoteFile(db *sql.DB, linkName string, u *url.URL, uploaderID int, file *io.Reader) (int, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	f := models.File{Name: linkName, URI: u.String(), Local: 0, UploaderID: uploaderID}
	err = f.Insert(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return 0, fmt.Errorf("unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return 0, err
	}

	return f.ID, nil
}
