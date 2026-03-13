# Changes

Code differences compared to source project.

## Makefile (+6 -1)

```diff
@@ -1,6 +1,11 @@
 GOHOSTOS:=$(shell go env GOHOSTOS)
 GOPATH:=$(shell go env GOPATH)
-VERSION=$(shell git describe --tags --always)
+#这是官方推荐的
+#VERSION=$(shell git describe --tags --always)
+#因为在开发阶段都是不打标签的，在很长的时间里可能都没有标签，这里使用较长的提交哈希
+#VERSION=$(shell git describe --tags 2>/dev/null || git rev-parse HEAD)
+#这样就能涵盖需要的
+VERSION=$(shell git describe --tags --always --abbrev=40 --dirty=+code)
 
 ifeq ($(GOHOSTOS), windows)
 	#the `find.exe` is different from `find` in bash/shell.
```

## cmd/demo2kratos/main.go (+42 -14)

```diff
@@ -3,18 +3,23 @@
 import (
 	"flag"
 	"os"
+	"path/filepath"
 
 	"github.com/go-kratos/kratos/v2"
 	"github.com/go-kratos/kratos/v2/config"
 	"github.com/go-kratos/kratos/v2/config/file"
-	"github.com/go-kratos/kratos/v2/log"
-	"github.com/go-kratos/kratos/v2/middleware/tracing"
 	"github.com/go-kratos/kratos/v2/transport/grpc"
 	"github.com/go-kratos/kratos/v2/transport/http"
 	"github.com/yylego/done"
+	"github.com/yylego/kratos-examples/demo2kratos"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
+	"github.com/yylego/kratos-zap/zapkratos"
 	"github.com/yylego/must"
+	"github.com/yylego/osexistpath/osmustexist"
 	"github.com/yylego/rese"
+	"github.com/yylego/tern/zerotern"
+	"github.com/yylego/zaplog"
+	"go.uber.org/zap"
 )
 
 // go build -ldflags "-X main.Version=x.y.z"
@@ -31,13 +36,13 @@
 	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
 }
 
-func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
+func newApp(gs *grpc.Server, hs *http.Server, zapKratos *zapkratos.ZapKratos) *kratos.App {
 	return kratos.New(
 		kratos.ID(done.VCE(os.Hostname()).Omit()),
 		kratos.Name(Name),
 		kratos.Version(Version),
 		kratos.Metadata(map[string]string{}),
-		kratos.Logger(logger),
+		kratos.Logger(zapKratos.NewLogger("network-service")),
 		kratos.Server(
 			gs,
 			hs,
@@ -47,15 +52,38 @@
 
 func main() {
 	flag.Parse()
-	logger := log.With(log.NewStdLogger(os.Stdout),
-		"ts", log.DefaultTimestamp,
-		"caller", log.DefaultCaller,
-		"service.id", kratos.ID(done.VCE(os.Hostname()).Omit()),
-		"service.name", Name,
-		"service.version", Version,
-		"trace.id", tracing.TraceID(),
-		"span.id", tracing.SpanID(),
-	)
+
+	{
+		rootBin := osmustexist.ROOT(filepath.Join(demo2kratos.SourceRoot(), "bin"))
+		path1 := filepath.Join(rootBin, "log-newest.log")
+		path2 := filepath.Join(rootBin, "log-oldest.log")
+
+		// Clean session log on startup
+		// 启动时清空会话日志
+		if osmustexist.IsFile(path1) {
+			must.Done(os.Truncate(path1, 0))
+		}
+
+		// Set default zap log to stdout and disk-file
+		// 设置默认 zap 日志输出到标准输出和日志文件
+		zaplog.SetLog(rese.P1(zaplog.NewZapLog(zaplog.NewConfig().
+			AddOutputPaths(
+				path1, path2, // Also log to file // 也输出到文件
+			))).With(
+			zap.String("service", zerotern.VF(Name, func() string {
+				return filepath.Base(demo2kratos.SourceRoot())
+			})),
+			zap.String("version", zerotern.VV(Version, "v0.0.0")),
+		))
+	}
+
+	// Create zapkratos logger with default zaplog
+	// 使用默认的 zaplog 创建 zapkratos 日志
+	zapKratos := zapkratos.NewZapKratos(zaplog.LOGGER, zapkratos.NewOptions())
+	zapLog := zapKratos.SubZap()
+	zapLog.LOG.Info("application starting...")
+	zapLog.LOG.Info("reading-config-from-path", zap.String("config", flagconf))
+
 	c := config.New(
 		config.WithSource(
 			file.NewSource(flagconf),
@@ -68,7 +96,7 @@
 	var cfg conf.Bootstrap
 	must.Done(c.Scan(&cfg))
 
-	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
+	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, zapKratos))
 	defer cleanup()
 
 	// start and wait for stop signal
```

