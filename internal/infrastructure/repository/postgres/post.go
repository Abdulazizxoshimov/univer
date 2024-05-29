package postgres

import (
	"context"
	"fmt"
	"strings"
	"univer/internal/entity"
	"univer/internal/pkg/otlp"
	postgres "univer/internal/pkg/storage"

	"github.com/Masterminds/squirrel"
)

const (
	viewsTableName          = "views"
	postServiceTableName    = "posts"
	serviceNamePostsService = "postServiceRepo"
	spanNamePostsService    = "postSpanRepo"
)

type postRepo struct {
	tableName string
	db        *postgres.PostgresDB
}

func NewPostRepo(db *postgres.PostgresDB) *postRepo {
	return &postRepo{
		tableName: postServiceTableName,
		db:        db,
	}
}

func (p *postRepo) postsSelectQueryPrefix() squirrel.SelectBuilder {
	return p.db.Sq.Builder.
		Select(
			"id",
			"user_id",
			"theme",
			"path",
			"views",
			"science",
			"category_id",
			"price_status",
			"price",
			"created_at",
			"updated_at",
		).From(p.tableName)
}

func (p postRepo) CreatePost(ctx context.Context, post *entity.Post) (*entity.Post, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService+"CreatePost")
	defer span.End()

	data := map[string]any{
		"id":           post.Id,
		"user_id":      post.UserId,
		"theme":        post.Theme,
		"path":         post.Path,
		"views":        post.Views,
		"science":      post.Science,
		"category_id":  post.CategoryId,
		"price_status": post.PriceStatus,
		"price":        post.Price,
		"created_at":   post.CreatedAt,
		"updated_at":   post.UpdatedAt,
	}
	query, args, err := p.db.Sq.Builder.Insert(p.tableName).SetMap(data).ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "create"))
	}

	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, p.db.Error(err)
	}

	return post, nil
}

func (p postRepo) UpdatePost(ctx context.Context, post *entity.PostUpdateReq) (*entity.PostUpdateReq, error) {

	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService+"UpdatePost")
	defer span.End()

	clauses := map[string]any{
		"theme":        post.Theme,
		"path":         post.Path,
		"science":      post.Science,
		"category_id":  post.CategoryId,
		"price_status": post.PriceStatus,
		"price":        post.Price,
		"updated_at":   post.UpdatedAt,
	}
	sqlStr, args, err := p.db.Sq.Builder.
		Update(p.tableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("id", post.Id)).
		Where("deleted_at is null").
		ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, p.tableName+" update")
	}

	commandTag, err := p.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return nil, p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return nil, p.db.Error(fmt.Errorf("no sql rows"))
	}

	return post, nil
}

func (p postRepo) DeletePost(ctx context.Context, req *entity.DeleteReq) error {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService+"DeletePost")
	defer span.End()

	clauses := map[string]interface{}{
		"deleted_at": req.DeletedAt,
	}

	sqlStr, args, err := p.db.Sq.Builder.
		Update(p.tableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("id", req.Id)).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return p.db.ErrSQLBuild(err, p.tableName+" delete")
	}

	commandTag, err := p.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return p.db.Error(fmt.Errorf("no sql rows"))
	}

	return nil
}

func (p postRepo) GetPost(ctx context.Context, params map[string]string) (*entity.Post, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService+"GetPost")
	defer span.End()

	var (
		post entity.Post
		cnt  int
	)

	queryBuilder := p.postsSelectQueryPrefix()

	for key, value := range params {
		if key == "id" {
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal(key, value))
		} else if key == "del" {
			cnt++
		}
	}
	if cnt == 0 {
		queryBuilder = queryBuilder.Where("deleted_at is null")
	}
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "get"))
	}

	if err = p.db.QueryRow(ctx, query, args...).Scan(
		&post.Id,
		&post.UserId,
		&post.Theme,
		&post.Path,
		&post.Views,
		&post.Science,
		&post.CategoryId,
		&post.PriceStatus,
		&post.Price,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		return nil, p.db.Error(err)
	}

	return &post, nil
}

