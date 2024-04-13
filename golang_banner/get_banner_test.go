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
	wrapper := ServerInterfaceWrapper{
		Handler: server,
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/user_banner?tag_id=3&feature_id=2&use_last_revision=false", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("token", "IMACREEP")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, wrapper.GetUserBanner(c)) {
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
	wrapper := ServerInterfaceWrapper{
		Handler: server,
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/user_banner?tag_id=3&feature_id=2&use_last_revision=false", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("token", "IMACREEP")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, wrapper.GetUserBanner(c)){
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
	wrapper := ServerInterfaceWrapper{
		Handler: server,
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/user_banner?tag_id=3&feature_id=2&use_last_revision=true", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("token", "IMACREEP")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, wrapper.GetUserBanner(c)){
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
	wrapper := ServerInterfaceWrapper{
		Handler: server,
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/user_banner?tag_id=3&feature_id=2&use_last_revision=true", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("token", "IMACREEP")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, wrapper.GetUserBanner(c)){
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
	wrapper := ServerInterfaceWrapper{
		Handler: server,
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/user_banner?tag_id=3&feature_id=2&use_last_revision=true", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("token", "IMACREEP")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, wrapper.GetUserBanner(c)){
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestUserBannerGetUnauth(t *testing.T) {
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
	wrapper := ServerInterfaceWrapper{
		Handler: server,
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/user_banner?tag_id=3&feature_id=2&use_last_revision=true", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("token", "SCAMMER")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, wrapper.GetUserBanner(c)){
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestUserBannerGetBadReq(t *testing.T){
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
	wrapper := ServerInterfaceWrapper{
		Handler: server,
	}
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/user_banner?feature_id=2", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err = wrapper.GetUserBanner(c)
	httpError := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusBadRequest, httpError.Code)
	assert.Equal(t, "code=400, message=Invalid format for parameter tag_id: query parameter 'tag_id' is required", err.Error())
	
	req = httptest.NewRequest(http.MethodGet, "/user_banner?tag_id=2", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	assert.EqualError(t, wrapper.GetUserBanner(c), "code=400, message=Invalid format for parameter feature_id: query parameter 'feature_id' is required")

	req = httptest.NewRequest(http.MethodGet, "/user_banner?tag_id=3&feature_id=2", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = wrapper.GetUserBanner(c)
	httpError = err.(*echo.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, httpError.Code)
	assert.Equal(t, "code=401, message=No token was provided", err.Error())
}