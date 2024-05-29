package postgres

import (
	"context"
	"fmt"
	"univer/internal/entity"
	"univer/internal/pkg/otlp"
	postgres "univer/internal/pkg/storage"

	"github.com/Masterminds/squirrel"
)

const (
	likeTableName             = "likes"
	commentServiceTableName   = "comments"
	serviceNameCommentService = "commentServiceRepo"
	spanNameCommentService    = "commentSpanRepo"
)

type commentRepo struct {
	tableName string
	db        *postgres.PostgresDB
}

func NewCommentRepo(db *postgres.PostgresDB) *commentRepo {
	return &commentRepo{
		tableName: commentServiceTableName,
		db:        db,
	}
}

func (p commentRepo) comentSelectQueryPrefix() squirrel.SelectBuilder {
	return p.db.Sq.Builder.
		Select(
			"id",
			"post_id",
			"owner_id",
			"message",
			"likes",
			"dislikes",
			"created_at",
			"updated_at",
		).From(p.tableName)
}

func (p commentRepo) CreateComment(ctx context.Context, comment *entity.Comment) (*entity.Comment, error) {

	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"CreateComment")
	defer span.End()

	data := map[string]any{
		"id":         comment.Id,
		"post_id":    comment.PostId,
		"owner_id":   comment.OwnerId,
		"message":    comment.Message,
		"likes":      comment.Likes,
		"dislikes":   comment.Dislikes,
		"created_at": comment.CreatedAt,
		"updated_at": comment.UpdatedAt,
	}
	query, args, err := p.db.Sq.Builder.Insert(p.tableName).SetMap(data).ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "create"))
	}

	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, p.db.Error(err)
	}

	return comment, nil
}

func (p commentRepo) UpdateComment(ctx context.Context, category *entity.CommentUpdateReq) (*entity.CommentUpdateReq, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"UpdateComment")
	defer span.End()

	clauses := map[string]any{
		"message":    category.Message,
		"updated_at": category.UpdatedAt,
	}
	sqlStr, args, err := p.db.Sq.Builder.
		Update(p.tableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("id", category.Id)).
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

	return category, nil
}

func (p commentRepo) DeleteComment(ctx context.Context, req *entity.DeleteReq) error {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"DeleteComment")
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

func (p commentRepo) GetComment(ctx context.Context, params map[string]string) (*entity.Comment, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"GetComment")
	defer span.End()

	var (
		comment entity.Comment
		cnt     int
	)

	queryBuilder := p.comentSelectQueryPrefix()

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
		&comment.Id,
		&comment.PostId,
		&comment.OwnerId,
		&comment.Message,
		&comment.Likes,
		&comment.Dislikes,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	); err != nil {
		return nil, p.db.Error(err)
	}

	return &comment, nil
}

func (p commentRepo) ListComment(ctx context.Context, limit int, offset int, params map[string]string) (*entity.CommentListRes, error) {

	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"ListComment")
	defer span.End()

	var (
		comments entity.CommentListRes
	)
	queryBuilder := p.comentSelectQueryPrefix()

	if limit != 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit)).Offset(uint64(offset))
	}

	for key, value := range params {
		if key == "owner_id" {
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal(key, value))
		}
		if key == "post_id" {
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

	for rows.Next() {
		var comment entity.Comment

		if err = rows.Scan(
			&comment.Id,
			&comment.PostId,
			&comment.OwnerId,
			&comment.Message,
			&comment.Likes,
			&comment.Dislikes,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, p.db.Error(err)
		}

		comments.Comment = append(comments.Comment, &comment)
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
		comments.TotalCount = 0
	}
	comments.TotalCount = int(count)

	return &comments, nil
}

func (p commentRepo) UpdateLike(ctx context.Context, req *entity.Like) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"UpdateLike")
	defer span.End()

	clauses := map[string]any{
		"status": req.Status,
	}
	sqlStr, args, err := p.db.Sq.Builder.
		Update(likeTableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("comment_id", req.CommentId)).
		Where(p.db.Sq.Equal("owner_id", req.OwnerId)).
		Where(p.db.Sq.Equal("post_id", req.PostId)).
		ToSql()

	if err != nil {
		return false, p.db.ErrSQLBuild(err, likeTableName+" update")
	}

	commandTag, err := p.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return false, p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return false, p.db.Error(fmt.Errorf("no sql rows"))
	}

	return true, nil
}

func (p commentRepo) GetLike(ctx context.Context, req *entity.Like) (*entity.Like, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"GerLike")
	defer span.End()

	var (
		like entity.Like
	)

	queryBuilder := p.db.Sq.Builder.
		Select(
			"comment_id",
			"post_id",
			"owner_id",
			"status",
		).From(likeTableName)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", likeTableName, "get"))
	}

	if err = p.db.QueryRow(ctx, query, args...).Scan(
		&like.CommentId,
		&like.PostId,
		&like.OwnerId,
		&like.Status,
	); err != nil {
		return nil, p.db.Error(err)
	}

	return &like, nil
}

func (p commentRepo) IsUnique(ctx context.Context, OwnerId, PostId, CommentId string) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"IsUnique")
	defer span.End()

	queryBuilder := p.db.Sq.Builder.Select("COUNT(1)").
		From(likeTableName).
		Where(squirrel.Eq{"comment_id": CommentId, "post_id": PostId, "owner_id": OwnerId})

	query, args, err := queryBuilder.ToSql()

	if err != nil {
		return false, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", likeTableName, "isUnique"))
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

func (p commentRepo) CreateLike(ctx context.Context, req *entity.Like) (bool, error) {
	data := map[string]any{
		"owner_id":   req.OwnerId,
		"comment_id": req.CommentId,
		"post_id":    req.PostId,
		"status":     req.Status,
	}
	query, args, err := p.db.Sq.Builder.Insert(likeTableName).SetMap(data).ToSql()

	if err != nil {
		return false, err
	}
	_, err = p.db.Exec(ctx, query, args...)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (p commentRepo) DeleteLike(ctx context.Context, req *entity.Like) error {

	sqlStr, args, err := p.db.Sq.Builder.
		Delete(likeTableName).
		Where(p.db.Sq.Equal("owner_id", req.OwnerId)).
		Where(p.db.Sq.Equal("post_id", req.PostId)).
		Where(p.db.Sq.Equal("comment_id", req.CommentId)).
		ToSql()

	if err != nil {
		return err
	}

	commandTag, err := p.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return p.db.Error(fmt.Errorf("no sql rows"))
	}

	return nil
}

func (p commentRepo) UpdateCommentLike(ctx context.Context, Id string, status bool) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"UpdateLike")
	defer span.End()

	var query string
	if status {
		query = `update comments set likes = likes + 1 where id = $1`
	} else {
		query = `update comments set likes = likes - 1 where id = $1`
	}

	commandTag, err := p.db.Exec(ctx, query, Id)
	if err != nil {
		return false, p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return false, p.db.Error(fmt.Errorf("no sql rows"))
	}

	return true, nil

}
func (p commentRepo) UpdateCommentDislike(ctx context.Context, Id string, status bool) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNameCommentService, spanNameCommentService+"UpdateLike")
	defer span.End()

	var query string
	if status {
		query = `update comments set dislikes = dislikes + 1 where id = $1`
	} else {
		query = `update comments set dislikes = dislikes - 1 where id = $1`
	}

	commandTag, err := p.db.Exec(ctx, query, Id)
	if err != nil {
		return false, p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return false, p.db.Error(fmt.Errorf("no sql rows"))
	}

	return true, nil

}
