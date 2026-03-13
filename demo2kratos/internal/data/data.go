package data

import (
	"github.com/google/wire"
	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
	"github.com/yylego/kratos-zap/zapkratos"
	"github.com/yylego/must"
	"github.com/yylego/rese"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData)

type Data struct {
	db *gorm.DB
}

func NewData(c *conf.Data, zapKratos *zapkratos.ZapKratos) (*Data, func(), error) {
	zapLog := zapKratos.SubZap()
	zapLog.SUG.Info("creating data resources")
	must.Same(c.Database.Driver, "sqlite3")
	db := rese.P1(gorm.Open(sqlite.Open(c.Database.Source), &gorm.Config{}))
	cleanup := func() {
		zapLog.SUG.Info("closing the data resources")
		_ = rese.P1(db.DB()).Close()
	}
	return &Data{db: db}, cleanup, nil
}
