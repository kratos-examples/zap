# Changes

Code differences compared to source project.

## cmd/demo1kratos/main.go (+12 -14)

```diff
@@ -7,14 +7,15 @@
 	"github.com/go-kratos/kratos/v2"
 	"github.com/go-kratos/kratos/v2/config"
 	"github.com/go-kratos/kratos/v2/config/file"
-	"github.com/go-kratos/kratos/v2/log"
-	"github.com/go-kratos/kratos/v2/middleware/tracing"
 	"github.com/go-kratos/kratos/v2/transport/grpc"
 	"github.com/go-kratos/kratos/v2/transport/http"
 	"github.com/yylego/done"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
+	"github.com/yylego/kratos-zap/zapkratos"
 	"github.com/yylego/must"
 	"github.com/yylego/rese"
+	"github.com/yylego/zaplog"
+	"go.uber.org/zap"
 )
 
 // go build -ldflags "-X main.Version=x.y.z"
@@ -31,13 +32,13 @@
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
@@ -47,15 +48,12 @@
 
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
+	zapKratos := zapkratos.NewZapKratos(zaplog.LOGGER, zapkratos.NewOptions())
+	zapLog := zapKratos.SubZap()
+	zapLog.LOG.Info("application starting...")
+	zapLog.LOG.Info("reading-config-from-path", zap.String("config", flagconf))
+
 	c := config.New(
 		config.WithSource(
 			file.NewSource(flagconf),
@@ -68,7 +66,7 @@
 	var cfg conf.Bootstrap
 	must.Done(c.Scan(&cfg))
 
-	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
+	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, zapKratos))
 	defer cleanup()
 
 	// start and wait for stop signal
```

## cmd/demo1kratos/wire.go (+2 -2)

```diff
@@ -6,16 +6,16 @@
 
 import (
 	"github.com/go-kratos/kratos/v2"
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/google/wire"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/server"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
 // wireApp init kratos application.
-func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
+func wireApp(*conf.Server, *conf.Data, *zapkratos.ZapKratos) (*kratos.App, func(), error) {
 	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
 }
```

## cmd/demo1kratos/wire_gen.go (+8 -8)

```diff
@@ -7,27 +7,27 @@
 
 import (
 	"github.com/go-kratos/kratos/v2"
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/server"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
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
-	studentUsecase := biz.NewStudentUsecase(dataData, logger)
-	studentService := service.NewStudentService(studentUsecase)
-	grpcServer := server.NewGRPCServer(confServer, studentService, logger)
-	httpServer := server.NewHTTPServer(confServer, studentService, logger)
-	app := newApp(logger, grpcServer, httpServer)
+	studentUsecase := biz.NewStudentUsecase(dataData, zapKratos)
+	studentService := service.NewStudentService(studentUsecase, zapKratos)
+	grpcServer := server.NewGRPCServer(confServer, studentService, zapKratos)
+	httpServer := server.NewHTTPServer(confServer, studentService, zapKratos)
+	app := newApp(grpcServer, httpServer, zapKratos)
 	return app, func() {
 		cleanup()
 	}, nil
```

## internal/biz/student.go (+10 -5)

```diff
@@ -4,10 +4,11 @@
 	"context"
 
 	"github.com/brianvoe/gofakeit/v7"
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
+	"github.com/yylego/kratos-zap/zapkratos"
+	"github.com/yylego/zaplog"
 )
 
 type Student struct {
@@ -18,15 +19,19 @@
 }
 
 type StudentUsecase struct {
-	data *data.Data
-	log  *log.Helper
+	data   *data.Data
+	zapLog *zaplog.Zap
 }
 
-func NewStudentUsecase(data *data.Data, logger log.Logger) *StudentUsecase {
-	return &StudentUsecase{data: data, log: log.NewHelper(logger)}
+func NewStudentUsecase(data *data.Data, zapKratos *zapkratos.ZapKratos) *StudentUsecase {
+	return &StudentUsecase{
+		data:   data,
+		zapLog: zapKratos.SubZap(),
+	}
 }
 
 func (uc *StudentUsecase) CreateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
+	uc.zapLog.SUG.Infof("CreateStudent: %v", s)
 	var res Student
 	if err := gofakeit.Struct(&res); err != nil {
 		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("fake: %v", err))
```

## internal/data/data.go (+5 -3)

```diff
@@ -1,9 +1,9 @@
 package data
 
 import (
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/google/wire"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
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
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
-func NewGRPCServer(c *conf.Server, student *service.StudentService, logger log.Logger) *grpc.Server {
+func NewGRPCServer(c *conf.Server, student *service.StudentService, zapKratos *zapkratos.ZapKratos) *grpc.Server {
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
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
+	"github.com/yylego/kratos-zap/zapkratos"
 )
 
-func NewHTTPServer(c *conf.Server, student *service.StudentService, logger log.Logger) *http.Server {
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
@@ -5,23 +5,32 @@
 
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
 	v, ebz := s.uc.CreateStudent(ctx, nil)
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
+	s.zapLog.LOG.Info("reply-create-student-message", zap.Int64("id", v.ID))
 	return &pb.CreateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName}}, nil
 }
 
```

