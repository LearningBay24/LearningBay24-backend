package exam

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strconv"
	"time"

	"learningbay24.de/backend/course"
	"learningbay24.de/backend/dbi"
	"learningbay24.de/backend/errs"
	"learningbay24.de/backend/models"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ExamService interface {
	GetExamByID(examId int) (*models.Exam, error)
	GetRegisteredExamsFromUser(userId int) (models.ExamSlice, error)
	GetExamsFromCourse(courseId int) (models.ExamSlice, error)
	GetAttendedExamsFromUser(userId int) ([]*GradedExam, error)
	GetPassedExamsFromUser(userId int) ([]*GradedExam, error)
	GetCreatedExamsFromUser(userId int) (models.ExamSlice, error)
	CreateExam(name, description string, date time.Time, duration, courseId, creatorId int, online int8, location null.String, registerDeadLine, deregisterDeadLine null.Time) (int, error)
	EditExam(name, description string, date time.Time, duration, examId, creatorId int, online null.Int8, location null.String, registerDeadLine, deregisterDeadLine null.Time) error
	UploadExamFile(fileName string, uri string, uploaderId, examId int, local bool, file io.Reader) error
	DeleteExamFile(tx *sql.Tx, examId int) error
	RegisterToExam(userId, examId int) (*models.User, error)
	DeregisterFromExam(userId, examId int) error
	AttendExam(examId, userId int) error
	GetFileFromExam(examId int) ([]*models.File, error)
	SubmitAnswer(fileName, uri string, local bool, file io.Reader, examId, userId int) error
	GetRegisteredUsersFromExam(examId, userId int) (models.UserHasExamSlice, error)
	GetAnswerFromAttendee(userId, examId int) (*models.File, error)
	GradeAnswer(examId, creatorId, userId int, grade null.Int, passed null.Int8, feedback null.String) error
	SetAttended(examId, userId int) error
	GetUnregisteredExams(userId int) (models.ExamSlice, error)
	DeleteExam(examId int) (int, error)
	GetCourseFromExam(examId int) (*models.Course, error)
}

