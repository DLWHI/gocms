# Languages

These routes are responsible for language management. Languages are required for entites to be translatable. Entity cannot hold language that is not in cms record.

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/languages` | [List languages](#get-cmslanguages) |
| POST | `/cms/languages` | [Create a language](#post-cmslanguages) |
| DELETE | `/cms/languages/{code}` | [Delete a language](#delete-cmslanguagescode) |

## Routes

### `GET /cms/languages`

List all languages. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if languages retreival succeeds

  ```json
  {
    "languages": [
      {
        "name": "English",
        "code": "en",
      },
      {
        "name": "Русский",
        "code": "ru",
      },
      {
        "name": "Deutsch",
        "code": "de",
      },
    ]
  }
  ```

  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `POST /cms/languages`

Create a language. Entites are available to be translated to that language after creation.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:** None.
- **Request body:**

  ```json
  {
    "name": "English",  // optional
    "code": "en",  // ISO 639-1 language code
  },
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — if language was created successfully. Returns empty body.
  - `400 Bad Request` — if request body or request parameters are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `409 Conflict` — if language already exists. Returns a plain-text message.

### `DELETE /cms/languages/{code}`

Delete language. Note that all records in that language are not deleted, but become unaccessable and records in deleted language are preserved if language code is created again. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `code` — target language code.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if language deleted successfully. Returns empty body.
  - `400 Bad Request` — if language code is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — target language was not found. Returns a plain-text message.
