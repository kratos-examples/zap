package biz

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/yylego/kratos-ebz/ebzkratos"
	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
	"github.com/yylego/kratos-zap/zapkratos"
	"github.com/yylego/zaplog"
)

type Student struct {
	ID        int64
	Name      string
	Age       int32
	ClassName string
}

type StudentUsecase struct {
	data   *data.Data
	zapLog *zaplog.Zap
}

func NewStudentUsecase(data *data.Data, zapKratos *zapkratos.ZapKratos) *StudentUsecase {
	return &StudentUsecase{
		data:   data,
		zapLog: zapKratos.SubZap(),
	}
}

func (uc *StudentUsecase) CreateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
	uc.zapLog.SUG.Infof("CreateStudent: %v", s)
	var res Student
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("fake: %v", err))
	}
	return &res, nil
}

func (uc *StudentUsecase) UpdateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
	var res Student
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorServerError("fake: %v", err))
	}
	return &res, nil
}

func (uc *StudentUsecase) DeleteStudent(ctx context.Context, id int64) *ebzkratos.Ebz {
	return nil
}

func (uc *StudentUsecase) GetStudent(ctx context.Context, id int64) (*Student, *ebzkratos.Ebz) {
	var res Student
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorServerError("fake: %v", err))
	}
	return &res, nil
}

func (uc *StudentUsecase) ListStudents(ctx context.Context, page int32, pageSize int32) ([]*Student, int32, *ebzkratos.Ebz) {
	var items []*Student
	gofakeit.Slice(&items)
	return items, int32(len(items)), nil
}
