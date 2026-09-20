//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/test/testutil"
	"gorm.io/gorm"
)

var (
	testDB  *gorm.DB
	testCfg *config.Config
)

func TestMain(m *testing.M) {
	testDB = testutil.SetupTestDB()
	testCfg = testutil.LoadTestConfig()

	os.Exit(m.Run())
}
