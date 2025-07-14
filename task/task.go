package task

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"github.com/Olegnemlii/14.07.2025/config"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID           string
	URLs         []string
	Status       TaskStatus
	ResultURL    string
	Errors       []string
	Config       *Config // Add config to Task
	mu           sync.Mutex
	ctx          context.Context // Добавлено поле context
	done         chan struct{}   // Channel to signal completion
	cancel       context.CancelFunc
	creationTime time.Time
}

func NewTask(id string, config *Config) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	return &Task{
		ID:           id,
		URLs:         make([]string, 0),
		Status:       StatusPending,
		Errors:       make([]string, 0),
		Config:       config,
		ctx:          ctx, // Инициализация context
		done:         make(chan struct{}),
		cancel:       cancel,
		creationTime: time.Now(),
	}
}

func (t *Task) AddURL(url string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.URLs) >= t.Config.MaxFilesPerTask {
		return errors.New("maximum number of files per task reached")
	}

	t.URLs = append(t.URLs, url)
	return nil
}

func (t *Task) Run() {
	t.mu.Lock()
	t.Status = StatusRunning
	t.mu.Unlock()

	defer func() {
		close(t.done)
	}()

	zipFilename := fmt.Sprintf("%s.zip", t.ID)
	zipFile, err := os.Create(zipFilename)
	if err != nil {
		t.setError(fmt.Sprintf("Error creating zip file: %v", err))
		t.setStatus(StatusFailed)
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// ctx, cancel := context.WithCancel(context.Background()) //  Удаляем, так как он уже есть в структуре Task
	// defer cancel()

	var wg sync.WaitGroup

	for _, url := range t.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			select {
			case <-t.ctx.Done(): // Используем t.ctx
				log.Printf("Task %s: Download cancelled for %s", t.ID, url)
				return // Exit goroutine if cancelled
			default:
				filename := filepath.Base(url)
				ext := filepath.Ext(filename)

				if !t.isAllowedExtension(ext) {
					t.setError(fmt.Sprintf("Task %s: File type %s not allowed for %s", t.ID, ext, url))
					return
				}

				err := t.downloadAndAdd(t.ctx, zipWriter, url) // Используем t.ctx
				if err != nil {
					t.setError(fmt.Sprintf("Task %s: Error downloading %s: %v", t.ID, url, err))
					return
				}
			}
		}(url)
	}

	wg.Wait() // Wait for all downloads to complete (or be cancelled)

	t.mu.Lock()
	if len(t.Errors) > 0 {
		t.Status = StatusFailed
	} else {
		t.Status = StatusCompleted
		t.ResultURL = "/" + zipFilename // Relative URL to download the zip
	}
	t.mu.Unlock()
}

func (t *Task) downloadAndAdd(ctx context.Context, zipWriter *zip.Writer, url string) error {
	log.Printf("Task %s: Downloading %s", t.ID, url)

	// Create a context with a timeout
	downloadCtx, cancel := context.WithTimeout(ctx, 30*time.Second) // Set timeout to 30 seconds
	defer cancel()

	req, err := http.NewRequestWithContext(downloadCtx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	filename := filepath.Base(url)

	zipEntry, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(zipEntry, resp.Body)
	if err != nil {
		return err
	}
	log.Printf("Task %s: Downloaded and added %s", t.ID, url)

	return nil
}

func (t *Task) isAllowedExtension(ext string) bool {
	ext = strings.ToLower(ext)
	for _, allowedExt := range t.Config.AllowedExtensions {
		if strings.ToLower(allowedExt) == ext {
			return true
		}
	}
	return false
}

func (t *Task) setError(err string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Errors = append(t.Errors, err)
}

func (t *Task) setStatus(status TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status
}

func (t *Task) GetStatus() (TaskStatus, string, []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Status, t.ResultURL, t.Errors
}

func (t *Task) Cancel() {
	t.cancel() // Cancel the context, stopping downloads
	t.setStatus(StatusFailed)
	log.Printf("Task %s: Cancelled", t.ID)
}

func (t *Task) GetCreationTime() time.Time {
	return t.creationTime
}
