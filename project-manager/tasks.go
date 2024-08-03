package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var errNameRequired = errors.New("task name is required")
var errProjectIDRequired = errors.New("project ID is required")
var errUSerIDRequired = errors.New("user ID is required")

type TasksService struct {
	store Store
}

func NewTasksService(s Store) *TasksService {
	return &TasksService{store: s}
}

func (s *TasksService) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("POST /tasks", s.handleCreateTask)
	r.HandleFunc("GET /tasks/{id}", s.handleGetTask)
}

func (s *TasksService) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, ErrorResponse{
			Error: "Error reading request body: " + err.Error(),
		})
		return
	}

	defer r.Body.Close()

	var task *Task
	err = json.Unmarshal(body, &task)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, ErrorResponse{
			Error: "Invalid JSON payload: " + err.Error(),
		})
		return
	}

	if err := task.validate(); err != nil {
		WriteJson(w, http.StatusBadRequest, ErrorResponse{
			Error: "Invalid task payload: " + err.Error(),
		})
		return
	}

	t, err := s.store.CreateTask(task)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, ErrorResponse{
			Error: "Error creating task: " + err.Error(),
		})
		return
	}

	WriteJson(w, http.StatusCreated, t)
}

func (s *TasksService) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteJson(w, http.StatusBadRequest, ErrorResponse{
			Error: "Task ID is required",
		})
		return
	}

	t, err := s.store.GetTask(id)
	if err != nil {
		WriteJson(w, http.StatusNotFound, ErrorResponse{
			Error: "Error getting task: " + err.Error(),
		})
		return
	}

	WriteJson(w, http.StatusOK, t)
}

func (t *Task) validate() error {
	if t.Name == "" {
		return errNameRequired
	}

	if t.ProjectID == 0 {
		return errProjectIDRequired
	}

	if t.AssignedToID == 0 {
		return errUSerIDRequired
	}

	return nil
}
