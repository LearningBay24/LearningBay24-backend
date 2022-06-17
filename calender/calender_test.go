package calender

import (
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type AnyTime struct{}
type AnyString struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

// Match satisfies sqlmock.Argument interface
func (a AnyString) Match(v driver.Value) bool {
	_, ok := v.(null.String)
	return ok
}

func TestAddCourseToCalender(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected", err)
	}

	oldDB := boil.GetDB()
	defer func() {
		db.Close()
		boil.SetDB(oldDB)
	}()
	boil.SetDB(db)

	ctrl := &PublicController{db}

	// start
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `appointment` (`location`,`online`,`course_id`,`created_at`,`updated_at`,`deleted_at`) VALUES (?,?,?,?,?,?)")).WithArgs(1, AnyTime{}, AnyString{}, 1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `course` WHERE `id`=? and `deleted_at` is null")).WithArgs(1).WillReturnRows().RowsWillBeClosed()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `appointment` (`location`,`online`,`course_id`,`created_at`,`updated_at`,`deleted_at`) VALUES (?,?,?,?,?,?)")).WithArgs(1, AnyTime{}, AnyString{}, 1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	/*
		ID        string
		Date      string
		Location  string
		Online    string
		CourseID  string
	*/

	id, err := ctrl.AddCourseToCalender(time.Time{}, null.String{String: "Home", Valid: true}, 1, 1, false, 2, time.Time{})
	assert.NoError(t, err)
	assert.NotEqual(t, id, 0)
}
