package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testGetAllUser(t *testing.T) {
	req, err := http.NewRequest("GET", "/get", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(GetAllUser)
	handler.ServerHTTP(w, req)
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)

	}
	expected := `[{"id":1,"name":"test","balance":600000,"created_time":"12:00:00","modified_time":"12:00:00"},{"id":2,"name":"test2","balance":700000,"created_time":"12:00:00","modified_time":"12:00:00"},{"id":3,"name":"test3","balance":800000,"created_time":"12:00:00","modified_time":"12:00:00"}]`
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected bot: got %v want %v", w.Body.String(), expected)
	}
}
func testGetUserById(t *testing.T) {
	req, err := http.NewRequest("GET", "/get/{id}", nil)
	if err != nil {
		t.Fatal(err)
	}
	q := req.URL.Query()
	q.Add("id", "1")
	req.URL.RawQuery = q.Encode()
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(GetUserByID)
	handler.ServerHTTP(w, req)
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)

	}
	expected := `[{"id":1,"name":"test","balance":600000,"created_time":"12:00:00","modified_time":"12:00:00"}]`
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected bot: got %v want %v", w.Body.String(), expected)
	}
}

func testUpdateUserById(t *testing.T) {
	var jsonStr = []byte(`{"id":1,"name":"test","balance":600000,"created_time":"12:00:00","modified_time":"12:00:00"}`)
	req, err := http.NewRequest("PUT", "/update/{id}", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(GetUserByID)
	handler.ServerHTTP(w, req)
	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)

	}
	expected := `[{"id":1,"name":"test","balance":600000,"created_time":"12:00:00","modified_time":"12:00:00"}]`
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected bot: got %v want %v", w.Body.String(), expected)
	}
}
