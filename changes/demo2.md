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
 
 .PHONY: init
 # init env
```

## cmd/demo2kratos/main.go (+12 -18)

```diff
@@ -2,20 +2,20 @@
 
 import (
 	"flag"
-	"log/slog"
 	"os"
 
-	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
 	"github.com/go-kratos/kratos/v3"
 	"github.com/go-kratos/kratos/v3/config"
 	"github.com/go-kratos/kratos/v3/config/file"
-	"github.com/go-kratos/kratos/v3/log"
 	"github.com/go-kratos/kratos/v3/transport/grpc"
 	"github.com/go-kratos/kratos/v3/transport/http"
 	"github.com/yylego/done"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
+	"github.com/yylego/kratos-zap/zapkratos"
 	"github.com/yylego/must"
 	"github.com/yylego/rese"
+	"github.com/yylego/zaplog"
+	"go.uber.org/zap"
 
 	_ "go.uber.org/automaxprocs"
 )
@@ -34,13 +34,13 @@
 	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
 }
 
-func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
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
@@ -50,18 +50,12 @@
 
 func main() {
 	flag.Parse()
-	logger := log.NewLogger(
-		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
-			AddSource: true,
-			Level:     slog.LevelInfo,
-		}),
-		log.WithExtractor(tracing.TraceAttrs),
-	).With(
-		slog.String("service.id", done.VCE(os.Hostname()).Omit()),
-		slog.String("service.name", Name),
-		slog.String("service.version", Version),
-	)
-	log.SetDefault(logger)
+
+	zapKratos := zapkratos.NewZapKratos(zaplog.LOGGER, zapkratos.NewOptions())
+	zapLog := zapKratos.SubZap()
+	zapLog.LOG.Info("application starting...")
+	zapLog.LOG.Info("reading-config-from-path", zap.String("config", flagconf))
+
 	c := config.New(
 		config.WithSource(
 			file.NewSource(flagconf),
@@ -74,7 +68,7 @@
 	var cfg conf.Bootstrap
 	must.Done(c.Scan(&cfg))
 
-	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
+	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, zapKratos))
 	defer cleanup()
 
 	// start and wait for stop signal
