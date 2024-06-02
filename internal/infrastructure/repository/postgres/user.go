package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"univer/internal/entity"
	"univer/internal/pkg/otlp"
	postgres "univer/internal/pkg/storage"

	"github.com/Masterminds/squirrel"
)

const (
	userServiceTableName   = "users"
	serviceNameUserService = "userServiceRepo"
	spanNameUserService    = "userSpanRepo"
)

type userRepo struct {
	tableName string
	db        *postgres.PostgresDB
}

func NewUserRepo(db *postgres.PostgresDB) *userRepo {
	return &userRepo{
		tableName: userServiceTableName,
		db:        db,
	}
}

func (p *userRepo) usersSelectQueryPrefix() squirrel.SelectBuilder {
	return p.db.Sq.Builder.
		Select(
			"id",
			"username",
			"email",
			"phone_number",
			"password",
			"bio",
			"image_url",
			"role",
			"refresh_token",
			"created_at",
			"updated_at",
		).From(p.tableName)
}

func (p userRepo) Create(ctx context.Context, user *entity.User) (*entity.User, error) {

	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService+"CreateUser")
	defer span.End()

	data := map[string]any{
		"id":            user.Id,
		"username":      user.UserName,
		"email":         user.Email,
		"phone_number":  user.PhoneNumber,
		"password":      user.Password,
		"role":          user.Role,
		"bio":           user.Bio,
		"image_url":     user.ImageUrl,
		"refresh_token": user.RefreshToken,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
	}
	query, args, err := p.db.Sq.Builder.Insert(p.tableName).SetMap(data).ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "create"))
	}
	
	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, p.db.Error(err)
	}

	return user, nil
}

func (p userRepo) Update(ctx context.Context, user *entity.User) (*entity.User, error) {

	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService+"UpdateUser")
	defer span.End()

	clauses := map[string]any{
		"username":     user.UserName,
		"email":        user.Email,
		"phone_number": user.PhoneNumber,
		"bio":          user.Bio,
		"image_url":    user.ImageUrl,
		"updated_at":   user.UpdatedAt,
	}
	sqlStr, args, err := p.db.Sq.Builder.
		Update(p.tableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("id", user.Id)).
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

	return user, nil
}

func (p userRepo) Delete(ctx context.Context, req *entity.DeleteReq) error {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService+"DeleteUser")
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

func (p userRepo) Get(ctx context.Context, params map[string]string) (*entity.User, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService+"GetUser")
	defer span.End()

	var (
		user entity.User
		cnt int
	)

	queryBuilder := p.usersSelectQueryPrefix()
    
	for key, value := range params {
		if key == "id" {
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal(key, value))
		} else if key == "email" {
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal(key, value))
		} else if key == "refresh_token" {
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal(key, value))
		}else if key == "username"{
			queryBuilder = queryBuilder.Where(p.db.Sq.Equal(key, value))
		}else if key == "del"{
			cnt ++
		}
	}
	if cnt == 0{
		queryBuilder = queryBuilder.Where("deleted_at is null")
	}
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "get"))
	}
	var (
		nullPhoneNumber sql.NullString
		nullBio         sql.NullString
		nullImageUrl    sql.NullString
		nullRefresh     sql.NullString
	)
	
	if err = p.db.QueryRow(ctx, query, args...).Scan(
		&user.Id,
		&user.UserName,
		&user.Email,
		&nullPhoneNumber,
		&user.Password,
		&nullBio,
		&nullImageUrl,
		&user.Role,
		&nullRefresh,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, p.db.Error(err)
	}
	if nullPhoneNumber.Valid {
		user.PhoneNumber = nullPhoneNumber.String
	}
	if nullBio.Valid {
		user.Bio = nullBio.String
	}
	if nullImageUrl.Valid {
		user.ImageUrl = nullImageUrl.String
	}
	if nullRefresh.Valid {
		user.RefreshToken = nullRefresh.String
	}

	return &user, nil
}

func (p userRepo) List(ctx context.Context, limit int, offset int, filter map[string]string) (*entity.ListUserRes, error) {

	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService+"ListUsers")
	defer span.End()

	var (
		users entity.ListUserRes
	)
	queryBuilder := p.usersSelectQueryPrefix()

	if limit != 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit)).Offset(uint64(offset))
	}

	role := filter["role"]
	queryBuilder = queryBuilder.Where(p.db.Sq.Equal("role", role))
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
		var (
			user            entity.User
			nullPhoneNumber sql.NullString
			nullBio         sql.NullString
			nullImageUrl    sql.NullString
			nullRefresh     sql.NullString
		)
		if err = rows.Scan(
			&user.Id,
			&user.UserName,
			&user.Email,
			&nullPhoneNumber,
			&user.Password,
			&nullBio,
			&nullImageUrl,
			&user.Role,
			&nullRefresh,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, p.db.Error(err)
		}
		if nullPhoneNumber.Valid {
			user.PhoneNumber = nullPhoneNumber.String
		}
		if nullBio.Valid {
			user.Bio = nullBio.String
		}
		if nullImageUrl.Valid {
			user.ImageUrl =  nullImageUrl.String
		}
		if nullRefresh.Valid {
			user.RefreshToken = nullRefresh.String
		}

		users.User = append(users.User, &user)
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
		users.TotalCount = 0
	}
	users.TotalCount = int64(count)

	return &users, nil
}

