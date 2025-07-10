# MiniVault API

A lightweight local REST API that simulates ModelVault's prompt-response functionality. This implementation includes both Ollama-based LLM responses and fallback stub responses.

## Features

- POST `/generate` endpoint for prompt-response interaction
- POST `/generate/stream` endpoint for streaming responses using chunked transfer
- Local LLM integration via Ollama (with automatic fallback)
- Chunked streaming with newline-delimited JSON
- Detailed request/response logging with system metrics
- Clean error handling
- Modular code structure
- Comprehensive test suite
- Postman collection for easy testing
- Swagger documentation
- Nginx reverse proxy
- Docker support with health checks

## Quick Start with Docker

The easiest way to run the entire stack is using Docker Compose:

```bash
# Pull and start all services
docker-compose up -d

# Watch Ollama logs to see progress
docker-compose logs -f ollama
```

This will start:
- The API service (with automatic Ollama/stub fallback)
- Nginx reverse proxy
- Ollama LLM service with smollm:135m model

The API will be available at:
- Main API: http://localhost/
- Swagger UI: http://localhost/swagger/index.html

### Environment Variables

The API service supports the following environment variables:
- `LLM_TYPE`: LLM implementation to use ("ollama" or "stub", default: "ollama")
- `OLLAMA_HOST`: Ollama server URL (default: http://localhost:11434)
- `OLLAMA_MODEL`: Ollama model to use (default: smollm:135m)
- `PORT`: Server port (default: 8080)

## API Usage

### Generate Response (Non-Streaming)

**Endpoint:** `POST /generate`

**Request:**
```bash
curl -X POST http://localhost/generate \
    -H "Content-Type: application/json" \
    -d '{"prompt": "Tell me a joke"}'
```

**Response:**
```json
{
    "response": "Generated text response"
}
```

### Generate Response (Streaming)

**Endpoint:** `POST /generate/stream`

The streaming endpoint uses HTTP chunked transfer encoding to send responses as newline-delimited JSON.

**Request:**
```bash
# Using curl
curl -N -X POST http://localhost/generate/stream \
    -H "Content-Type: application/json" \
    -d '{"prompt": "Tell me a story"}'
```

**Response Format:**
```jsonl
{"token":"Once"}
{"token":"upon"}
{"token":"a"}
{"token":"time"}
...
```

## Logging

All interactions are logged to `logs/log.jsonl` in a detailed JSONL format:

```json
{
    "id": "1704067200-12345",           // Unique request ID
    "timestamp": "2024-01-01T12:00:00Z", // ISO 8601 timestamp
    "duration_ms": 150,                  // Request duration

    "prompt": "Tell me a joke",          // Input prompt
    "llm_type": "ollama",               // LLM implementation used
    "llm_model": "smollm:135m",         // Model name (for Ollama)
    "streaming": false,                  // Whether streaming was used

    "response": "Why did...",           // Generated response
    "token_count": 15,                  // Number of tokens in response
    "response_size": 85,                // Response size in bytes

    "success": true,                    // Request success status
    "error": "error message",           // Error message if any

    "go_version": "go1.22",             // Go runtime version
    "goroutines": 10,                   // Active goroutines
    "memory_bytes": 1048576             // Memory usage
}
```

### Log Fields

1. Request Details:
   - `id`: Unique request identifier
   - `timestamp`: Request time in ISO 8601 format
   - `duration_ms`: Request processing time in milliseconds

2. Input Details:
   - `prompt`: The input prompt
   - `llm_type`: Type of LLM used ("ollama" or "stub")
   - `llm_model`: Model name when using Ollama
   - `streaming`: Whether streaming was used

3. Response Details:
   - `response`: Generated text
   - `token_count`: Approximate token count
   - `response_size`: Response size in bytes

4. Status Details:
   - `success`: Whether request succeeded
   - `error`: Error message if failed

5. System Metrics:
   - `go_version`: Go runtime version
   - `goroutines`: Number of active goroutines
   - `memory_bytes`: Memory usage in bytes

### Log Analysis

The JSONL format makes it easy to analyze logs using standard tools:

```bash
# Count successful vs failed requests
cat logs/log.jsonl | jq 'select(.success == true) | length'

# Average response time
cat logs/log.jsonl | jq '.duration_ms' | awk '{sum+=$1} END {print sum/NR}'

# Memory usage over time
cat logs/log.jsonl | jq -r '[.timestamp, .memory_bytes] | @tsv'
```

**JavaScript Client Example:**
```javascript
async function streamText(prompt) {
    const response = await fetch('http://localhost/generate/stream', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ prompt })
    });

    // Read response as newline-delimited JSON
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
        const {value, done} = await reader.read();
        if (done) break;
        
        // Add new data to buffer and split by newlines
        buffer += decoder.decode(value, {stream: true});
        const lines = buffer.split('\n');
        
        // Process complete lines
        for (let i = 0; i < lines.length - 1; i++) {
            if (lines[i].trim()) {
                const token = JSON.parse(lines[i]).token;
                // Use the token (e.g., append to UI)
                console.log(token);
            }
        }
        
        // Keep incomplete line in buffer
        buffer = lines[lines.length - 1];
    }
}

// Usage
streamText("Tell me a story").catch(console.error);
```

**Python Client Example:**
```python
import requests
import json

def stream_text(prompt):
    response = requests.post(
        'http://localhost/generate/stream',
        json={'prompt': prompt},
        stream=True
    )
    
    for line in response.iter_lines():
        if line:
            token = json.loads(line)['token']
            # Use the token
            print(token, end='', flush=True)

# Usage
stream_text("Tell me a story")
```

## Development

### Running Tests

The project includes comprehensive tests:

```bash
# Run all tests
go test ./... -v

# Run specific test suite
go test ./src/api -v
go test ./src/service -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out  # View coverage in browser
```

### Test Structure

1. Unit Tests:
   - API Handlers (`src/api/handlers_test.go`)
   - Service Layer (`src/service/generator_test.go`)
   - Logger (`src/service/logger_test.go`)

2. Integration Tests:
   - End-to-end tests (`e2e/e2e_test.go`)

3. Mock Implementations:
   - MockGenerator for LLM interface
   - MockLogger for logging service

### Docker Volumes

The project uses named Docker volumes for persistence:

1. `api_logs`: API service logs
   - Contains request/response logs
   - JSONL format with detailed metrics
   - Persists across container restarts

2. `nginx_logs`: Nginx access and error logs
   - Standard Nginx log format
   - Useful for monitoring API usage

3. `ollama_data`: Ollama model storage
   - Contains downloaded models
   - Persists model data across restarts
   - Shared between containers if needed

To manage volumes:
```bash
# List volumes
docker volume ls

# Inspect logs
docker exec minivault-api cat /app/logs/log.jsonl
docker exec minivault-nginx cat /var/log/nginx/access.log

# Clean up volumes (caution: this deletes all data)
docker-compose down -v
```

### Project Structure

```
minivault-api/
├── src/
│   ├── api/          # API handlers and routing
│   ├── service/      # Business logic and services
│   ├── llm/          # LLM implementations
│   └── types/        # Data types and models
├── e2e/              # End-to-end tests
├── nginx/            # Nginx configuration
├── postman/          # Postman collection
├── logs/             # Request logs
├── main.go           # Application entry point
├── Dockerfile        # Docker build file
├── docker-compose.yml # Docker Compose configuration
├── go.mod            # Go module file
└── README.md         # This file
```

## Implementation Notes

### Architecture
- Clean separation between LLM interface and implementations
- Pluggable LLM system for easy addition of new providers
- Chunked transfer streaming for real-time updates
- Automatic fallback mechanism for resilience
- Health checks for all services

### Design Choices
- Used Gin framework for its performance and ease of use
- Implemented JSONL logging for easy parsing and analysis
- Chunked transfer encoding for streaming (instead of SSE)
- Integrated with Ollama for local LLM support
- Nginx reverse proxy for production-ready setup
- Docker support with health checks and proper service dependencies

### Streaming Implementation
- Uses standard HTTP chunked transfer encoding
- Newline-delimited JSON format for tokens
- No special event formatting needed
- Works with standard HTTP clients
- Supports backpressure naturally

### Docker Configuration
- API service waits for Ollama to be healthy
- Ollama service includes model pulling
- Nginx configured for chunked transfer
- All services include proper restart policies
- Volumes for persistent logs and Ollama data

### Future Improvements
1. Add response caching
2. Implement request rate limiting
3. Add support for multiple LLM models
4. Add configuration file for easy customization
5. Implement response validation
6. Add metrics and monitoring

## Error Handling

The API handles several error cases:
- Invalid JSON format
- Empty prompts
- LLM failures (with automatic fallback)
- Server errors
- Logging failures

## License

MIT