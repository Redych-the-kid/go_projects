package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v8"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserBannerCacheGet(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	data := map[string]interface{}{
		"key": "value",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to serialize JSON: %s", err)
	}
	
	cache, cache_mock := redismock.NewClientMock()
	cache_mock.ExpectGet("2:3").SetVal(string(jsonData))
	cache_mock.ExpectGet("2:3:isactive").SetVal("true")
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
		TagId:           3,
		FeatureId:       2,
		UseLastRevision: new(bool),   // Initialize a pointer to a bool
		Token:           new(string), // Initialize a pointer to a string
	}
	*params.UseLastRevision = false // Set the value of the bool pointer
	*params.Token = "IMACREEP"
	if assert.NoError(t, server.GetUserBanner(c, params)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		body, err := io.ReadAll(rec.Body)
		if err != nil{
			t.Fail()
		}
		var jsonBody map[string]interface{}
		err = json.Unmarshal(body, &jsonBody)
		if err != nil{
			t.Fatalf("Error occcured: %s", err.Error())
		}
		assert.Equal(t, jsonBody, data)
	} else {
		t.Fail()
	}
	if err := cache_mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUserBannerGetCacheForbidden(t *testing.T){
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	data := map[string]interface{}{
		"key": "value",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to serialize JSON: %s", err)
	}
	cache, cache_mock := redismock.NewClientMock()
	cache_mock.ExpectGet("2:3").SetVal(string(jsonData))
	cache_mock.ExpectGet("2:3:isactive").SetVal("false")
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
		TagId:           3,
		FeatureId:       2,
		UseLastRevision: new(bool),   // Initialize a pointer to a bool
		Token:           new(string), // Initialize a pointer to a string
	}
	*params.UseLastRevision = false // Set the value of the bool pointer
	*params.Token = "IMACREEP"
	if assert.NoError(t, server.GetUserBanner(c, params)){
		assert.Equal(t, http.StatusForbidden, rec.Code)
	}
}

func TestUserBannerGetDB(t *testing.T){
	db, db_mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	
	defer db.Close()
	data := map[string]interface{}{
		"key": "value",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to serialize JSON: %s", err)
	}
	cache, cache_mock := redismock.NewClientMock()
	cache_mock.ExpectSet("2:3", jsonData, 5 * time.Minute).SetVal("OK")
	cache_mock.ExpectSet("2:3:isactive", true, 5 * time.Minute).SetVal("OK")
	rows := sqlmock.NewRows([]string{"content", "is_active"}).
    AddRow(jsonData, true)
	db_mock.ExpectQuery("SELECT content, is_active FROM banners WHERE feature_id = ($1) AND ($2) = ANY(tag_ids)").
    WithArgs(2, 3).
    WillReturnRows(rows)

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
		TagId:           3,
		FeatureId:       2,
		UseLastRevision: new(bool),   // Initialize a pointer to a bool
		Token:           new(string), // Initialize a pointer to a string
	}
	*params.UseLastRevision = true // Set the value of the bool pointer
	*params.Token = "IMACREEP"
	if assert.NoError(t, server.GetUserBanner(c, params)){
		body, err := io.ReadAll(rec.Body)
		if err != nil{
			t.Fail()
		}
		var jsonBody map[string]interface{}
		err = json.Unmarshal(body, &jsonBody)
		if err != nil{
			t.Fatalf("Error occcured: %s", err.Error())
		}
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, jsonBody, data)
	}
	if err := db_mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	if err := cache_mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUserBannerGetDBForbidden(t *testing.T){
	db, db_mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	
	defer db.Close()
	data := map[string]interface{}{
		"key": "value",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to serialize JSON: %s", err)
	}
	cache, _ := redismock.NewClientMock()
	rows := sqlmock.NewRows([]string{"content", "is_active"}).
    AddRow(jsonData, false)
	db_mock.ExpectQuery("SELECT content, is_active FROM banners WHERE feature_id = ($1) AND ($2) = ANY(tag_ids)").
    WithArgs(2, 3).
    WillReturnRows(rows)
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
		TagId:           3,
		FeatureId:       2,
		UseLastRevision: new(bool),   // Initialize a pointer to a bool
		Token:           new(string), // Initialize a pointer to a string
	}
	*params.UseLastRevision = true // Set the value of the bool pointer
	*params.Token = "IMACREEP"
	if assert.NoError(t, server.GetUserBanner(c, params)){
		assert.Equal(t, http.StatusForbidden, rec.Code)
	}
}

func TestUserBannerGetDBNotFound(t *testing.T){
	db, db_mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	cache, _ := redismock.NewClientMock()
	db_mock.ExpectQuery("SELECT content, is_active FROM banners WHERE feature_id = ($1) AND ($2) = ANY(tag_ids)").
    WithArgs(2, 3).
    WillReturnError(errors.New("Not found!"))
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
		TagId:           3,
		FeatureId:       2,
		UseLastRevision: new(bool),   // Initialize a pointer to a bool
		Token:           new(string), // Initialize a pointer to a string
	}
	*params.UseLastRevision = true // Set the value of the bool pointer
	*params.Token = "IMACREEP"
	if assert.NoError(t, server.GetUserBanner(c, params)){
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestUserBannerGetDBUnauth(t *testing.T) {
	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	cache, _ := redismock.NewClientMock()
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
		TagId:           3,
		FeatureId:       2,
		UseLastRevision: new(bool),   // Initialize a pointer to a bool
		Token:           new(string), // Initialize a pointer to a string
	}
	*params.UseLastRevision = true // Set the value of the bool pointer
	*params.Token = "SCAMMER"
	if assert.NoError(t, server.GetUserBanner(c, params)){
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}