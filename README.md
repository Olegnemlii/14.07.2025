# Archive Creation Service

This service creates ZIP archives from a list of URLs.

## Features

*   Creates tasks for archive creation.
*   Adds URLs to a task.
*   Checks the status of a task.
*   Downloads the resulting ZIP archive.

## Configuration

The service is configured using `config.yaml`:

```yaml
port: 8080
max_tasks: 3
max_files_per_task: 3
allowed_extensions:
  - .pdf
  - .jpeg
  - .jpg
API Endpoints
POST /task: Create a task.
Returns: { "task_id": "<uuid>" }
GET /task/{id}: Get task status.
Returns:
json

{
  "status": "pending | running | completed | failed",
  "result_url": "/<filename>.zip" (if completed),
  "errors": ["list", "of", "errors"],
  "task_id": "<task_id>"
}
POST /task/{id}/url: Add a URL to the task.
Body:
json

{ "url": "<url>" }
Returns: { "message": "URL added" }
Usage
Build: go build -o archive-service main.go
Run: ./archive-service
Example
Create a task:

bash

curl -X POST http://localhost:8080/task
Response:

json

{ "task_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef" }
Add a URL:

bash

curl -X POST http://localhost:8080/task/a1b2c3d4-e5f6-7890-1234-567890abcdef/url \
  -H "Content-Type: application/json" \
  -d '{ "url": "https://example.com/image.jpg" }'
Get status:

bash

curl http://localhost:8080/task/a1b2c3d4-e5f6-7890-1234-567890abcdef
Response (completed):

json

{ "status": "completed", "result_url": "/a1b2c3d4-e5f6-7890-1234-567890abcdef.zip", "task_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef" }
Notes
Basic error handling.
In-memory task queue (no persistence).