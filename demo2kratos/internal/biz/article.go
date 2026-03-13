package biz

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/yylego/kratos-ebz/ebzkratos"
	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
	"github.com/yylego/kratos-zap/zapkratos"
	"github.com/yylego/zaplog"
)

type Article struct {
	ID        int64
	Title     string
	Content   string
	StudentID int64
}

type ArticleUsecase struct {
	data   *data.Data
	zapLog *zaplog.Zap
}

func NewArticleUsecase(data *data.Data, zapKratos *zapkratos.ZapKratos) *ArticleUsecase {
	return &ArticleUsecase{
		data:   data,
		zapLog: zapKratos.SubZap(),
	}
}

func (uc *ArticleUsecase) CreateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
	uc.zapLog.SUG.Infof("CreateArticle: %v", a)
	var res Article
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorArticleCreateFailure("fake: %v", err))
	}
	return &res, nil
}

func (uc *ArticleUsecase) UpdateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
	var res Article
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorServerError("fake: %v", err))
	}
	return &res, nil
}

func (uc *ArticleUsecase) DeleteArticle(ctx context.Context, id int64) *ebzkratos.Ebz {
	return nil
}

func (uc *ArticleUsecase) GetArticle(ctx context.Context, id int64) (*Article, *ebzkratos.Ebz) {
	var res Article
	if err := gofakeit.Struct(&res); err != nil {
		return nil, ebzkratos.New(pb.ErrorServerError("fake: %v", err))
	}
	return &res, nil
}

func (uc *ArticleUsecase) ListArticles(ctx context.Context, page int32, pageSize int32) ([]*Article, int32, *ebzkratos.Ebz) {
	var items []*Article
	gofakeit.Slice(&items)
	return items, int32(len(items)), nil
}
