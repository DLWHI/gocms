# Storage API reference

This document covers CMS routes that are responsible for file storage access and management.
Return to the [route index](README.md).

## Routes

<a id="storage-tree"></a>

### `GET /storage/tree/{path...}`

List the contents of a directory and report storage capacity.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`.
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
  - `401 Unauthorized` — if the authorization header is invalid or missing. Returns a plain-text message.
  - `404 Not Found` — if the requested path does not exist. Returns a plain-text message.
  - `500 Internal Server Error` — if listing entries, retrieving storage capacity, or validating the session fails. Returns a plain-text message.

<a id="storage-media-get"></a>

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
  - `500 Internal Server Error` — if reading the file fails. Returns a plain-text message.

<a id="storage-media-put"></a>

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
  - `401 Unauthorized` — if the authorization header is invalid or missing. Returns a plain-text message.
  - `500 Internal Server Error` — if writing the file or validating the session fails. Returns a plain-text message.

<a id="storage-media-delete"></a>

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
  - `400 Bad Request` — if the requested path or `recursive` parameter is invalid, or if the directory is not empty while `recursive=false`. Returns a plain-text message.
  - `404 Not Found` — if the requested path does not exist and `recursive=false`. Returns a plain-text message.
  - `401 Unauthorized` — if the authorization header is invalid or missing. Returns a plain-text message.
  - `500 Internal Server Error` — if deleting the file fails. Returns a plain-text message.

<a id="storage-folders"></a>

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
  - `401 Unauthorized` — if the authorization header is invalid or missing. Returns a plain-text message.
  - `500 Internal Server Error` — if creating the folder or validating the session fails. Returns a plain-text message.

<a id="storage-preview"></a>

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
  - `500 Internal Server Error` — if preview generation fails. Returns a plain-text message.
