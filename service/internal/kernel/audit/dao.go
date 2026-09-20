package audit

import (
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type LogDAO interface {
	dao.CRUD[Log]
}

type logDAO struct {
	dao.Base[Log]
	db *gorm.DB
}

func NewLogDAO(db *gorm.DB) LogDAO {
	return logDAO{Base: dao.NewBase[Log](db), db: db}
}
