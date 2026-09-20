package organization

import (
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type DAOMock struct {
	dao.CRUDMock[reference.Organization]
}
