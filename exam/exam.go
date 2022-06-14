package exam

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"learningbay24.de/backend/dbi"
	"learningbay24.de/backend/models"
)

type ExamService interface {
	GetExam(examId int) (*models.Exam, error)
	GetAllExamsFromUser(userId int) (models.ExamSlice, error)
	GetAttendedExamsFromUser(userId int) (models.ExamSlice, error)
	GetPassedExamsFromUser(userId int) (models.ExamSlice, error)
	GetCreatedExamsFromUser(userId int) (models.ExamSlice, error)
	CreateExam(name, description string, date time.Time, duration, courseId, creatorId int, online int8, location null.String, registerDeadLine, deregisterDeadLine null.Time) (int, error)
	EditExam(fileName string, examId, creatorId int, local int8, file *io.Reader, date time.Time, duration int) (int, error)
	RegisterToExam(userId, examId int) (*models.User, error)
	DeregisterFromExam(userId, examId int) error
	AttendExam(userId, examId int) (*models.Exam, error)
	GetFileFromExam(examId int) ([]*models.File, error)
	SubmitAnswer(fileName, uri string, local bool, file io.Reader, examId, userId int) error
	GetRegisteredUsersFromExam(examId int) (models.UserHasExamSlice, error)
	GetAnswerFromAttendee(fileId int) (*models.File, error)
	GradeAnswer(examId, userId int, grade null.Int, passed null.Int8, feedback null.String) error
}

type PublicController struct {
	Database *sql.DB
}

// GetExam takes an examId and returns a struct of the exam with this ID
func (p *PublicController) GetExam(examId int) (*models.Exam, error) {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return nil, err
	}
	return ex, nil
}

// GetRegisteredExamsFromUser takes a userId and returns a slice of exams associated with it where the user is registered in
func (p *PublicController) GetRegisteredExamsFromUser(userId int) (models.ExamSlice, error) {
	var exams []*models.Exam
	err := queries.Raw("select * from exam, user_has_exam where user_has_exam.user_id=? AND user_has_exam.exam_id=exam.id AND user_has_exam.attended=0 AND user_has_exam.passed is null AND user_has_exam.deleted_at is null AND exam.deleted_at is null", userId).Bind(context.Background(), p.Database, &exams)
	if err != nil {
		return nil, err
	}

	return exams, nil
}

func (p *PublicController) GetExamsFromCourse(courseId int) (models.ExamSlice, error) {
	var exams []*models.Exam
	err := queries.Raw("select * from exam where course_id=? AND deleted_at is null", courseId).Bind(context.Background(), p.Database, &exams)
	if err != nil {
		return nil, err
	}
	return exams, nil
}

// GetAttendedExamsFromUser takes a userId and returns a slice of exams associated with it that are attended
func (p *PublicController) GetAttendedExamsFromUser(userId int) (models.ExamSlice, error) {
	// TODO: add a grade to the model
	exams, err := models.Exams(
		qm.From(models.TableNames.UserHasExam),
		qm.Where("user_has_exam.user_id=?", userId),
		qm.And("user_has_exam.exam_id=exam.id"),
		qm.And("user_has_exam.attended=1"),
		qm.And("user_has_exam.passed is null")).
		All(context.Background(), p.Database)
	if err != nil {
		return nil, err
	}

	return exams, nil
}

// GetPassedExamsFromUser GetAttendedExamsFromUser takes a userId and returns a slice of exams associated with it that are passed
func (p *PublicController) GetPassedExamsFromUser(userId int) (models.ExamSlice, error) {
	// TODO: add a grade to the model

	exams, err := models.Exams(
		qm.From(models.TableNames.UserHasExam),
		qm.Where("user_has_exam.user_id=?", userId),
		qm.And("user_has_exam.exam_id=exam.id"),
		qm.And("user_has_exam.passed=?", null.Int8From(1)),
		qm.And("user_has_exam.deleted_at is null")).
		All(context.Background(), p.Database)
	if err != nil {
		return nil, err
	}

	return exams, nil
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
		name = c.Name
	}
	ex := &models.Exam{Name: name, Description: description, Date: date, Duration: duration, CourseID: courseId, CreatorID: creatorId, Online: online, Location: location, RegisterDeadline: registerDeadLine, DeregisterDeadline: deregisterDeadLine}
	err = ex.Insert(context.Background(), p.Database, boil.Infer())
	if err != nil {
		return 0, err
	}
	return ex.ID, nil
}

