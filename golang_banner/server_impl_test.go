package main

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
	"net/http"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v8"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func testUserBannerCacheGet(t *testing.T){
	db, db_mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()
	data := map[string]interface{}{
		"key" : "value",
	}
	jsonData, err := json.Marshal(data)
	if err != nil{
		t.Fatalf("failed to serialize JSON: %s", err)
	}

	rows := sqlmock.NewRows([]string{"id", "tag_ids", "feature_id", "content", "is_active", "created_at", "updated_at"}).
	AddRow(1, []int{1, 2, 3}, 2, jsonData, true, time.Now, time.Now)
	db_mock.ExpectQuery("select content,is_active from banners where feature_id = ($1) and ($2) = ANY(tag_ids)").WithArgs(2,3).
	WillReturnRows(rows)
	cache, cache_mock := redismock.NewClientMock()
	cache_mock.ExpectGet("2:3").SetVal(string(jsonData))
	
	server := &Server{
		tokens: map[string]string{
			"IGOTTHEPOWER!": "admin",
			"IMACREEP":      "user",
		},
		db:    db,
		cache: cache,
		ctx:   context.Background(),
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	params := GetUserBannerParams{
		TagId:          3,
		FeatureId:      2,
		UseLastRevision: new(bool), // Initialize a pointer to a bool
		Token:          new(string), // Initialize a pointer to a string
	}
	*params.UseLastRevision = true // Set the value of the bool pointer
	*params.Token = "IMACREEP"
	if assert.NoError(t, server.GetUserBanner(c, params)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
}