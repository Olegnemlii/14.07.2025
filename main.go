package main

import (
	"log"
	"runtime"
	"github.com/Olegnemlii/14.07.2025/config"
    "github.com/Olegnemlii/14.07.2025/task"
)

func main() {
	// Устанавливаем количество ядер, используемых приложением
	runtime.GOMAXPROCS(runtime.NumCPU())

	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	server := NewAPIServer(":8080", config)

	// Run task processor
	go func() {
		for task := range server.taskQueue {
			task.Run()
		}
	}()

	server.Run()
}