// EditExam takes a fileName, examId, creatorId, file-handle, date, duration, and an indicator if the file is local
func (p *PublicController) EditExam(fileName string, examId, creatorId int, local int8, file *io.Reader, date time.Time, duration int) (int, error) {
	ex, err := p.GetExam(examId)
	if err != nil {
		return 0, err
	}
	if creatorId == ex.CreatorID {
		tx, err := p.Database.BeginTx(context.Background(), nil)
		if err != nil {
			return 0, err
		}
		// if exam is online and has a file and filename: upload file
		// TODO: restrict to pdfs only
		if ex.Online != 0 && fileName != "" && file != nil {
			fileId, err := dbi.SaveFile(p.Database, fileName, uri, creatorId, local, &file)
			if err != nil {
				if e := tx.Rollback(); e != nil {
					return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
				}
				return 0, err
			}

			f, err := models.FindFile(context.Background(), p.Database, fileId)
			if err != nil {
				if e := tx.Rollback(); e != nil {
					return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
				}
				return 0, err
			}

			err = ex.SetFiles(context.Background(), tx, false, f)
			if err != nil {
				if e := tx.Rollback(); e != nil {
					return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
				}
				return 0, err
			}
		}
		ex.Date = date
		ex.Duration = duration

		_, err = ex.Update(context.Background(), tx, boil.Infer())
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
			}

			return 0, err
		}
		if e := tx.Commit(); e != nil {
			return 0, fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
		}
		return ex.ID, nil
	}
	return 0, fmt.Errorf("invalid value for variable creatorId: %d doesn't match exam's creatorId", creatorId)
}

// RegisterToExam takes a userId and examId
// Created struct gets inserted into database
func (p *PublicController) RegisterToExam(userId, examId int) (*models.User, error) {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return nil, err
	}

	u, err := models.FindUser(context.Background(), p.Database, userId)
	if err != nil {
		return nil, err
	}

	// Fails if trying to register to an exam while deadline has passed
	curTime := time.Now()
	diff := curTime.Sub(ex.RegisterDeadline.Time)
	if diff.Minutes() <= 0 {
		uhex := models.UserHasExam{UserID: userId, ExamID: examId, Attended: 0}

		err = uhex.Insert(context.Background(), p.Database, boil.Infer())
		if err != nil {
			return nil, err
		}
		return u, nil
	}
	return nil, fmt.Errorf("can't register from exam: RegisterDeadline has passed")
}

// DeregisterFromExam takes a userId and examId and deactivates the registration
func (p *PublicController) DeregisterFromExam(userId, examId int) error {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return err
	}
	// Fails if trying to deregister from an exam while deadline has passed
	curTime := time.Now()
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
	return fmt.Errorf("can't deregister from exam: DeregisterDeadline has passed")

}

// AttendExam takes an userId and examId and marks the user's exam as attended
func (p *PublicController) AttendExam(userId, examId int) (*models.Exam, error) {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return nil, err
	}

	// Can attend to exam if exam start <= current time <= exam end
	curTime := time.Now()
	end := ex.Date.Add(time.Minute * time.Duration(ex.Duration))
	diffBegin := curTime.Sub(ex.Date)
	diffEnd := end.Sub(curTime)
	if diffBegin.Minutes() >= 0 {
		if diffEnd.Minutes() >= 0 {

			uhex, err := models.FindUserHasExam(context.Background(), p.Database, userId, examId)
			if err != nil {
				return nil, err
			}

			uhex.Attended = 1
			_, err = uhex.Update(context.Background(), p.Database, boil.Infer())
			if err != nil {
				return nil, err
			}
			return ex, nil
		}
		return nil, fmt.Errorf("can't attend exam: exam ended at %s, current time: %s, diff: %f", end.String(), curTime.String(), diffEnd.Minutes())
	}
	return nil, fmt.Errorf("can't attend exam: exam hasn't started yet")
}

