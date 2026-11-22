# Authentication

These routes are responsible for user authentication.

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/profile` | [Get user profile](#get-cmsprofile) |
| GET | `/cms/permissions` | [Get list of defined permissions](#get-cmspermissions) |
| POST | `/cms/login` | [Authenticate user](#post-cmslogin) |
| POST | `/cms/logout` | [Logout from CMS](#post-cmslogout) |
| PATCH | `/cms/password` | [Change user's password](#patch-cmspassword) |

## Routes

### `GET /cms/profile`

Get current authenticated user data.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if profile retreival succeeds

  ```json
  {
    "username": "string",
    "email": "email",
    "permissions": 4294967295,  // bit field as number
    "last_login": 2141453433,  // Unix timestamp
  }
  ```

  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `GET /cms/permissions`

List all configured permissions.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:**

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` —

  ```json
  {
    "entities": {
      "posts": 0,  // bit shift
      "categories": 3,
      "advertisements": 6
    },
    "actions": {
      "create": 1,  // bit mask
      "edit": 2,
      "delete": 4,
    }
  },
  ```

  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `POST /cms/login`

Authenticate user.

- **Headers:**
  - `Content-Type` — `application/json`
- **Path parameters:** None.
- **Request body:**

  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `200 OK` — if login succeeds.

  ```json
  {
    "token": "string", // authentication token, attach to `Authorization` header.
  }
  ```

  - `400 Bad Request` — if request body or parameters are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if login credentails is invalid. Returns a plain-text message.

### `POST /cms/logout`

Log out user.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:** None.
- **Request body:** None.
- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — on successful logout

### `PATCH /cms/password`

Change authenticated user's password.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:** None.
- **Request body:**

  ```json
  {
    "old_password": "string",
    "new_password": "string"
  }
  ```

- **Request limit:** 4 KiB.
- **Response:**
  - `204 No Content` — if password was changed successfully.
  - `400 Bad Request` — if request body or parameters are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if provided old password is invalid. Returns a plain-text message.
