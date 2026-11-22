# Advertisements

These routes are responsible for advertisements management.

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/ads/groups` | [List advertisement groups](#get-cmsadsgroups) |
| GET | `/cms/ads/{id}` | [Get an advertisement](#get-cmsadsid) |
| POST | `/cms/ads` | [Create an advertisement](#post-cmsads) |
| POST | `/cms/ads/search` | [Search advertisements](#post-cmsadssearch) |
| PATCH | `/cms/ads/{id}` | [Update an advertisement](#patch-cmsadsid) |
| DELETE | `/cms/ads/{id}` | [Delete an advertisement](#delete-cmsadsid) |

## Routes

### `GET /cms/ads/groups`

Get all group values. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if group retreival succeeds

  ```json
  {
    "groups": [
      "array",
      "of",
      "strings"
    ]
  }
  ```

  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `GET /cms/ads/{id}`

Get advertisement by id. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — target advertisement ID, positive integer value
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if advertisement retreival succeeds

  ```json
  {
    "id": 1,
    "name": "string",
    "description": "string",
    "url": "https://outer-domain.com/referral",
    "priority": 1, // unsigned integer
    "group": "string",
    "expires_at": 2141453433, // Unix timestamp
    "banner": "https://domain.com/media/ad_banner.webp"
  }
  ```

  - `400 Bad Request` — if advertisement ID is missing or invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `404 Not Found` — if advertisement was not found. Returns a plain-text message.

### `POST /cms/ads`

Create an advertisement. Note that subsequent calls to this route event with existing advertisement ID results a new advertisement.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:** None
- **Request body:**

  ```json
  {
    "name": "string",
    "description": "string",  // optional
    "url": "https://outer-domain.com/referral",
    "priority": 1, // optional, unsigned integer 
    "group": "string", // optional
    "expires_at": 2141453433, // optional, Unix timestamp, null value means ad never expires
    "banner": "https://domain.com/media/ad_banner.webp"
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `201 Created` — if category was created successfully.

  ```json
  {
    "id": "id", // assigned id
    "location": "/cms/ads/{id}"
  }
  ```

  - `400 Bad Request` — if request body or parameters are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.

### `POST /cms/ads/search`

Search ads. Search is performed on ids, name and descriptions.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:** None
- **Request body:** Omitted/null fields are defaulted to zero-values unless stated otherwise.

  ```json
  {
    "query": "string",
    "sort_by": "column:asc|desc",  // sortable columns: id, name, priority, expires_at, group
    "page": 1,  // defaults to 1, clamped to 1
    "count": 15,  // defaults to 1, clamped to [1, 100]
    "groups": ["array", "of", "group", "slugs"], // ignored if empty or omitted
    "expired": boolean, // ignored if omitted
    "expires_after": 2141453433,  // Unix timestamp, ignored if omitted
    "expires_before": 2141453433,  // Unix timestamp, ignored if omitted
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if search completed successfully

  ```json
  {
    "results": [
      {
        "id": 1,
        "name": "string",
        "description": "string",
        "url": "https://outer-domain.com/referral",
        "priority": 1, // unsigned integer
        "group": "string",
        "expires_at": 2141453433, // Unix timestamp
        "banner": "https://domain.com/media/ad_banner.webp"
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body or parameters are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `PATCH /cms/ads/{id}`

Update advertisement.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `id` — target advertisement, positive integer value
- **Request body:** Omitted/null fields are ignored at update unless stated otherwise.

```json
  {
    "name": "string",
    "description": "string",
    "url": "https://outer-domain.com/referral",
    "priority": 1, // unsigned integer
    "group": "string",
    "expires_at": 2141453433, // Unix timestamp, always assigned, defaults to null if omitted
    "banner": "https://domain.com/media/ad_banner.webp"
  }
```

- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if advertisement update succeeds. Returns empty body.
  - `400 Bad Request` — if request body, parameters or advertisement ID is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target advertisement ID not found. Returns a plain-text message.

### `DELETE /cms/ads/{id}`

Deletes target advertisement.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — target advertisement ID, positive integer value
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if advertisement deleted successfully. Returns empty body.
  - `400 Bad Request` — if advertisement ID is missing or invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — target advertisement was not found. Returns a plain-text message.
