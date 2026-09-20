package jobs

import (
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type JobRunDAO interface {
	dao.CRUD[JobRun]
}

type jobRunDAO struct {
	dao.Base[JobRun]
}

func NewJobRunDAO(db *gorm.DB) JobRunDAO {
	return jobRunDAO{Base: dao.NewBase[JobRun](db)}
}
