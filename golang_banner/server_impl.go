package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

type Server struct {
	tokens map[string]string
	db     *sql.DB
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
	query := getBannerQueryBuilder(params)
	rows, err := s.db.Query(query)
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
		return ctx.HTML(http.StatusBadRequest, err.Error())
	}
	var content map[string]interface{}
	var feature_id int
	var tag_ids []int
	var is_active bool
	if !jsonToParams(data, &content, &feature_id, &tag_ids, &is_active) {
		fmt.Println(content, feature_id, tag_ids, is_active)
		return ctx.HTML(http.StatusBadRequest, "Некорректные данные")
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return ctx.HTML(http.StatusBadRequest, err.Error())
	}
	var id int
	query := `INSERT INTO banners (content, feature_id, tag_ids, is_active) VALUES ($1, $2, $3, $4) RETURNING id`
	err = s.db.QueryRow(query, contentJSON, feature_id, pq.Array(tag_ids), is_active).Scan(&id)
	if err != nil {
		return ctx.HTML(http.StatusInternalServerError, err.Error())
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
			return ctx.HTML(http.StatusNotFound, err.Error())
		} else {
			return ctx.HTML(http.StatusInternalServerError, err.Error())
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
		return ctx.HTML(http.StatusBadRequest, err.Error())
	}
	var content map[string]interface{}
	var feature_id int
	var tag_ids []int
	var is_active bool
	if !jsonToParams(data, &content, &feature_id, &tag_ids, &is_active) {
		fmt.Println(content, feature_id, tag_ids, is_active)
		return ctx.HTML(http.StatusBadRequest, "Некорректные данные")
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return ctx.HTML(http.StatusBadRequest, "Некорректные данные")
	}
	query := ` UPDATE banners
	SET content = $1, feature_id = $2, tag_ids = $3, is_active = $4
	WHERE id = $5;`
	res, err := s.db.Exec(query, contentJSON, feature_id, pq.Array(tag_ids), is_active, id)
	if err != nil {
		return ctx.HTML(http.StatusInternalServerError, err.Error())
	}

	count, err := res.RowsAffected()
	if err != nil {
		return ctx.HTML(http.StatusInternalServerError, err.Error())
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
	query := "select content,is_active from banners where feature_id = ($1) and ($2) = ANY(tag_ids)"
	var jsonData []byte
	var is_active bool
	err := s.db.QueryRow(query, params.FeatureId, params.TagId).Scan(&jsonData, &is_active)
	if err != nil {
		return ctx.HTML(http.StatusNotFound, "Баннер не найден")
	}
	if !is_active && !validateAdminToken(*params.Token, s.tokens) {
		return ctx.HTML(http.StatusForbidden, "Пользователь не имеет доступа")
	}
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return ctx.HTML(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, result)
}
