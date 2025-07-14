# Archive Creation Service

This service creates ZIP archives from a list of URLs.

## Features

*   Creates tasks for archive creation.
*   Adds URLs to a task.
*   Handles inaccessible resources gracefully, packaging available resources.
*   Configurable limits: 3 objects per archive, specific file types (.pdf, .jpeg, .jpg).
*   Provides task status, including a link to the archive upon completion.
*   Manages concurrent tasks, limiting to 3 active archives at a time.
*   Uses standard practices and patterns, such as:
    *   Configuration via YAML file.
    *   Gorilla Mux for routing.
    *   Concurrency management with goroutines and channels.
    *   Error handling and reporting.

## Configuration

The service is configured using a `config.yaml` file:

```yaml
port: ":8080"
max_tasks: 3
max_files_per_task: 3
allowed_extensions:
  - .pdf
  - .jpeg
    • port: The port the server will listen on (e.g., “:8080”).
    • max_tasks: The maximum number of concurrent archive creation tasks.
    • max_files_per_task: The maximum number of files allowed per archive.
    • allowed_extensions: A list of allowed file extensions (lowercase).
API Endpoints
    • POST /task: Create a new task.
        ◦ Returns: {"task_id": "<uuid>"}
    • GET /task/{id}: Get the status of a task.
        ◦ Returns:
          json
          {
            "status": "pending | running | completed | failed",
            "result_url": "/<filename>.zip" (only if status is "completed"),
            "errors": ["list", "of", "errors"]
          }
    • POST /task/{id}/url: Add a URL to the task.
        ◦ Body: {"url": "<url>"}
        ◦ Returns: {"message": "URL added"}
Usage
    1. Build the application:
       bash
       go build -o archive-service main.go
    2. Run the application:
       bash
       ./archive-service
       The server will start on the port specified in config.yaml (default: 8080).
Example Workflow
    1. Create a task:
       bash
       curl -X POST http://localhost:8080/task
       Response:
       json
       { "task_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef" }
    2. Add URLs to the task (repeat up to max_files_per_task times):
       bash
       curl -X POST http://localhost:8080/task/a1b2c3d4-e5f6-7890-1234-567890abcdef/url \
         -H "Content-Type: application/json" \
         -d '{"url": "https://www.easygifanimator.net/images/samples/video-to-gif-sample.gif"}'
    3. Check the task status:
       bash
       curl http://localhost:8080/task/a1b2c3d4-e5f6-7890-1234-567890abcdef
       Possible responses:
        ◦ Pending:
          json
          { "status": "pending", "errors": [] }
        ◦ Running:
          json
          { "status": "running", "errors": [] }
        ◦ Completed:
          json
          { "status": "completed", "result_url": "/static/a1b2c3d4-e5f6-7890-1234-567890abcdef.zip", "errors": [] }
        ◦ Failed:
          json
          { "status": "failed", "errors": ["list", "of", "errors"] }
    4. Download the archive (if the status is “completed”):
       bash
       curl -O http://localhost:8080/static/a1b2c3d4-e5f6-7890-1234-567890abcdef.zip
Notes
    • The service saves created ZIP archives in the static/ directory.
    • Error handling includes reporting inaccessible resources while packaging available ones.
    • Concurrency is managed to allow a maximum of 3 active archive creation tasks.
# 14.07.2025
