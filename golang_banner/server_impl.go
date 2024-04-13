package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

type Server struct {
	tokens map[string]string
	db     *sql.DB
	cache *redis.Client
	ctx context.Context
}

type Banner struct {
	ID        int64           `json:"id"`
	TagIDs    []int64         `json:"tag_ids"`
	FeatureID int             `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (s *Server) GetBanner(ctx echo.Context, params GetBannerParams) error {
	if !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusUnauthorized, "Пользователь не авторизован")
	}
	if !validateAdminToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusForbidden, "Пользователь не имеет доступа")
	}
	query, args := getBannerQueryBuilder(params)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	var banners []Banner
	for rows.Next() {
		var banner Banner
		err := rows.Scan(&banner.ID, pq.Array(&banner.TagIDs), &banner.FeatureID, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		banners = append(banners, banner)
	}
	return ctx.JSON(http.StatusOK, banners)
}

func (s *Server) PostBanner(ctx echo.Context, params PostBannerParams) error {
	if !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusUnauthorized, "Пользователь не авторизован")
	}
	if !validateAdminToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusForbidden, "Пользователь не имеет доступа")
	}
	var data map[string]interface{}
	if err := ctx.Bind(&data); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	var content map[string]interface{}
	var feature_id int
	var tag_ids []int
	var is_active bool
	err := jsonToParams(data, &content, &feature_id, &tag_ids, &is_active)
	if err != nil{
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	var id int
	query := `INSERT INTO banners (content, feature_id, tag_ids, is_active) VALUES ($1, $2, $3, $4) RETURNING id`
	err = s.db.QueryRow(query, contentJSON, feature_id, pq.Array(tag_ids), is_active).Scan(&id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusCreated, id)
}

func (s *Server) DeleteBannerId(ctx echo.Context, id int, params DeleteBannerIdParams) error {
	if !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusUnauthorized, "Пользователь не авторизован")
	}
	if !validateAdminToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusForbidden, "Пользователь не имеет доступа")
	}
	var scanID int
	query := "DELETE FROM banners WHERE id = $1 RETURNING  id"
	err := s.db.QueryRow(query, id).Scan(&scanID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.HTML(http.StatusNotFound, "Баннер не найден")
		} else {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
	}
	return ctx.HTML(http.StatusNoContent, "Баннер успешно удалён")
}

func (s *Server) PatchBannerId(ctx echo.Context, id int, params PatchBannerIdParams) error {
	if !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusUnauthorized, "Пользователь не авторизован")
	}
	if !validateAdminToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusForbidden, "Пользователь не имеет доступа")
	}
	var data map[string]interface{} = make(map[string]interface{})
	if err := ctx.Bind(&data); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	var content map[string]interface{}
	var feature_id int
	var tag_ids []int
	var is_active bool
	err := jsonToParams(data, &content, &feature_id, &tag_ids, &is_active)
	if err != nil{
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	query := ` UPDATE banners
	SET content = $1, feature_id = $2, tag_ids = $3, is_active = $4
	WHERE id = $5;`
	res, err := s.db.Exec(query, contentJSON, feature_id, pq.Array(tag_ids), is_active, id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	count, err := res.RowsAffected()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	if count == 0 {
		return ctx.HTML(http.StatusNotFound, "Баннер не найден")
	}
	return ctx.HTML(http.StatusOK, "OK")
}

func (s *Server) GetUserBanner(ctx echo.Context, params GetUserBannerParams) error {
	if !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusUnauthorized, "Пользователь не авторизован")
	}
	var result map[string]interface{}
	var is_active bool
	if(params.UseLastRevision == nil || !*params.UseLastRevision){
		value, err := s.cache.Get(s.ctx, fmt.Sprintf("%d:%d", params.FeatureId, params.TagId)).Result()
		if err == nil {
			err := json.Unmarshal([]byte(value), &result)
			if err != nil{
				return ctx.JSON(http.StatusInternalServerError, err.Error())
			}
			activeVal, err := s.cache.Get(s.ctx, fmt.Sprintf("%d:%d:isactive", params.FeatureId, params.TagId)).Result()
			if err != nil{
				return ctx.JSON(http.StatusInternalServerError, err.Error())
			}
			is_active, err = strconv.ParseBool(activeVal)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err.Error())
			}
			if !is_active && !validateAdminToken(*params.Token, s.tokens) {
				return ctx.HTML(http.StatusForbidden, "Пользователь не имеет доступа")
			}
			return ctx.JSON(http.StatusOK, result)
		}
	}
	query := "SELECT content, is_active FROM banners WHERE feature_id = ($1) AND ($2) = ANY(tag_ids)"
	var jsonData []byte
	err := s.db.QueryRow(query, params.FeatureId, params.TagId).Scan(&jsonData, &is_active)
	if err != nil {
		return ctx.HTML(http.StatusNotFound, "Баннер не найден")
	}
	if !is_active && !validateAdminToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusForbidden, "Пользователь не имеет доступа")
	}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	err = s.cache.Set(s.ctx, fmt.Sprintf("%d:%d", params.FeatureId, params.TagId), jsonData, 5 * time.Minute).Err()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	err = s.cache.Set(s.ctx, fmt.Sprintf("%d:%d:isactive", params.FeatureId, params.TagId), is_active, 5 * time.Minute).Err()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, result)
}
