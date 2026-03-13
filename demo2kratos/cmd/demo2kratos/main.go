package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/yylego/done"
	"github.com/yylego/kratos-examples/demo2kratos"
	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
	"github.com/yylego/kratos-zap/zapkratos"
	"github.com/yylego/must"
	"github.com/yylego/osexistpath/osmustexist"
	"github.com/yylego/rese"
	"github.com/yylego/tern/zerotern"
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
}

func newApp(gs *grpc.Server, hs *http.Server, zapKratos *zapkratos.ZapKratos) *kratos.App {
	return kratos.New(
		kratos.ID(done.VCE(os.Hostname()).Omit()),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(zapKratos.NewLogger("network-service")),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()

	{
		rootBin := osmustexist.ROOT(filepath.Join(demo2kratos.SourceRoot(), "bin"))
		path1 := filepath.Join(rootBin, "log-newest.log")
		path2 := filepath.Join(rootBin, "log-oldest.log")

		// Clean session log on startup
		// 启动时清空会话日志
		if osmustexist.IsFile(path1) {
			must.Done(os.Truncate(path1, 0))
		}

		// Set default zap log to stdout and disk-file
		// 设置默认 zap 日志输出到标准输出和日志文件
		zaplog.SetLog(rese.P1(zaplog.NewZapLog(zaplog.NewConfig().
			AddOutputPaths(
				path1, path2, // Also log to file // 也输出到文件
			))).With(
			zap.String("service", zerotern.VF(Name, func() string {
				return filepath.Base(demo2kratos.SourceRoot())
			})),
			zap.String("version", zerotern.VV(Version, "v0.0.0")),
		))
	}

	// Create zapkratos logger with default zaplog
	// 使用默认的 zaplog 创建 zapkratos 日志
	zapKratos := zapkratos.NewZapKratos(zaplog.LOGGER, zapkratos.NewOptions())
	zapLog := zapKratos.SubZap()
	zapLog.LOG.Info("application starting...")
	zapLog.LOG.Info("reading-config-from-path", zap.String("config", flagconf))

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer rese.F0(c.Close)

	must.Done(c.Load())

	var cfg conf.Bootstrap
	must.Done(c.Scan(&cfg))

	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, zapKratos))
	defer cleanup()

	// start and wait for stop signal
	must.Done(app.Run())
}
