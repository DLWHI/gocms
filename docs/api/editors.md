# Editors

These routes are responsible for editors management. Editors allow posts to have author, users that own rights to editor may publish posts from editor's name. These routes allow to create, modify, delete and bind editor rights and particular translations of them. Keep in mind that separate translations are treated as one entity. This means if some translation changes might be reflected on other translations

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/editors/{lang}` | [List editors](#get-cmseditorslang) |
| GET | `/cms/editors/{lang}/{slug}` | [Get editor](#get-cmseditorslangslug) |
| POST | `/cms/editors/{lang}` | [Create editor](#post-cmseditorslang) |
| POST | `/cms/editors/{lang}/search` | [Search editors](#post-cmseditorslangsearch) |
| PATCH | `/cms/editors/{lang}/{slug}` | [Update editor](#patch-cmseditorslangslug) |
| DELETE | `/cms/editors/{slug}` | [Delete editor](#delete-cmseditorsslug) |

## Routes

### `GET /cms/editors/{lang}`

List editors. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if editor retreival succeeds

  ```json
  {
    "editors": [
      {
        "name": "Editor",
        "slug": "editor",
        "lang": "en",
      },
      {
        "name": "Another editor",
        "slug": "another-editor",
        "lang": "en",
      }
    ]
  }
  ```

  - `400 Bad Request` — if target language is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `GET /cms/editors/{lang}/{slug}`

Get editor translation by slug. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `slug` — target editor slug.
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if editor retreival succeeds

  ```json
  {
    "name": "Editor",
    "slug": "editor",
    "lang": "en",
    "bio": "Bio text in any format, supplied as string", 
    "deleted_at": 2141453433 // Unix timestamp
  }
  ```

  - `400 Bad Request` — if slug or language is missing or invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `404 Not Found` — if editor translataion not found. Note that it does not mean that requested `{slug}` does not exist. Returns a plain-text message.

### `POST /cms/editors/{lang}`

Create editor. Note that creating existing editor with different translation succeeds without affecting shared data.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:**

  ```json
  {
    "slug": "string",
    "bio": "Bio text in any format, supplied as string", // optional
    "name": "string", // optional
  }
  ```

- **Request limit:** 2 MiB.
- **Response:**
  - `201 Created` — if editor was created successfully.

  ```json
  {
    "slug": "string", // provided slug in canonical form
    "location": "/cms/editors/{lang}/{slug}"
  }
  ```

  - `400 Bad Request` — if request body, parameters or language are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `409 Conflict` — if editor already exists. Note that existing editor slug with different language does not result this code. Returns a plain-text message.

### `POST /cms/editors/{lang}/search`

Search editors. Search is performed on slugs and names.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** Omitted/null fields are defaulted to zero-values unless stated otherwise.

  ```json
  {
    "query": "string",
    "sort_by": "column:asc|desc",  // sortable columns: slug, name, deleted_at
    "page": 1, // defaults to 1, clamped to 1
    "count": 15,  // defaults to 1, clamped to [1, 100]
    "deleted": boolean, // ignored if omitted
    "deleted_after": 2141453433,  // Unix timestamp, ignored if omitted
    "deleted_before": 2141453433,  // Unix timestamp, ignored if omitted
  }
  ```

- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if search completed successfully

  ```json
  {
    "results": [
      {
        "name": "Editor",
        "slug": "editor",
        "lang": "en",
      },
      {
        "name": "Another editor",
        "slug": "another-editor",
        "lang": "en",
        "deleted_at": 2141453433 // Unix timestamp
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body, parameters or language are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `PATCH /cms/editors/{lang}/{slug}`

Update target editor slug. Note that if provided language is new, this will create translation for that editor.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `slug` — target editor slug.
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** Omitted/null fields are ignored at update.

```json
  {
    "name": "Editor",
    "slug": "editor",
    "bio": "Bio text in any format, supplied as string"
  }
```

- **Request limit:** 2 MiB.
- **Response:**
  - `204 No Content` — if editor update succeeds. Returns empty body.
  - `400 Bad Request` — if request body, parameters, slug or language is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target editor slug not found. Returns a plain-text message.
  - `409 Conflict` — if target slug rename conflicts with existing editor. Returns a plain-text message.

### `DELETE /cms/editors/{slug}`

Delete editor slug. Note that this soft deletes editor from database, `GET /cms/editors/{slug}` can still retrieve info about editor. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `slug` — target editor slug.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `204 No Content` — if editor deleted successfully. Returns empty body.
  - `400 Bad Request` — if slug is missing or invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — target editor slug was not found. Returns a plain-text message.
