# Storage API reference

This document covers CMS routes that are responsible for file storage access and management.
Return to the [route index](README.md).

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/storage/tree/{path...}` | [List a storage directory](#get-storagetreepath) |
| GET | `/storage/media/{path...}` | [Read a storage file](#get-storagemediapath) |
| PUT | `/storage/media/{path...}` | [Upload a storage file](#put-storagemediapath) |
| DELETE | `/storage/media/{path...}` | [Delete a storage path](#delete-storagemediapath) |
| PUT | `/storage/folders/{path...}` | [Create a storage folder](#put-storagefolderspath) |
| GET | `/storage/preview/{path...}` | [Read a storage preview path](#get-storagepreviewpath) |

## Routes

### `GET /storage/tree/{path...}`

List the contents of a directory and report storage capacity.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `path` — target path
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if the operation succeeds.

  ```json
    {
      "files": [
        {
          "id": "folder/file.jpg",
          "type": "image", // "file" | "folder" | "image" | "video"
          "size": 123,
          "modified": "2025-01-01T00:00:00Z"
        }
      ],
      "info": { "total": 1000, "used": 100, "free": 900 }
    }
  ```

  - `400 Bad Request` — if the requested path is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `404 Not Found` — if the requested path does not exist. Returns a plain-text message.

### `GET /storage/media/{path...}`

Stream a stored file.

- **Headers:** Not required.
- **Path parameter:**
  - `path` — target file path
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Raw file bytes with the detected `Content-Type`. Streaming may fail after the `200` response.
  - `400 Bad Request` — if the requested path is invalid. Returns a plain-text message.
  - `404 Not Found` — if the requested path does not exist. Returns a plain-text message.

### `PUT /storage/media/{path...}`

Create or replace a file. Handler reads raw request bytes rather than a multipart form.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameter:**
  - `path` — target file path
- **Request body:** Raw file bytes.
- **Request limit:** 512 MiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — if the requested path is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `DELETE /storage/media/{path...}`

Delete a file or directory.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameter:**
  - `path` — target file path
- **Query parameters:**
  - `recursive` — If `false`, the request refuses to delete a non-empty directory. This parameter does not affect files.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Empty body. Also returned for a non-existent path if the `recursive` query parameter is set to `true`.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `400 Bad Request` — if the requested path or `recursive` parameter is invalid, or if the directory is not empty while `recursive=false`. Returns a plain-text message.
  - `404 Not Found` — if the requested path does not exist and `recursive=false`. Returns a plain-text message.

### `PUT /storage/folders/{path...}`

Create a directory.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameter:**
  - `path` — destination directory path
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Empty body.
  - `400 Bad Request` — if the requested path is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `GET /storage/preview/{path...}`

Get a preview of an image file.

- **Headers:** Not required.
- **Path parameter:**
  - `path` — target image file path
- **Query parameters:**
  - `w` — preview width. Default: 360 px.
  - `h` — preview height. Default: 360 px.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — Raw file bytes with `Content-Type: image/webp`.
  - `400 Bad Request` — if the requested path, `w`, or `h` is invalid, or if the target file is not an image. Returns a plain-text message.
  - `404 Not Found` — if the requested path does not exist. Returns a plain-text message.
