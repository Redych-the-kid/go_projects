package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

type Server struct {
	tokens map[string]string
	db *sql.DB
}

type Banner struct {
    ID          int64          `json:"id"`
    TagIDs      []int64         `json:"tag_ids"`
    FeatureID   int            `json:"feature_id"`
    Content     json.RawMessage `json:"content"`
    IsActive    bool           `json:"is_active"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

func (s *Server) GetBanner(ctx echo.Context, params GetBannerParams) error {
	if params.Token == nil || !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(401, "Пользователь не авторизован")
	}
	if !validateAdminToken(*params.Token, s.tokens){
		return ctx.HTML(403, "Пользователь не имеет доступа")
	}
	query := getBannerQueryBuilder(params)
	rows, err := s.db.Query(query)
	println(query)
	if err != nil {
		print("Transaction failure")
		return ctx.HTML(500, "Внутренняя ошибка сервера")
	}
	defer rows.Close()
	var banners []Banner
	for rows.Next(){
		var banner Banner
		err := rows.Scan(&banner.ID, pq.Array(&banner.TagIDs), &banner.FeatureID, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt)
        if err != nil {
            return ctx.JSON(http.StatusInternalServerError, err.Error())
        }
        banners = append(banners, banner)
	}
	return ctx.JSON(200, banners)
}

func (s *Server) PostBanner(ctx echo.Context, params PostBannerParams) error {
	if params.Token == nil || !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(401, "Пользователь не авторизован")
	}
	if !validateAdminToken(*params.Token, s.tokens){
		return ctx.HTML(403, "Пользователь не имеет доступа")
	}
	var data map[string]interface{}
	if err := ctx.Bind(&data); err != nil {
		return ctx.HTML(400, "Некорректные данные")
	}
	var content map[string]interface{}
	var feature_id int
	var tag_ids []int
	var is_active bool
	if !jsonToParams(data, &content , &feature_id, &tag_ids, &is_active){
		fmt.Println(content, feature_id, tag_ids, is_active)
		return ctx.HTML(400, "Некорректные данные")
	}
	contentJSON, err := json.Marshal(content)
	if err != nil{
		return ctx.HTML(400, "Некорректные данные")
	}
	var id int
	query := `INSERT INTO banners (content, feature_id, tag_ids, is_active) VALUES ($1, $2, $3, $4) RETURNING id`
	err = s.db.QueryRow(query, contentJSON, feature_id, pq.Array(tag_ids), is_active).Scan(&id)
	if err != nil{
		return ctx.HTML(500, "Внутренняя ошибка сервера")
	}
	return ctx.JSON(201, id)
}

func (s *Server) DeleteBannerId(ctx echo.Context, id int, params DeleteBannerIdParams) error {
	if params.Token == nil || !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(401, "Пользователь не авторизован")
	}
	if !validateAdminToken(*params.Token, s.tokens){
		return ctx.HTML(403, "Пользователь не имеет доступа")
	}
	var scanID int
	query := "DELETE FROM banners WHERE id = $1 RETURNING  id"
	err := s.db.QueryRow(query, id).Scan(&scanID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.HTML(404, "Баннер для тэга не найден")
		} else {
			return ctx.HTML(500, "Внутренняя ошибка сервера")
		}
	}
	return ctx.HTML(204, "Баннер успешно удалён")
}

func (s *Server) PatchBannerId(ctx echo.Context, id int, params PatchBannerIdParams) error {
	if params.Token == nil || !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(401, "Пользователь не авторизован")
	}
	if !validateAdminToken(*params.Token, s.tokens){
		return ctx.HTML(403, "Пользователь не имеет доступа")
	}
	var data map[string]interface{} = make(map[string]interface{})
	if err := ctx.Bind(&data); err != nil {
		return ctx.HTML(400, "Некорректные данные")
	}
	var content map[string]interface{}
	var feature_id int
	var tag_ids []int
	var is_active bool
	if !jsonToParams(data, &content , &feature_id, &tag_ids, &is_active){
		fmt.Println(content, feature_id, tag_ids, is_active)
		return ctx.HTML(400, "Некорректные данные")
	}
	contentJSON, err := json.Marshal(content)
	if err != nil{
		return ctx.HTML(400, "Некорректные данные")
	}
	query := ` UPDATE banners
	SET content = $1, feature_id = $2, tag_ids = $3, is_active = $4
	WHERE id = $5;`
	res, err := s.db.Exec(query, contentJSON, feature_id, pq.Array(tag_ids), is_active, id)
	if err != nil {
		return ctx.HTML(500, "Внутренняя ошибка сервера")
	}

	count, err := res.RowsAffected()
	if err != nil {
		return ctx.HTML(500, "Внутренняя ошибка сервера")
	}
	if count == 0{
		return ctx.HTML(404, "Баннер не найден")
	}
	return ctx.HTML(200, "OK")
}

func (s *Server) GetUserBanner(ctx echo.Context, params GetUserBannerParams) error {
	if params.Token == nil || !validateToken(*params.Token, s.tokens) {
		return ctx.HTML(401, "Пользователь не авторизован")
	}
	query := "select content,is_active from banners where feature_id = ($1) and ($2) = ANY(tag_ids)"
	var jsonData []byte
	var is_active bool
	err := s.db.QueryRow(query, params.FeatureId, params.TagId).Scan(&jsonData, &is_active)
	if err != nil{
		return ctx.HTML(404, "Баннер не найден")
	}
	if(!is_active && !validateAdminToken(*params.Token, s.tokens)){
		return ctx.HTML(403, "Пользователь не имеет доступа")
	}
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return ctx.HTML(400, "Некорректные данные")
	}
	return ctx.JSON(200, result)
}

// Validation functions for tokens
func validateToken(token string, tokens map[string]string) bool {
	if(token != ""){
		return tokens[token] != ""
	} else{
		return false
	}
}

func validateAdminToken(token string, tokens map[string]string) bool {
	return tokens[token] == "admin"
}

func jsonToParams(data map[string]interface{}, content *map[string]interface{}, featureID *int, tagIDs *[]int, isActive *bool) bool{
	var ok bool
	*content, ok = data["content"].(map[string]interface{})
	if !ok{
		return false
	}
	float_val, ok := data["feature_id"].(float64)
	if !ok{
		return false
	}
	*featureID = int(float_val)
	ifaceSlice, ok := data["tag_ids"].([]interface{})
		if ok {
			*tagIDs = make([]int, len(ifaceSlice))
			for i, v := range ifaceSlice {
				(*tagIDs)[i] = int(v.(float64))
			}
		} else{
			return false
		}
	*isActive, ok = data["is_active"].(bool)
	return ok
}

func getBannerQueryBuilder(params GetBannerParams) string{
	query := "SELECT * FROM banners"
	if params.FeatureId != nil || params.TagId != nil{
		query += " WHERE"
		before := false
		if(params.FeatureId != nil){
			query += fmt.Sprintf(" feature_id = %d", *params.FeatureId)
			before = true
		}
		if(params.TagId != nil){
			if(before){
				query += " AND"
			}
			query += fmt.Sprintf(" %d = ANY(tag_ids)", *params.TagId)
		}
	}
	if params.Limit != nil {
		query += fmt.Sprintf(" LIMIT %d", *params.Limit)
	}
	if params.Offset != nil {
		query += fmt.Sprintf(" OFFSET %d", *params.Offset)
	}

	return query
}