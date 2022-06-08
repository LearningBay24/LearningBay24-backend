package exam

import (
	"context"
	"database/sql"
	"github.com/volatiletech/null/v8"
	"learningbay24.de/backend/models"
	"time"
)

type ExamService interface {
	GetExam(id int) (*models.Exam, error)
	CreateExam(name, description string, date time.Time, duration, courseId, creatorId int, online, graded int8, location null.String, registerDeadLine null.Time)
}

type PublicController struct {
	Database *sql.DB
}

func (p *PublicController) GetExam(id int) (*models.Exam, error) {
	e, err := models.FindExam(context.Background(), p.Database, id)
	if err != nil {
		return nil, err
	}
	return e, nil
}
