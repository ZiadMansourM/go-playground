package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTask(t *testing.T) {
	t.Run("Name is required", func(t *testing.T) {
		payload := &Task{
			Name: "",
		}
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()
		router := http.NewServeMux()

		ms := &MockStore{}
		service := NewTasksService(ms)
		service.RegisterRoutes(router)

		router.HandleFunc("/tasks", service.handleCreateTask)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("Task creation success", func(t *testing.T) {
		payload := &Task{
			Name:         "Test Task",
			ProjectID:    1,
			AssignedToID: 42,
		}

		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()
		router := http.NewServeMux()

		ms := &MockStore{}
		service := NewTasksService(ms)
		service.RegisterRoutes(router)

		router.HandleFunc("/tasks", service.handleCreateTask)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, rec.Code)
		}
	})
}

func TestGetTask(t *testing.T) {
	ms := &MockStore{}
	service := NewTasksService(ms)

	t.Run("Return task", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/tasks/42", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()
		router := http.NewServeMux()

		router.HandleFunc("GET /tasks/{id}", service.handleGetTask)

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