type GradedExam struct {
	models.Exam `boil:",bind"`
	UserID      int `boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	ExamID      int `boil:"exam_id" json:"exam_id" toml:"exam_id" yaml:"exam_id"`
	// Whether the user has attended the exam or not.
	Attended int8     `boil:"attended" json:"attended" toml:"attended" yaml:"attended"`
	Grade    null.Int `boil:"grade" json:"grade,omitempty" toml:"grade" yaml:"grade,omitempty"`
	// If the user that attended the exam passed it or not.
	Passed null.Int8 `boil:"passed" json:"passed,omitempty" toml:"passed" yaml:"passed,omitempty"`
	// The feedback given to the user about their solution to the exam.
	Feedback null.String `boil:"feedback" json:"feedback,omitempty" toml:"feedback" yaml:"feedback,omitempty"`

	FileID null.Int `boil:"file_id" json:"file_id,omitempty" toml:"file_id" yaml:"file_id,omitempty"`
}

type Attendee struct {
	models.User `boil:",bind"`
	// Whether the user has attended the exam or not.
	Attended int8     `boil:"attended" json:"attended" toml:"attended" yaml:"attended"`
	Grade    null.Int `boil:"grade" json:"grade,omitempty" toml:"grade" yaml:"grade,omitempty"`
	// If the user that attended the exam passed it or not.
	Passed null.Int8 `boil:"passed" json:"passed,omitempty" toml:"passed" yaml:"passed,omitempty"`
	// The feedback given to the user about their solution to the exam.
	Feedback null.String `boil:"feedback" json:"feedback,omitempty" toml:"feedback" yaml:"feedback,omitempty"`

	FileID null.Int `boil:"file_id" json:"file_id,omitempty" toml:"file_id" yaml:"file_id,omitempty"`
}

type PublicController struct {
	Database *sql.DB
}

// GetExamByID takes an examId and returns a struct of the exam with this ID
func (p *PublicController) GetExamByID(examId int) (*models.Exam, error) {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return nil, err
	}
	return ex, nil
}

// GetRegisteredExamsFromUser takes a userId and returns a slice of exams associated with it where the user is registered in
func (p *PublicController) GetRegisteredExamsFromUser(userId int) (models.ExamSlice, error) {
	var exams []*models.Exam
	err := queries.Raw("select * from exam, user_has_exam where user_has_exam.user_id=? AND user_has_exam.exam_id=exam.id AND user_has_exam.attended=0 AND user_has_exam.passed is null AND user_has_exam.deleted_at is null AND exam.deleted_at is null ORDER BY date ASC", userId).Bind(context.Background(), p.Database, &exams)
	if err != nil {
		return nil, err
	}

	return exams, nil
}

func (p *PublicController) GetExamsFromCourse(courseId int) (models.ExamSlice, error) {
	var exams []*models.Exam
	err := queries.Raw("select * from exam where course_id=? AND deleted_at is null ORDER BY date ASC", courseId).Bind(context.Background(), p.Database, &exams)
	if err != nil {
		return nil, err
	}
	return exams, nil
}

// GetAttendedExamsFromUser takes a userId and returns a slice of exams associated with it that are attended
func (p *PublicController) GetAttendedExamsFromUser(userId int) ([]*GradedExam, error) {
	var gex []*GradedExam

	err := models.NewQuery(
		qm.Select("exam.*", "user_has_exam.*"),
		qm.From(models.TableNames.Exam),
		qm.InnerJoin("user_has_exam on exam.id = user_has_exam.exam_id"),
		qm.Where("user_has_exam.attended=1"),
		qm.And("user_has_exam.user_id = ?", userId),
		qm.And("(user_has_exam.passed is null"),
		qm.Or("user_has_exam.passed = 0)"),
	).Bind(context.Background(), p.Database, &gex)
	if err != nil {
		return nil, err
	}

	return gex, nil
}

// GetPassedExamsFromUser GetAttendedExamsFromUser takes a userId and returns a slice of exams associated with it that are passed
func (p *PublicController) GetPassedExamsFromUser(userId int) ([]*GradedExam, error) {
	var gex []*GradedExam

	err := models.NewQuery(
		qm.Select("exam.*", "user_has_exam.*"),
		qm.From(models.TableNames.Exam),
		qm.InnerJoin("user_has_exam on exam.id = user_has_exam.exam_id"),
		qm.Where("user_has_exam.attended=1"),
		qm.And("user_has_exam.user_id = ?", userId),
		qm.And("user_has_exam.passed = 1"),
	).Bind(context.Background(), p.Database, &gex)
	if err != nil {
		return nil, err
	}

	return gex, nil
}

// GetCreatedExamsFromUser takes a userId and returns a slice of exams associated with it that got created by the user
func (p *PublicController) GetCreatedExamsFromUser(userId int) (models.ExamSlice, error) {
	exams, err := models.Exams(models.ExamWhere.CreatorID.EQ(userId)).All(context.Background(), p.Database)
	if err != nil {
		return nil, err
	}
	return exams, nil
}

// CreateExam takes a name, description, date, duration, location, courseId, creatorId, de-, and register-deadline and indicators for online and graded
// Created struct gets inserted into database
func (p *PublicController) CreateExam(name, description string, date time.Time, duration, courseId, creatorId int, online int8, location null.String, registerDeadLine, deregisterDeadLine null.Time) (int, error) {
	c, err := models.FindCourse(context.Background(), p.Database, courseId)
	if err != nil {
		return 0, err
	}
	if name == "" {
		y, m, d := date.Date()
		creator, err := models.FindUser(context.Background(), p.Database, creatorId)
		if err != nil {
			return 0, err
		}
		name = c.Name + ", " + strconv.Itoa(d) + "." + strconv.Itoa(int(m)) + "." + strconv.Itoa(y) + ", " + creator.Surname
	}
	ex := &models.Exam{Name: name, Description: description, Date: date, Duration: duration, CourseID: courseId, CreatorID: creatorId, Online: online, Location: location, RegisterDeadline: registerDeadLine, DeregisterDeadline: deregisterDeadLine}
	err = ex.Insert(context.Background(), p.Database, boil.Infer())
	if err != nil {
		return 0, err
	}
	return ex.ID, nil
}

// EditExam takes a fileName, examId, creatorId, file-handle, date, duration, and an indicator if the file is local
func (p *PublicController) EditExam(name, description string, date time.Time, duration, examId int, online null.Int8, location null.String, registerDeadLine, deregisterDeadLine null.Time) error {
	ex, err := p.GetExamByID(examId)
	if err != nil {
		return err
	}
	tx, err := p.Database.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	// if exam is online and has a file and filename: upload file

	if name != "" {
		ex.Name = name
	}

	if description != "" {
		ex.Description = description
	}

	if !date.IsZero() {
		ex.Date = date
	}
	if duration != 0 {
		ex.Duration = duration
	}

	if online.Valid {
		ex.Online = online.Int8
	}

	if location.String != "" {
		ex.Location.String = location.String
	}

	if !registerDeadLine.IsZero() {
		ex.RegisterDeadline.Time = registerDeadLine.Time
	}

	if !deregisterDeadLine.IsZero() {
		ex.DeregisterDeadline.Time = deregisterDeadLine.Time
	}

	_, err = ex.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return err
	}
	if e := tx.Commit(); e != nil {
		return fmt.Errorf("unable to commit transaction on error: %s; %w", err, e)
	}
	return nil
}

// UploadExamFile takes a fileName, URI, associated uploaderId, examId and indicator if file is local or remote
// Created struct gets inserted into database
func (p *PublicController) UploadExamFile(fileName string, uri string, uploaderId, examId int, local bool, file io.Reader, fileSize int) error {
	ex, err := p.GetExamByID(examId)
	if err != nil {
		return err
	}

	fileId, err := dbi.SaveFile(p.Database, fileName, uri, uploaderId, local, &file, fileSize)
	if err != nil {
		return err
	}

	tx, err := p.Database.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	f, err := models.FindFile(context.Background(), tx, fileId)
	if err != nil {
		// NOTE: disregard error, only god can help us now
		_ = dbi.DeleteFile(p.Database, fileId)
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}
		return err
	}

	err = p.DeleteExamFile(tx, examId)
	if err != nil {
		// NOTE: disregard error, only god can help us now
		_ = dbi.DeleteFile(p.Database, fileId)
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}
		return err
	}

	err = ex.SetFiles(context.Background(), tx, false, f)
	if err != nil {
		// NOTE: disregard error, only god can help us now
		_ = dbi.DeleteFile(p.Database, fileId)
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}
		return err
	}

	if e := tx.Commit(); e != nil {
		// NOTE: disregard error, only god can help us now
		_ = dbi.DeleteFile(p.Database, fileId)
		return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
	}
	return nil
}

// DeleteExamFile takes a transaction and examId and deletes the file associated to the exam
func (p *PublicController) DeleteExamFile(tx *sql.Tx, examId int) error {
	var files []*models.File
	err := queries.Raw("select * from file, exam_has_files where exam_has_files.exam_id=? AND exam_has_files.file_id = file.id AND file.id is null", examId).Bind(context.Background(), p.Database, &files)
	if err != nil {
		return err
	}

	if len(files) != 0 {
		_, err = files[0].Delete(context.Background(), tx, false)
		if err != nil {
			return err
		}
	}
	return nil
}

// RegisterToExam takes a userId and examId
// Created struct gets inserted into database
func (p *PublicController) RegisterToExam(userId, examId int) (*models.User, error) {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return nil, err
	}

	if ex.CreatorID == userId {
		return nil, errs.ErrSelfRegisterExam
	}

	u, err := models.FindUser(context.Background(), p.Database, userId)
	if err != nil {
		return nil, err
	}

	// Fails if trying to register to an exam while deadline has passed
	curTime := time.Now()
	curTime = curTime.Add(time.Minute * 120)
	diff := curTime.Sub(ex.RegisterDeadline.Time)
	if diff.Minutes() <= 0 {
		// need to set DeletedAt back to zero-value if row already exists

		var uhex models.UserHasExamSlice
		// first check if relation already exists in the database and either insert a new row or reset deleted_at
		err = queries.Raw("select * from user_has_exam where exam_id=? AND user_id=?", examId, userId).Bind(context.Background(), p.Database, &uhex)
		if err != nil {
			return nil, err
		}
		if len(uhex) > 0 {
			uhex[0].DeletedAt = null.TimeFromPtr(nil)
			_, err = uhex[0].Update(context.Background(), p.Database, boil.Infer())
			if err != nil {
				return nil, err
			}
			return u, nil
		}

		newUhex := models.UserHasExam{UserID: userId, ExamID: examId}
		err = newUhex.Insert(context.Background(), p.Database, boil.Infer())
		if err != nil {
			return nil, err
		}

		return u, nil
	}

	return nil, errs.ErrRegisterDeadlinePassed
}

// DeregisterFromExam takes a userId and examId and deactivates the registration
func (p *PublicController) DeregisterFromExam(userId, examId int) error {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return err
	}
	// Fails if trying to deregister from an exam while deadline has passed
	curTime := time.Now()
	curTime = curTime.Add(time.Minute * 120)
	diff := curTime.Sub(ex.DeregisterDeadline.Time)
	if diff.Minutes() <= 0 {
		uhex, err := models.FindUserHasExam(context.Background(), p.Database, userId, examId)
		if err != nil {
			return err
		}

		_, err = uhex.Delete(context.Background(), p.Database, false)
		if err != nil {
			return err
		}
		return nil
	}

	return errs.ErrUnregisterDeadlinePassed
}

// AttendExam takes an userId and examId and marks the user's exam as attended
func (p *PublicController) AttendExam(examId, userId int) error {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return err
	}

	// Can attend to exam if exam start <= current time <= exam end
	curTime := time.Now()
	curTime = curTime.Add(time.Minute * 120)
	end := ex.Date.Add(time.Minute * time.Duration(ex.Duration))
	diffBegin := curTime.Sub(ex.Date)
	diffEnd := end.Sub(curTime)
	if diffBegin.Minutes() >= 0 {
		if diffEnd.Minutes() >= 0 {
			uhex, err := models.FindUserHasExam(context.Background(), p.Database, userId, examId)
			if err != nil {
				return err
			}

			uhex.Attended = 1
			_, err = uhex.Update(context.Background(), p.Database, boil.Infer())
			if err != nil {
				return err
			}
			return nil
		}

		return errs.ErrExamEnded
	}

	return errs.ErrExamHasntStarted
}

// GetFileFromExam takes an examId and returns a slice with the file associated to the exam
func (p *PublicController) GetFileFromExam(examId int) ([]*models.File, error) {
	exists, err := models.ExamExists(context.Background(), p.Database, examId)
	if !exists || err != nil {
		return nil, err
	}

	var files []*models.File
	// NOTE: raw query is used because sqlboiler seems to not be able to query the database properly in this case when used with query building
	err = queries.Raw("select * from file, exam_has_files where exam_has_files.exam_id=? AND exam_has_files.file_id=file.id", examId).Bind(context.Background(), p.Database, &files)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, errs.ErrNoUploads
	}
	return files, nil
}

// SubmitAnswer takes a filename, uri, local-indicator, file, examId, and userId and uploads the file as an answer
func (p *PublicController) SubmitAnswer(fileName, uri string, examId, userId int, local bool, file io.Reader, fileSize int) error {
	fileId, err := dbi.SaveFile(p.Database, fileName, uri, userId, local, &file, fileSize)
	if err != nil {
		return err
	}

	uhex, err := models.FindUserHasExam(context.Background(), p.Database, userId, examId)
	if err != nil {
		return err
	}
	fid := null.IntFrom(fileId)
	uhex.FileID = fid

	_, err = uhex.Update(context.Background(), p.Database, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}

// GetRegisteredUsersFromExam takes an examId and userId and returns a slice of relations between the exam and all of it's registered users
func (p *PublicController) GetRegisteredUsersFromExam(examId, userId int) ([]*Attendee, error) {
	var attendees []*Attendee

	err := models.NewQuery(
		qm.Select("user.*", "user_has_exam.*"),
		qm.From(models.TableNames.User),
		qm.InnerJoin("user_has_exam on user.id = user_has_exam.user_id"),
		qm.Where("user_has_exam.exam_id=?", examId),
		qm.And("user_has_exam.deleted_at is null"),
		qm.And("user.deleted_at is null"),
	).Bind(context.Background(), p.Database, &attendees)
	if err != nil {
		return nil, err
	}

	return attendees, nil
}

// GetAttendeesFromExam takes an examId and userId and returns a slice of relations between the exam and all of it's registered users
func (p *PublicController) GetAttendeesFromExam(examId, userId int) ([]*Attendee, error) {
	var attendees []*Attendee

	err := models.NewQuery(
		qm.Select("user.*", "user_has_exam.*"),
		qm.From(models.TableNames.User),
		qm.InnerJoin("user_has_exam on user.id = user_has_exam.user_id"),
		qm.Where("user_has_exam.exam_id=?", examId),
		qm.And("user_has_exam.attended=1"),
		qm.And("user_has_exam.deleted_at is null"),
		qm.And("user.deleted_at is null"),
	).Bind(context.Background(), p.Database, &attendees)
	if err != nil {
		return nil, err
	}

	return attendees, nil
}

// GetAnswerFromAttendee takes a fileId and returns a struct of the file with the corresponding ID
func (p *PublicController) GetAnswerFromAttendee(userId, examId int) (*models.File, error) {
	uhex, err := models.FindUserHasExam(context.Background(), p.Database, userId, examId)
	if err != nil {
		return nil, err
	}

	cm, err := models.FindFile(context.Background(), p.Database, uhex.FileID.Int)
	if err != nil {
		return nil, err
	}

	return cm, err
}

// GradeAnswer takes an examId, creatorId, userId, grade, passed-indicator, and feedback and grades the associated answer
// If every answer of an exam has a grade it sets itself to graded
func (p *PublicController) GradeAnswer(examId, creatorId, userId int, grade null.Int, passed null.Int8, feedback null.String) error {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return err
	}

	uhex, err := models.FindUserHasExam(context.Background(), p.Database, userId, examId)
	if err != nil {
		return err
	}

	tx, err := p.Database.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	uhex.Grade = grade
	uhex.Passed = passed
	uhex.Feedback = feedback
	_, err = uhex.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return err
	}

	attendees, err := p.GetRegisteredUsersFromExam(examId, creatorId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return err
	}

	for _, att := range attendees {
		if att.Grade.Int == 1 {
			if e := tx.Commit(); e != nil {
				return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
			}
			return nil
		}
	}

	ex.Graded = 1
	_, err = ex.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return err
	}
	if e := tx.Commit(); e != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("unable to rollback transaction on error: %s; %w", err, e)
		}

		return err
	}
	return nil

}

// SetAttended takes an examId and userId and sets the corresponding registered exam of the user to attended
func (p *PublicController) SetAttended(examId, userId int) error {
	uhex, err := models.FindUserHasExam(context.Background(), p.Database, userId, examId)
	if err != nil {
		return err
	}

	uhex.Attended = 1
	_, err = uhex.Update(context.Background(), p.Database, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}

// GetUnregisteredExams takes an userId and returns a slice of exams associated with it
func (p *PublicController) GetUnregisteredExams(userId int) (models.ExamSlice, error) {
	var exams []*models.Exam

	err := queries.Raw("select distinct exam.* from user_has_course, exam \n"+
		"where user_has_course.user_id=? \nAND user_has_course.course_id=exam.course_id \nAND exam.creator_id !=? \nAND user_has_course.deleted_at is null \nAND exam.deleted_at is null \n"+
		"AND exam.id not in( \n"+
		"select distinct exam.id from user_has_course, exam, user_has_exam \n"+
		"where user_has_exam.user_id=? AND user_has_exam.exam_id=exam.id AND user_has_exam.deleted_at is null)", userId, userId, userId).Bind(context.Background(), p.Database, &exams)

	if err != nil {
		return nil, err
	}

	return exams, nil
}

// DeleteExam takes an examId and soft-deletes the associated exam
func (p *PublicController) DeleteExam(examId int) (int, error) {
	uhex, err := models.UserHasExams(models.UserHasExamWhere.ExamID.EQ(examId)).Count(context.Background(), p.Database)
	if err != nil {
		return 0, err
	}
	if uhex > 0 {
		return 0, errs.ErrDeleteExamNotEmpty
	}

	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return 0, err
	}

	_, err = ex.Delete(context.Background(), p.Database, false)
	if err != nil {
		return 0, err
	}

	return ex.ID, nil
}

func (p *PublicController) GetCourseFromExam(examId int) (*models.Course, error) {

	exam, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return nil, err
	}
	c, err := course.GetCourse(p.Database, exam.CourseID)
	if err != nil {
		return nil, err
	}
	return c, nil
}