func (p userRepo) CheckUnique(ctx context.Context, filter *entity.GetReq) (bool, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "UniqueEmail")
	defer span.End()

	queryBuilder := p.db.Sq.Builder.Select("COUNT(1)").
	From(p.tableName)

	for key, value := range filter.Filter{
		if key == "email"{
			queryBuilder = queryBuilder.Where(squirrel.Eq{key: value})
		}
		if key == "username"{
			queryBuilder = queryBuilder.Where(squirrel.Eq{key: value})
		}
		if key == "phone_number"{
			queryBuilder = queryBuilder.Where(squirrel.Eq{key: value})
		}
	}

	queryBuilder = queryBuilder.Where("deleted_at IS NULL")

	
	query, args, err := queryBuilder.ToSql()

	if err != nil {
		return false, p.db.ErrSQLBuild(err, fmt.Sprintf("%s %s", p.tableName, "isUnique"))
	}
	
	var count int
	err = p.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return true, p.db.Error(err)
	}
	if count != 0 {
		return true, nil
	}

	return false, nil
}

func (p userRepo) UpdateRefresh(ctx context.Context, request *entity.UpdateRefresh) (*entity.Response, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "UpdateRefresh")
	defer span.End()

	clauses := map[string]any{
		"refresh_token": request.RefreshToken,
	}
	sqlStr, args, err := p.db.Sq.Builder.
		Update(p.tableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("id", request.UserID)).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return &entity.Response{Status: false}, p.db.ErrSQLBuild(err, p.tableName+" update")
	}

	commandTag, err := p.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return &entity.Response{Status: false}, p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return &entity.Response{Status: false}, p.db.Error(fmt.Errorf("no sql rows"))
	}

	return &entity.Response{Status: true}, nil
}

func (p userRepo) UpdatePassword(ctx context.Context, request *entity.UpdatePassword) (*entity.Response, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "UpdatePassword")
	defer span.End()

	clauses := map[string]any{
		"password": request.NewPassword,
	}
	sqlStr, args, err := p.db.Sq.Builder.
		Update(p.tableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("id", request.UserID)).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return &entity.Response{Status: false}, p.db.ErrSQLBuild(err, p.tableName+" update")
	}

	commandTag, err := p.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return &entity.Response{Status: false}, p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return &entity.Response{Status: false}, p.db.Error(fmt.Errorf("no sql rows"))
	}

	return &entity.Response{Status: true}, nil
}

func (p userRepo) UpdateProfile(ctx context.Context, request *entity.UpdateProfile) (*entity.Response, error) {
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService + "UpdatePassword")
	defer span.End()

	clauses := map[string]any{
		"image_url": request.ImageUrl,
	}
	sqlStr, args, err := p.db.Sq.Builder.
		Update(p.tableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("id", request.Id)).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return &entity.Response{Status: false}, p.db.ErrSQLBuild(err, p.tableName+" update")
	}

	commandTag, err := p.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return &entity.Response{Status: false}, p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return &entity.Response{Status: false}, p.db.Error(fmt.Errorf("no sql rows"))
	}

	return &entity.Response{Status: true}, nil
}

func (p userRepo) DeleteProfile(ctx context.Context, id string)error{
	ctx, span := otlp.Start(ctx, serviceNameUserService, spanNameUserService+"DeleteUser")
	defer span.End()
	clauses := map[string]interface{}{
		"image_url": "",
	}

	sqlStr, args, err := p.db.Sq.Builder.
		Update(p.tableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("id", id)).
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

func (p userRepo) UpdateToPremium(ctx context.Context, id string) (*entity.Response, error) {
	clauses := map[string]any{
		"role": "prouser",
	}
	sqlStr, args, err := p.db.Sq.Builder.
		Update(p.tableName).
		SetMap(clauses).
		Where(p.db.Sq.Equal("id", id)).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return &entity.Response{Status: false}, p.db.ErrSQLBuild(err, p.tableName+" update")
	}

	commandTag, err := p.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return &entity.Response{Status: false}, p.db.Error(err)
	}

	if commandTag.RowsAffected() == 0 {
		return &entity.Response{Status: false}, p.db.Error(fmt.Errorf("no sql rows"))
	}

	return &entity.Response{Status: true}, nil
}