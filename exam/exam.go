package exam

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"learningbay24.de/backend/dbi"
	"learningbay24.de/backend/models"
)

type ExamService interface {
	GetExamByID(examId int) (*models.Exam, error)
	GetAllExamsFromUser(userId int) (models.ExamSlice, error)
	GetAttendedExamsFromUser(userId int) (models.ExamSlice, error)
	GetPassedExamsFromUser(userId int) (models.ExamSlice, error)
	GetCreatedExamsFromUser(userId int) (models.ExamSlice, error)
	CreateExam(name, description string, date time.Time, duration, courseId, creatorId int, online int8, location null.String, registerDeadLine, deregisterDeadLine null.Time) (int, error)
	EditExam(fileName string, examId, creatorId int, local int8, file *io.Reader, date time.Time, duration int) (int, error)
	RegisterToExam(userId, examId int) error
	DeregisterFromExam(userId, examId int) error
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

// GetAllExamsFromUser takes a userId and returns a slice of exams associated with it
func (p *PublicController) GetAllExamsFromUser(userId int) (models.ExamSlice, error) {
	exams, err := models.Exams(models.UserHasExamWhere.UserID.EQ(userId)).All(context.Background(), p.Database)
	if err != nil {
		return nil, err
	}

	return exams, nil
}

// GetAttendedExamsFromUser takes a userId and returns a slice of exams associated with it that are attended
func (p *PublicController) GetAttendedExamsFromUser(userId int) (models.ExamSlice, error) {
	exams, err := models.Exams(
		qm.From(models.TableNames.UserHasExam),
		qm.Where("user_has_exam.user_id=?", userId),
		qm.And("user_has_exam.attended=?", 1)).
		All(context.Background(), p.Database)
	if err != nil {
		return nil, err
	}

	return exams, nil
}

// GetPassedExamsFromUser GetAttendedExamsFromUser takes a userId and returns a slice of exams associated with it that are passed
func (p *PublicController) GetPassedExamsFromUser(userId int) (models.ExamSlice, error) {
	exams, err := models.Exams(
		qm.From(models.TableNames.UserHasExam),
		qm.Where("user_has_exam.user_id=?", userId),
		qm.And("user_has_exam.passed=?", null.Int8{Int8: 1})).
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
func (p *PublicController) EditExam(fileName string, examId, creatorId int, local int8, uri string, file *io.Reader, date time.Time, duration int) (int, error) {
	ex, err := p.GetExamByID(examId)
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
			var isLocal bool
			switch local {
			case 0:
				isLocal = false
			case 1:
				isLocal = true
			default:
				if e := tx.Rollback(); e != nil {
					return 0, fmt.Errorf("fatal: unable to rollback transaction on error: %s; %s", err.Error(), e.Error())
				}
				return 0, fmt.Errorf("invalid value for variable local: %d", local)
			}
			fileId, err := dbi.SaveFile(p.Database, fileName, uri, creatorId, isLocal, file)
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
func (p *PublicController) RegisterToExam(userId, examId int) error {
	ex, err := models.FindExam(context.Background(), p.Database, examId)
	if err != nil {
		return err
	}

	_, err = models.FindUser(context.Background(), p.Database, examId)
	if err != nil {
		return err
	}
	// Fails if trying to register to an exam while deadline has passed
	curTime := time.Now()
	diff := curTime.Sub(ex.RegisterDeadline.Time)
	if diff.Minutes() <= 0 {
		uhex := models.UserHasExam{UserID: userId, ExamID: examId}
		err = uhex.Insert(context.Background(), p.Database, boil.Infer())
		if err != nil {
			return err
		}

		return nil
	}
	return fmt.Errorf("can't register from exam: RegisterDeadline has passed")
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

func (p *PublicController) AttendExam(userId, examId int) (*models.File, error) {
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
		if diffEnd >= 0 {

			uhex, err := models.FindUserHasExam(context.Background(), p.Database, userId, examId)
			if err != nil {
				return nil, err
			}

			uhex.Attended = 1
			_, err = uhex.Update(context.Background(), p.Database, boil.Infer())
			if err != nil {
				return nil, err
			}

		}
		return nil, fmt.Errorf("can't append exam: Duration has passed")
	}
	return nil, fmt.Errorf("can't append exam: exam hasn't started yet")
}

func (p *PublicController) SubmitAnswer(userId, examId int) error {
	return nil
}
