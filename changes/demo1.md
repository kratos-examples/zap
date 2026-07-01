# Changes

Code differences compared to source project.

## cmd/demo1kratos/main.go (+12 -18)

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
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
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

## cmd/demo1kratos/wire.go (+2 -3)

```diff
@@ -5,8 +5,6 @@
 package main
 
 import (
-	"log/slog"
-
 	"github.com/go-kratos/kratos/v3"
 	"github.com/google/wire"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
@@ -14,9 +12,10 @@
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/server"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
 // wireApp init kratos application.
-func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
+func wireApp(*conf.Server, *conf.Data, *zapkratos.ZapKratos) (*kratos.App, func(), error) {
 	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
 }
```

## cmd/demo1kratos/wire_gen.go (+8 -8)

```diff
@@ -13,7 +13,7 @@
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/server"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
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
-	studentUsecase, err := biz.NewStudentUsecase(dataData, logger)
+	studentUsecase, err := biz.NewStudentUsecase(dataData, zapKratos)
 	if err != nil {
 		cleanup()
 		return nil, nil, err
 	}
-	studentService := service.NewStudentService(studentUsecase)
-	grpcServer := server.NewGRPCServer(confServer, studentService, logger)
-	httpServer := server.NewHTTPServer(confServer, studentService, logger)
-	app := newApp(logger, grpcServer, httpServer)
+	studentService := service.NewStudentService(studentUsecase, zapKratos)
+	grpcServer := server.NewGRPCServer(confServer, studentService, zapKratos)
+	httpServer := server.NewHTTPServer(confServer, studentService, zapKratos)
+	app := newApp(grpcServer, httpServer, zapKratos)
 	return app, func() {
 		cleanup()
 	}, nil
```

## internal/biz/student.go (+8 -7)

```diff
@@ -3,12 +3,13 @@
 import (
 	"context"
 	"errors"
-	"log/slog"
 
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
+	"github.com/yylego/kratos-zap/zapkratos"
 	"github.com/yylego/must"
+	"github.com/yylego/zaplog"
 	"gorm.io/gorm"
 	"gorm.io/gorm/clause"
 )
@@ -29,17 +30,17 @@
 // 用于级联删除的 Article 镜像模型定义在 article.go 中。
 
 type StudentUsecase struct {
-	data *data.Data
-	slog *slog.Logger
+	data   *data.Data
+	zapLog *zaplog.Zap
 }
 
-func NewStudentUsecase(data *data.Data, logger *slog.Logger) (*StudentUsecase, error) {
+func NewStudentUsecase(data *data.Data, zapKratos *zapkratos.ZapKratos) (*StudentUsecase, error) {
 	// Share one database with the article service: keep both tables in sync here
 	// 与文章服务共用一个库：在这里把两张表都建好
 	if err := data.DB().AutoMigrate(&Student{}, &Article{}); err != nil {
 		return nil, err
 	}
-	return &StudentUsecase{data: data, slog: logger}, nil
+	return &StudentUsecase{data: data, zapLog: zapKratos.SubZap()}, nil
 }
 
 func (uc *StudentUsecase) CreateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
@@ -49,7 +50,7 @@
 	if err := uc.data.DB().WithContext(ctx).Create(res).Error; err != nil {
 		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("create student: %v", err))
 	}
-	uc.slog.InfoContext(ctx, "created student", "id", res.ID, "name", res.Name)
+	uc.zapLog.SUG.Infow("created student", "id", res.ID, "name", res.Name)
 	return res, nil
 }
 
@@ -113,7 +114,7 @@
 	if notFound {
 		return ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", id))
 	}
-	uc.slog.InfoContext(ctx, "deleted student and cascaded articles", "student_id", id, "articles_removed", removedArticles)
+	uc.zapLog.SUG.Infow("deleted student and cascaded articles", "student_id", id, "articles_removed", removedArticles)
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
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
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
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
-func NewGRPCServer(c *conf.Server, student *service.StudentService, logger *slog.Logger) *grpc.Server {
+func NewGRPCServer(c *conf.Server, student *service.StudentService, zapKratos *zapkratos.ZapKratos) *grpc.Server {
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
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
-func NewHTTPServer(c *conf.Server, student *service.StudentService, logger *slog.Logger) *http.Server {
+func NewHTTPServer(c *conf.Server, student *service.StudentService, zapKratos *zapkratos.ZapKratos) *http.Server {
 	var opts = []http.ServerOption{
 		http.Middleware(
 			recovery.Recovery(),
+			logging.Server(zapKratos.GetLogger("http-request")),
 		),
 	}
 	if c.Http.Network != "" {
```

## internal/service/student.go (+12 -3)

```diff
@@ -5,19 +5,27 @@
 
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
+	"github.com/yylego/kratos-zap/zapkratos"
+	"github.com/yylego/zaplog"
+	"go.uber.org/zap"
 )
 
 type StudentService struct {
 	pb.UnimplementedStudentServiceServer
 
-	uc *biz.StudentUsecase
+	uc     *biz.StudentUsecase
+	zapLog *zaplog.Zap
 }
 
-func NewStudentService(uc *biz.StudentUsecase) *StudentService {
-	return &StudentService{uc: uc}
+func NewStudentService(uc *biz.StudentUsecase, zapKratos *zapkratos.ZapKratos) *StudentService {
+	return &StudentService{
+		uc:     uc,
+		zapLog: zapKratos.SubZap(),
+	}
 }
 
 func (s *StudentService) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentReply, error) {
+	s.zapLog.LOG.Info("receive-create-student-message")
 	if req.Name == "" {
 		return nil, pb.ErrorBadParam("NAME IS REQUIRED")
 	}
@@ -29,6 +37,7 @@
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
+	s.zapLog.LOG.Info("reply-create-student-message", zap.Int64("id", v.ID))
 	return &pb.CreateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName}}, nil
 }
 
```

