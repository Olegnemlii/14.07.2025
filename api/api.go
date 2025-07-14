package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/Olegnemlii/14.07.2025/config" // Импорт пакета config
	"github.com/Olegnemlii/14.07.2025/task"   // Импорт пакета task
)

type APIServer struct {
	listenAddr string
	config     *Config
	taskQueue  chan *Task
	tasks      map[string]*Task
	mu         sync.Mutex // Protects access to tasks map
}

func NewAPIServer(listenAddr string, config *Config) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		config:     config,
		taskQueue:  make(chan *Task, config.MaxTasks), // Buffered channel
		tasks:      make(map[string]*Task),
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/task", s.handleCreateTask).Methods("POST")
	router.HandleFunc("/task/{id}", s.handleGetTask).Methods("GET")
	router.HandleFunc("/task/{id}/url", s.handleAddURL).Methods("POST")
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(".")))) // Serve files from current directory (for zip downloads)

	log.Println("Server running on port:", s.listenAddr)
	log.Fatal(http.ListenAndServe(s.listenAddr, router))
}

func (s *APIServer) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if len(s.tasks) >= s.config.MaxTasks {
		s.mu.Unlock()
		respondWithError(w, http.StatusServiceUnavailable, "Server is busy, maximum number of tasks reached")
		return
	}
	s.mu.Unlock()

	taskID := uuid.New().String()
	task := NewTask(taskID, s.config)

	s.mu.Lock()
	s.tasks[taskID] = task
	s.mu.Unlock()

	s.taskQueue <- task // Add task to the queue

	respondWithJSON(w, http.StatusCreated, map[string]string{"task_id": taskID})
	log.Printf("Task %s created", taskID)
}

func (s *APIServer) handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	s.mu.Lock()
	task, ok := s.tasks[taskID]
	s.mu.Unlock()

	if !ok {
		respondWithError(w, http.StatusNotFound, "Task not found")
		return
	}

	status, resultURL, errors := task.GetStatus()

	response := map[string]interface{}{
		"status":  status,
		"errors":  errors,
		"task_id": taskID,
	}

	if status == StatusCompleted {
		response["result_url"] = resultURL
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleAddURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	s.mu.Lock()
	task, ok := s.tasks[taskID]
	s.mu.Unlock()

	if !ok {
		respondWithError(w, http.StatusNotFound, "Task not found")
		return
	}

	type addURLRequest struct {
		URL string `json:"url"`
	}

	var req addURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	err := task.AddURL(req.URL)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "URL added to task"})
	log.Printf("URL added to Task %s: %s", taskID, req.URL)

	// If task has enough URLs, start it
	s.mu.Lock()
	if len(task.URLs) == s.config.MaxFilesPerTask && task.Status == StatusPending {
		s.mu.Unlock()
		go task.Run() // Start task in a goroutine
		log.Printf("Task %s started", taskID)
	} else {
		s.mu.Unlock()
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