func (p postRepo) ListPost(ctx context.Context, limit int, offset int, filter map[string]string) (*entity.PostListRes, error) {

	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService+"ListPost")
	defer span.End()

	var (
		posts entity.PostListRes
	)
	queryBuilder := p.postsSelectQueryPrefix()

	if limit != 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit)).Offset(uint64(offset))
	}
	for key, value := range filter {
		if key == "user_id" {
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal(key, value))
		}
		if key == "role" {
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal(key, value))

		}
	}

	queryBuilder = queryBuilder.Where("deleted_at IS NULL").OrderBy("created_at")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "list"))
	}

	rows, err := p.db.Query(ctx, query, args...)
	if err != nil {
		return nil, p.db.Error(err)
	}
	defer rows.Close()
	var post entity.Post

	for rows.Next() {

		if err = rows.Scan(
			&post.Id,
			&post.UserId,
			&post.Theme,
			&post.Path,
			&post.Views,
			&post.Science,
			&post.CategoryId,
			&post.PriceStatus,
			&post.Price,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, p.db.Error(err)
		}

		posts.Post = append(posts.Post, &post)
	}

	var count uint64

	queryBuilder = p.db.Sq.Builder.Select("COUNT(*)").
		From(p.tableName).
		Where("deleted_at is null")

	query, _, err = queryBuilder.ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "list"))
	}

	if err := p.db.QueryRow(ctx, query).Scan(&count); err != nil {
		posts.TotalCount = 0
	}
	posts.TotalCount = int(count)

	return &posts, nil
}

func (p postRepo) Search(ctx context.Context, req *entity.ListReq) (*entity.PostListRes, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService+"Search")
	defer span.End()

	terms := strings.Fields(req.Filter["theme"])
	likeClause := "%" + strings.Join(terms, "%") + "%"
	queryBuilder := p.db.Sq.Builder.
	Select(
		"posts.id",
		"posts.user_id",
		"posts.theme",
		"posts.path",
		"posts.views",
		"posts.science",
		"category.name",
		"posts.price_status",
		"posts.price",
		"posts.created_at",
		"posts.updated_at",
	).From(p.tableName).
	Join(categoryServiceTableName + " on posts.category_id = category.id")
	for key, value := range req.Filter{
		if key == "theme"{
			queryBuilder = queryBuilder.Where(p.db.Sq.ILike("posts." + key, likeClause))
		}else if key == "category_id"{
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal("posts." + key, value))
		}else if key == "science"{
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal("posts." + key, value))
		}else if key == "price_status"{
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal("posts." + key, value))
		}
	}
	queryBuilder = queryBuilder.Where("deleted_at is null")


	if req.Limit != 0 {
		queryBuilder = queryBuilder.Limit(uint64(req.Limit)).Offset(uint64(req.Offset))
	}
	sqlStr, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("SQL build error: %w", err)
	}
	rows, err := p.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, p.db.Error(err)
	}
	defer rows.Close()

	posts := entity.PostListRes{}
	for rows.Next() {
		var post entity.Post
		err = rows.Scan(
			&post.Id,
			&post.UserId,
			&post.Theme,
			&post.Path,
			&post.Views,
			&post.Science,
			&post.CategoryId,
			&post.PriceStatus,
			&post.Price,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, p.db.Error(err)
		}
		posts.Post = append(posts.Post, &post)
	}
	var count uint64

	queryBuilder = p.db.Sq.Builder.Select("COUNT(*)").
		From(p.tableName).
		Where("deleted_at is null")

	query, _, err := queryBuilder.ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "list"))
	}

	if err := p.db.QueryRow(ctx, query).Scan(&count); err != nil {
		posts.TotalCount = 0
	}
	posts.TotalCount = int(count)

	return &posts, nil

}

func (p postRepo) CheckUnique(ctx context.Context, UserId, PostId string) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService+"checkUnique")
	defer span.End()

	queryBuilder := p.db.Sq.Builder.Select("COUNT(1)").
		From(viewsTableName).
		Where(squirrel.Eq{"post_id": PostId, "user_id": UserId})

	query, args, err := queryBuilder.ToSql()

	if err != nil {
		return false, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", likeTableName, "checkUnique"))
	}

	var count int

	if err = p.db.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return false, p.db.Error(err)

	}
	if count != 0 {
		return true, nil
	}
	return false, nil
}

func (p postRepo) CreateViews(ctx context.Context, userId, postId string) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService+"CreateViews")
	defer span.End()
	clauses := map[string]any{
		"user_id": userId,
		"post_id": postId,
	}
	query, args, err := p.db.Sq.Builder.Insert(viewsTableName).SetMap(clauses).ToSql()
	if err != nil {
		return false, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "create"))
	}
	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		return false, p.db.Error(err)
	}

	return true, nil
}
func (p postRepo) UpdateViews(ctx context.Context, postId string) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNamePostsService, spanNamePostsService+"UpdateViews")
	defer span.End()

	query := `UPDATE posts SET views = views + 1 WHERE id = $1`

	commandTag, err := p.db.Exec(ctx, query, postId)
	if err != nil {
		return false, p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return false, p.db.Error(fmt.Errorf("no sql rows"))
	}

	return true, nil
}
