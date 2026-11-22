# Categories

These routes are responsible for category management. They allow to create, modify and delete categories and particular translations of them. Keep in mind that separate translations are treated as one entity. This means if some translation changes might be reflected on other translations

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/categories/{lang}` | [List categories](#get-cmscategorieslang) |
| GET | `/cms/categories/{lang}/{slug}` | [Get a category](#get-cmscategorieslangslug) |
| POST | `/cms/categories/{lang}` | [Create a category](#post-cmscategorieslang) |
| POST | `/cms/categories/{lang}/search` | [Search categories](#post-cmscategorieslangsearch) |
| PATCH | `/cms/categories/{lang}/{slug}` | [Update a category](#patch-cmscategorieslangslug) |
| DELETE | `/cms/categories/{slug}` | [Delete a category](#delete-cmscategoriesslug) |

## Routes

### `GET /cms/categories/{lang}`

List all categories. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if category retreival succeeds

  ```json
  {
    "categories": [
      {
        "name": "Category",
        "slug": "category",
        "lang": "en",
        "hidden": false,
        "navbar": true
      },
      {
        "name": "Another category",
        "slug": "another-category",
        "lang": "en",
        "hidden": false,
        "navbar": false
      }
    ]
  }
  ```

  - `400 Bad Request` — if target language is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `GET /cms/categories/{lang}/{slug}`

Get category translation by slug. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `slug` — target category slug.
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if category retreival succeeds

  ```json
  {
    "name": "Category",
    "slug": "category",
    "lang": "en",
    "hidden": false,
    "navbar": true
  }
  ```

  - `400 Bad Request` — if slug or language is missing or invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `404 Not Found` — if category translataion not found. Note that it does not mean that requested `{slug}` does not exist. Returns a plain-text message.

### `POST /cms/categories/{lang}`

Create a category. Note that creating existing slug with different translation succeeds without affecting shared data (`hidden`, `navbar` flags).

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:**

  ```json
  {
    "slug": "string",
    "name": "string", // optional
    "hidden": boolean, // optional
    "navbar": boolean, // optional
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — if category was created successfully.

  ```json
  {
    "slug": "string", // provided slug in canonical form
    "location": "/cms/categories/{lang}/{slug}"
  }
  ```

  - `400 Bad Request` — if request body, parameters or language are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `409 Conflict` — if category translation already exists. Note that existing slug with different language does not result this code. Returns a plain-text message.

### `POST /cms/categories/{lang}/search`

Search categories. Search is performed on slugs and names.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** Omitted/null fields are defaulted to zero-values unless stated otherwise.

  ```json
  {
    "query": "string",
    "sort_by": "column:asc|desc",  // sortable columns: slug, name
    "page": 1,  // defaults to 1, clamped to 1
    "count": 15,  // defaults to 1, clamped to [1, 100]
    "hidden": boolean, // ignored if omitted
    "navbar": boolean, // ignored if omitted
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if search completed successfully

  ```json
  {
    "results": [
      {
        "name": "Category",
        "slug": "category",
        "lang": "en",
        "hidden": false,
        "navbar": true
      },
      {
        "name": "Another category",
        "slug": "another-category",
        "lang": "en",
        "hidden": false,
        "navbar": false
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body, parameters or language are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `PATCH /cms/categories/{lang}/{slug}`

Update target category slug. Note that if provided language is new, this will create translation for that category.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `slug` — target category slug.
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** Omitted/null fields are ignored at update.

```json
  {
    "name": "Category",
    "slug": "category",
    "hidden": false,
    "navbar": true
  }
```

- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if category update succeeds. Returns empty body.
  - `400 Bad Request` — if request body, parameters, slug or language is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target category slug not found. Returns a plain-text message.
  - `409 Conflict` — if target slug rename conflicts with existing category. Returns a plain-text message.

### `DELETE /cms/categories/{slug}`

Delete category slug. Note that this deletes all of the category translations.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `slug` — target category slug.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if category deleted successfully. Returns empty body.
  - `400 Bad Request` — if slug is missing or invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — target category slug was not found. Returns a plain-text message.