## cmd/demo2kratos/wire.go (+2 -2)

```diff
@@ -6,16 +6,16 @@
 
 import (
 	"github.com/go-kratos/kratos/v2"
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/google/wire"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/server"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
 // wireApp init kratos application.
-func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
+func wireApp(*conf.Server, *conf.Data, *zapkratos.ZapKratos) (*kratos.App, func(), error) {
 	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
 }
```

## cmd/demo2kratos/wire_gen.go (+8 -8)

```diff
@@ -7,27 +7,27 @@
 
 import (
 	"github.com/go-kratos/kratos/v2"
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/server"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
 // Injectors from wire.go:
 
 // wireApp init kratos application.
-func wireApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
-	dataData, cleanup, err := data.NewData(confData, logger)
+func wireApp(confServer *conf.Server, confData *conf.Data, zapKratos *zapkratos.ZapKratos) (*kratos.App, func(), error) {
+	dataData, cleanup, err := data.NewData(confData, zapKratos)
 	if err != nil {
 		return nil, nil, err
 	}
-	articleUsecase := biz.NewArticleUsecase(dataData, logger)
-	articleService := service.NewArticleService(articleUsecase)
-	grpcServer := server.NewGRPCServer(confServer, articleService, logger)
-	httpServer := server.NewHTTPServer(confServer, articleService, logger)
-	app := newApp(logger, grpcServer, httpServer)
+	articleUsecase := biz.NewArticleUsecase(dataData, zapKratos)
+	articleService := service.NewArticleService(articleUsecase, zapKratos)
+	grpcServer := server.NewGRPCServer(confServer, articleService, zapKratos)
+	httpServer := server.NewHTTPServer(confServer, articleService, zapKratos)
+	app := newApp(grpcServer, httpServer, zapKratos)
 	return app, func() {
 		cleanup()
 	}, nil
```

## internal/biz/article.go (+10 -5)

```diff
@@ -4,10 +4,11 @@
 	"context"
 
 	"github.com/brianvoe/gofakeit/v7"
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
+	"github.com/yylego/kratos-zap/zapkratos"
+	"github.com/yylego/zaplog"
 )
 
 type Article struct {
@@ -18,15 +19,19 @@
 }
 
 type ArticleUsecase struct {
-	data *data.Data
-	log  *log.Helper
+	data   *data.Data
+	zapLog *zaplog.Zap
 }
 
-func NewArticleUsecase(data *data.Data, logger log.Logger) *ArticleUsecase {
-	return &ArticleUsecase{data: data, log: log.NewHelper(logger)}
+func NewArticleUsecase(data *data.Data, zapKratos *zapkratos.ZapKratos) *ArticleUsecase {
+	return &ArticleUsecase{
+		data:   data,
+		zapLog: zapKratos.SubZap(),
+	}
 }
 
 func (uc *ArticleUsecase) CreateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
+	uc.zapLog.SUG.Infof("CreateArticle: %v", a)
 	var res Article
 	if err := gofakeit.Struct(&res); err != nil {
 		return nil, ebzkratos.New(pb.ErrorArticleCreateFailure("fake: %v", err))
```

## internal/data/data.go (+5 -3)