// GetFileFromExam takes an examId and returns a slice with the file associated to the exam
func (p *PublicController) GetFileFromExam(examId int) ([]*models.File, error) {
	exists, err := models.ExamExists(context.Background(), p.Database, examId)
	if !exists || err != nil {
		return nil, err
	}

	var files []*models.File
	// NOTE: raw query is used because sqlboiler seems to not be able to query the database properly in this case when used with query building
	err = queries.Raw("select * from file, exam_has_files where exam_has_files.exam_id=? AND exam_has_files.file_id=file.id AND exam_has_files.deleted_at is null", examId).Bind(context.Background(), p.Database, &files)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// SubmitAnswer takes a filename, uri, local-indicator, file, examId, and userId and uploads the file as an answer
func (p *PublicController) SubmitAnswer(fileName, uri string, local bool, file io.Reader, examId, userId int) error {
	// TODO: max upload size

	fileId, err := dbi.SaveFile(p.Database, fileName, uri, userId, local, &file)
	if err != nil {
		return err
	}

	uhex, err := models.FindUserHasExam(context.Background(), p.Database, userId, examId)
	if err != nil {
		return err
	}
	uhex.FileID = null.Int{Int: fileId}

	err = uhex.Insert(context.Background(), p.Database, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}

// GetAllAttendees takes an examId and returns a slice of relations between the exam and all of it's registered users
func (p *PublicController) GetRegisteredUsersFromExam(examId int) (models.UserHasExamSlice, error) {
	var attendees []*models.UserHasExam
	err := queries.Raw("select * from user_has_exam, exam where exam_id=? AND user_has_exam.deleted_at is null AND exam.deleted_at is null", examId).Bind(context.Background(), p.Database, &attendees)
	if err != nil {
		return nil, err
	}

	return attendees, nil
}

// GetAnswerFromAttendee takes a fileId and returns a struct of the file with the corresponding ID
func (p *PublicController) GetAnswerFromAttendee(fileId int) (*models.File, error) {
	cm, err := models.FindFile(context.Background(), p.Database, fileId)
	if err != nil {
		return nil, err
	}

	return cm, err
}

// GradeAnswer takes an examId, userId, grade, passed-indicator, and feedback and grades the associated answer
// If every answer of an exam has a grade it sets itself to graded
func (p *PublicController) GradeAnswer(examId, userId int, grade null.Int, passed null.Int8, feedback null.String) error {
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
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return err
	}

	attendees, err := p.GetRegisteredUsersFromExam(examId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return err
	}

	for _, att := range attendees {
		if att.Grade.Int == 1 {
			if e := tx.Commit(); e != nil {
				return fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
			}
			return nil
		}
	}

	ex.Graded = 1
	_, err = ex.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
		}

		return err
	}
	if e := tx.Commit(); e != nil {
		return fmt.Errorf("fatal: unable to commit transaction on error: %s; %s", err, e)
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

	err := queries.Raw("select distinct exam.* from user_has_course, exam "+
		"where user_has_course.user_id=? AND user_has_course.course_id=exam.course_id AND user_has_course.deleted_at is null AND exam.deleted_at is null "+
		"AND exam.id not in( "+
		"select distinct exam.id from user_has_course, exam, user_has_exam "+
		"where user_has_exam.user_id=? AND user_has_exam.exam_id=exam.id AND user_has_exam.deleted_at is null)", userId, userId).Bind(context.Background(), p.Database, &exams)

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
		return 0, errors.New("there are still people registered into the exam")
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
