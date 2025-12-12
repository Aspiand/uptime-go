# API Documentation

---

## Endpoint List

- [Health Check](#health-check)
- [Update Configuration](#update-configuration)
- [Monitoring Reports](#monitoring-reports)

---

## Health Check

This endpoint is used to verify the service's operational status.

- **Endpoint:** `GET /health`
- **Method:** `GET`

### Success Response

- **Code:** `200 OK`
- **Payload:** A static JSON object indicating the service is healthy.
  ```json
  {
    "status": "healthy",
    "service": "uptime-go"
  }
  ```

---

## Update Configuration

This endpoint is used to update the application's configuration file. **The application needs to be restarted** for the changes to take effect.

- **Endpoint:** `POST /api/uptime-go/config`
- **Method:** `POST`

### Request Body

The request body must contain the raw content of a valid YAML configuration file.

- **Data Type:** `text/yaml`

### Example Payload

The payload should follow the structure of `configs/uptime.yml`.

```yaml
monitor:
  - url: "http://example.com"
    enabled: true
    interval: 5m
    response_time_threshold: 5s
    certificate_monitoring: true
    certificate_expired_before: 31d
  - url: "https://google.com"
    enabled: true
    interval: 10m
    response_time_threshold: 2s
```

### Success Response

- **Code:** `200 OK`
- **Payload:**
  ```json
  {
    "message": "Configuration updated successfully. Please restart the application to apply changes."
  }
  ```

### Possible Errors

- **Code:** `400 Bad Request`
  - **Condition:** The request body cannot be read.
  - **Payload:**
    ```json
    {
      "message": "Failed to read request body",
      "error": "<error detail>"
    }
    ```

- **Code:** `500 Internal Server Error`
  - **Condition:** The server fails to write the new configuration to the file.
  - **Payload:**
    ```json
    {
      "message": "Failed to update configuration",
      "error": "<error detail>"
    }
    ```

---

## Monitoring Reports

This endpoint retrieves monitoring data. The JSON response structure is based on the `Monitor` and `MonitorHistory` models, excluding fields marked with `json:"-"`.

- **Endpoint:** `GET /api/uptime-go/reports`
- **Method:** `GET`

### Query Parameters

| Parameter | Type   | Optional | Default | Description                                                                            |
| :-------- | :----- | :------- | :------ | :------------------------------------------------------------------------------------- |
| `url`     | string | Yes      | (none)  | The URL of a specific monitor. If omitted, returns a list of all monitors.             |
| `limit`   | int    | Yes      | 1000    | Limits the number of history records for a specific monitor. Ignored if `url` is omitted. |

### Example Usage

1.  **Get all monitors:**
    `GET /api/uptime-go/reports`

2.  **Get a specific monitor:**
    `GET /api/uptime-go/reports?url=http://example.com`

3.  **Get a specific monitor with the last 20 history records:**
    `GET /api/uptime-go/reports?url=http://example.com&limit=20`

### Success Response (`200 OK`)

- **Payload (if `url` is omitted):**
  Returns an array of monitor objects. The `histories` field is not included.
  ```json
  [
    {
      "url": "http://example.com",
      "is_up": true,
      "status_code": 200,
      "response_time": 150,
      "certificate_expired_date": "2026-03-10T00:00:00Z",
      "last_up": "2025-12-02T10:20:00Z",
      "last_down": "2025-11-30T14:00:00Z",
      "last_check": "2025-12-02T10:20:00Z"
    }
  ]
  ```

- **Payload (if `url` is provided):**
  Returns a single monitor object, including its recent history.
  ```json
  {
    "url": "http://example.com",
    "is_up": true,
    "status_code": 200,
    "response_time": 150,
    "certificate_expired_date": "2026-03-10T00:00:00Z",
    "last_up": "2025-12-02T10:20:00Z",
    "last_down": "2025-11-30T14:00:00Z",
    "last_check": "2025-12-02T10:20:00Z",
    "histories": [
      {
        "is_up": true,
        "response_time": 150,
        "created_at": "2025-12-02T10:20:00Z"
      },
      {
        "is_up": true,
        "response_time": 145,
        "created_at": "2025-12-02T10:15:00Z"
      }
    ]
  }
  ```

### Possible Errors

- **Code:** `400 Bad Request`
  - **Condition:** The query parameters are invalid (e.g., `limit` is not a number).
  - **Payload:**
    ```json
    {
      "message": "Invalid query parameters",
      "error": "<error detail>"
    }
    ```

- **Code:** `404 Not Found`
  - **Condition:** The monitor specified by the `url` parameter does not exist.
  - **Payload:**
    ```json
    {
      "message": "Record not found"
    }
    ```

- **Code:** `500 Internal Server Error`
  - **Condition:** An error occurred while fetching data from the database.
  - **Payload:**
    ```json
    {
      "message": "Failed to retrieve monitor details",
      "error": "<error detail>"
    }
    ```