```

## cmd/demo2kratos/wire.go (+2 -3)

```diff
@@ -5,8 +5,6 @@
 package main
 
 import (
-	"log/slog"
-
 	"github.com/go-kratos/kratos/v3"
 	"github.com/google/wire"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
@@ -14,9 +12,10 @@
 	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/server"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
 // wireApp init kratos application.
-func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
+func wireApp(*conf.Server, *conf.Data, *zapkratos.ZapKratos) (*kratos.App, func(), error) {
 	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
 }
```

## cmd/demo2kratos/wire_gen.go (+8 -8)

```diff
@@ -13,7 +13,7 @@
 	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/server"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
-	"log/slog"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
 import (
@@ -23,20 +23,20 @@
 // Injectors from wire.go:
 
 // wireApp init kratos application.
-func wireApp(confServer *conf.Server, confData *conf.Data, logger *slog.Logger) (*kratos.App, func(), error) {
-	dataData, cleanup, err := data.NewData(confData, logger)
+func wireApp(confServer *conf.Server, confData *conf.Data, zapKratos *zapkratos.ZapKratos) (*kratos.App, func(), error) {
+	dataData, cleanup, err := data.NewData(confData, zapKratos)
 	if err != nil {
 		return nil, nil, err
 	}
-	articleUsecase, err := biz.NewArticleUsecase(dataData, logger)
+	articleUsecase, err := biz.NewArticleUsecase(dataData, zapKratos)
 	if err != nil {
 		cleanup()
 		return nil, nil, err
 	}
-	articleService := service.NewArticleService(articleUsecase)
-	grpcServer := server.NewGRPCServer(confServer, articleService, logger)
-	httpServer := server.NewHTTPServer(confServer, articleService, logger)
-	app := newApp(logger, grpcServer, httpServer)
+	articleService := service.NewArticleService(articleUsecase, zapKratos)
+	grpcServer := server.NewGRPCServer(confServer, articleService, zapKratos)
+	httpServer := server.NewHTTPServer(confServer, articleService, zapKratos)
+	app := newApp(grpcServer, httpServer, zapKratos)
 	return app, func() {
 		cleanup()
 	}, nil
```

## internal/biz/article.go (+8 -7)

```diff
@@ -3,12 +3,13 @@
 import (
 	"context"
 	"errors"
-	"log/slog"
 
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
+	"github.com/yylego/kratos-zap/zapkratos"
 	"github.com/yylego/must"
+	"github.com/yylego/zaplog"
 	"gorm.io/gorm"
 	"gorm.io/gorm/clause"
 )
@@ -29,18 +30,18 @@
 func (Article) TableName() string { return "articles" }
 
 type ArticleUsecase struct {
-	data *data.Data
-	slog *slog.Logger
+	data   *data.Data
+	zapLog *zaplog.Zap
 }
 
-func NewArticleUsecase(data *data.Data, logger *slog.Logger) (*ArticleUsecase, error) {
+func NewArticleUsecase(data *data.Data, zapKratos *zapkratos.ZapKratos) (*ArticleUsecase, error) {
 	// Migrate the owned table plus the mirrored students table (needed in the
 	// existence check); both services share one database
 	// 建好本服务拥有的 articles 表，外加镜像的 students 表（供存在性校验用）
 	if err := data.DB().AutoMigrate(&Article{}, &Student{}); err != nil {
 		return nil, err
 	}
-	return &ArticleUsecase{data: data, slog: logger}, nil
+	return &ArticleUsecase{data: data, zapLog: zapKratos.SubZap()}, nil
 }
 
 func (uc *ArticleUsecase) CreateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
@@ -68,7 +69,7 @@
 		}
 		return nil, ebzkratos.New(pb.ErrorArticleCreateFailure("create article: %v", err))
 	}
-	uc.slog.InfoContext(ctx, "created article", "id", res.ID, "student_id", res.StudentID)
+	uc.zapLog.SUG.Infow("created article", "id", res.ID, "student_id", res.StudentID)
 	return res, nil
 }
 
@@ -127,7 +128,7 @@
 	if del.RowsAffected == 0 {
 		return ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", id))
 	}
-	uc.slog.InfoContext(ctx, "deleted article", "id", id)
+	uc.zapLog.SUG.Infow("deleted article", "id", id)
 	return nil
 }
 
```

## internal/data/data.go (+5 -4)

```diff
@@ -1,10 +1,9 @@
 package data
 
 import (
-	"log/slog"
-
 	"github.com/google/wire"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
+	"github.com/yylego/kratos-zap/zapkratos"
 	"github.com/yylego/must"
 	"github.com/yylego/rese"
 	"gorm.io/driver/postgres"
@@ -24,11 +23,13 @@
 	return d.db
 }
 
-func NewData(c *conf.Data, logger *slog.Logger) (*Data, func(), error) {
+func NewData(c *conf.Data, zapKratos *zapkratos.ZapKratos) (*Data, func(), error) {
+	zapLog := zapKratos.SubZap()
+	zapLog.SUG.Info("creating data resources")
 	must.Same(c.Database.Driver, "postgres")
 	db := rese.P1(gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{}))
 	cleanup := func() {
-		logger.Info("closing the data resources")
+		zapLog.SUG.Info("closing the data resources")
 		_ = rese.P1(db.DB()).Close()
 	}
 	return &Data{db: db}, cleanup, nil
```

## internal/server/grpc.go (+4 -3)

```diff
@@ -1,19 +1,20 @@
 package server
 
 import (
-	"log/slog"
-
+	"github.com/go-kratos/kratos/v3/middleware/logging"
 	"github.com/go-kratos/kratos/v3/middleware/recovery"
 	"github.com/go-kratos/kratos/v3/transport/grpc"
 	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
-func NewGRPCServer(c *conf.Server, article *service.ArticleService, logger *slog.Logger) *grpc.Server {
+func NewGRPCServer(c *conf.Server, article *service.ArticleService, zapKratos *zapkratos.ZapKratos) *grpc.Server {
 	var opts = []grpc.ServerOption{
 		grpc.Middleware(
 			recovery.Recovery(),
+			logging.Server(zapKratos.GetLogger("grpc-request")),
 		),
 	}
 	if c.Grpc.Network != "" {
```

## internal/server/http.go (+4 -3)

```diff
@@ -1,19 +1,20 @@
 package server
 
 import (
-	"log/slog"
-
+	"github.com/go-kratos/kratos/v3/middleware/logging"
 	"github.com/go-kratos/kratos/v3/middleware/recovery"
 	"github.com/go-kratos/kratos/v3/transport/http"
 	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
-func NewHTTPServer(c *conf.Server, article *service.ArticleService, logger *slog.Logger) *http.Server {
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
@@ -5,19 +5,27 @@
 
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
 	if req.Title == "" {
 		return nil, pb.ErrorBadParam("TITLE IS REQUIRED")
 	}
@@ -32,6 +40,7 @@
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
+	s.zapLog.LOG.Info("reply-create-article-message", zap.Int64("id", v.ID))
 	return &pb.CreateArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.Title, Content: v.Content, StudentId: v.StudentID}}, nil
 }
 
```

