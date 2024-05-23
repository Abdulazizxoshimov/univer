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
	categoryServiceTableName   = "category"
	serviceNameCategoryService = "categoryServiceRepo"
	spanNameCategoryService    = "categorySpanRepo"
)

type categoryRepo struct {
	tableName string
	db        *postgres.PostgresDB
}

func NewCategoryRepo(db *postgres.PostgresDB) *categoryRepo {
	return &categoryRepo{
		tableName: categoryServiceTableName,
		db:        db,
	}
}

func (p *categoryRepo) categorySelectQueryPrefix() squirrel.SelectBuilder {
	return p.db.Sq.Builder.
		Select(
			"id",
			"name",
			"created_at",
			"updated_at",
		).From(p.tableName)
}

func (p categoryRepo) CreateCategory(ctx context.Context, category *entity.Category) (*entity.Category, error) {

	ctx, span := otlp.Start(ctx, serviceNameCategoryService, spanNameCategoryService+"CreateCategory")
	defer span.End()

	data := map[string]any{
		"id":         category.Id,
		"name":       category.Name,
		"created_at": category.CreatedAt,
		"updated_at": category.UpdatedAt,
	}
	query, args, err := p.db.Sq.Builder.Insert(p.tableName).SetMap(data).ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "create"))
	}

	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, p.db.Error(err)
	}

	return category, nil
}

func (p categoryRepo) UpdateCategory(ctx context.Context, category *entity.UpdateCategory) (*entity.UpdateCategory, error) {

	ctx, span := otlp.Start(ctx, serviceNameCategoryService, spanNameCategoryService+"UpdateCategory")
	defer span.End()

	clauses := map[string]any{
		"name":       category.Name,
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

func (p categoryRepo) DeleteCategory(ctx context.Context, req *entity.DeleteReq) error {
	ctx, span := otlp.Start(ctx, serviceNameCategoryService, spanNameCategoryService+"DeleteCategory")
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

func (p categoryRepo)GetCategory(ctx context.Context, params map[string]string) (*entity.Category, error) {
	ctx, span := otlp.Start(ctx, serviceNameCategoryService, spanNameCategoryService+"GetCategory")
	defer span.End()

	var (
		category entity.Category
		cnt  int
	)

	queryBuilder := p.categorySelectQueryPrefix()

	for key, value := range params {
		if key == "id" {
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal(key, value))
		} else if key == "name" {
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
		&category.Id,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
	); err != nil {
		return nil, p.db.Error(err)
	}
	

	return &category, nil
}

func (p categoryRepo) ListCategory(ctx context.Context, limit int, offset int) (*entity.ListCategoryRes, error) {

	ctx, span := otlp.Start(ctx, serviceNameCategoryService, spanNameCategoryService+"ListCategory")
	defer span.End()

	var (
		categories entity.ListCategoryRes
	)
	queryBuilder := p.categorySelectQueryPrefix()

	if limit != 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit)).Offset(uint64(offset))
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
		var category  entity.Category
		
		if err = rows.Scan(
			&category.Id,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, p.db.Error(err)
		}
		

		categories.Category = append(categories.Category, &category)
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
		categories.Totalcount = 0
	}
	categories.Totalcount = int(count)

	return &categories, nil
}