```diff
@@ -1,9 +1,9 @@
 package data
 
 import (
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/google/wire"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
+	"github.com/yylego/kratos-zap/zapkratos"
 	"github.com/yylego/must"
 	"github.com/yylego/rese"
 	"gorm.io/driver/sqlite"
@@ -16,11 +16,13 @@
 	db *gorm.DB
 }
 
-func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
+func NewData(c *conf.Data, zapKratos *zapkratos.ZapKratos) (*Data, func(), error) {
+	zapLog := zapKratos.SubZap()
+	zapLog.SUG.Info("creating data resources")
 	must.Same(c.Database.Driver, "sqlite3")
 	db := rese.P1(gorm.Open(sqlite.Open(c.Database.Source), &gorm.Config{}))
 	cleanup := func() {
-		log.NewHelper(logger).Info("closing the data resources")
+		zapLog.SUG.Info("closing the data resources")
 		_ = rese.P1(db.DB()).Close()
 	}
 	return &Data{db: db}, cleanup, nil
```

## internal/server/grpc.go (+4 -2)

```diff
@@ -1,18 +1,20 @@
 package server
 
 import (
-	"github.com/go-kratos/kratos/v2/log"
+	"github.com/go-kratos/kratos/v2/middleware/logging"
 	"github.com/go-kratos/kratos/v2/middleware/recovery"
 	"github.com/go-kratos/kratos/v2/transport/grpc"
 	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
-func NewGRPCServer(c *conf.Server, article *service.ArticleService, logger log.Logger) *grpc.Server {
+func NewGRPCServer(c *conf.Server, article *service.ArticleService, zapKratos *zapkratos.ZapKratos) *grpc.Server {
 	var opts = []grpc.ServerOption{
 		grpc.Middleware(
 			recovery.Recovery(),
+			logging.Server(zapKratos.GetLogger("grpc-request")),
 		),
 	}
 	if c.Grpc.Network != "" {
```

## internal/server/http.go (+4 -2)

```diff
@@ -1,18 +1,20 @@
 package server
 
 import (
-	"github.com/go-kratos/kratos/v2/log"
+	"github.com/go-kratos/kratos/v2/middleware/logging"
 	"github.com/go-kratos/kratos/v2/middleware/recovery"
 	"github.com/go-kratos/kratos/v2/transport/http"
 	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
-func NewHTTPServer(c *conf.Server, article *service.ArticleService, logger log.Logger) *http.Server {
+func NewHTTPServer(c *conf.Server, article *service.ArticleService, zapKratos *zapkratos.ZapKratos) *http.Server {
 	var opts = []http.ServerOption{
 		http.Middleware(
 			recovery.Recovery(),
+			logging.Server(zapKratos.GetLogger("http-request")),
 		),
 	}
 	if c.Http.Network != "" {
```

## internal/service/article.go (+12 -3)

```diff
@@ -5,23 +5,32 @@
 
 	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
+	"github.com/yylego/kratos-zap/zapkratos"
+	"github.com/yylego/zaplog"
+	"go.uber.org/zap"
 )
 
 type ArticleService struct {
 	pb.UnimplementedArticleServiceServer
 
-	uc *biz.ArticleUsecase
+	uc     *biz.ArticleUsecase
+	zapLog *zaplog.Zap
 }
 
-func NewArticleService(uc *biz.ArticleUsecase) *ArticleService {
-	return &ArticleService{uc: uc}
+func NewArticleService(uc *biz.ArticleUsecase, zapKratos *zapkratos.ZapKratos) *ArticleService {
+	return &ArticleService{
+		uc:     uc,
+		zapLog: zapKratos.SubZap(),
+	}
 }
 
 func (s *ArticleService) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleReply, error) {
+	s.zapLog.LOG.Info("receive-create-article-message")
 	v, ebz := s.uc.CreateArticle(ctx, nil)
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
+	s.zapLog.LOG.Info("reply-create-article-message", zap.Int64("id", v.ID))
 	return &pb.CreateArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.Title, Content: v.Content, StudentId: v.StudentID}}, nil
 }
 
```

