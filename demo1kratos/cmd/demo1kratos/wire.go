//go:build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
	"github.com/yylego/kratos-examples/demo1kratos/internal/server"
	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
	"github.com/yylego/kratos-zap/zapkratos"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *zapkratos.ZapKratos) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
