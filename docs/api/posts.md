# Posts

These routes are responsible for posts management: creation, publication, content. Users may create posts if they have sufficent rights, creation timestamp and creator are imprinted inside data and cannot be changed. Post may have editor as author, it is then said that this post is published
on behalf of that editor. Creator may do any actions on their posts, other must have permission to edit/delete them. Keep in mind that separate translations are treated as one entity. This means if some translation changes might be reflected on other translations

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/posts/{lang}/{id}` | [Get a post translation](#get-cmspostslangid) |
| POST | `/cms/posts/{lang}` | [Create a post translation](#post-cmspostslang) |
| POST | `/cms/posts/{lang}/search` | [Search posts](#post-cmspostslangsearch) |
| PATCH | `/cms/posts/{lang}/{id}` | [Update a post translation](#patch-cmspostslangid) |
| PUT | `/cms/posts/views/{id}` | [Set post views](#put-cmspostsviewsid) |
| DELETE | `/cms/posts/{id}` | [Delete a post and its translations](#delete-cmspostsid) |

## Routes

### `GET /cms/posts/{lang}/{id}`

Get editor translation by id. Request body is ignored if sent.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — target post ID, positive integer value
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if editor retreival succeeds

  ```json
  {
    "post_id": 1,
    "title": "string",
    "description": "string",
    "content": "Rich text (usually)",
    "slug": "string",
    "keywords": "comma, separated, strings",
    "language": "en",
    "created_at": 2141453433, // Unix timestamp
    "published_at": 2141453433, // Unix timestamp
    "views": 2341,
    "preview": "https://domain.com/preview/image.webp",
    "media": "https://domain.com/media/post_1.webp",
    "media_source": "string",
    "media_type": "image|video", // or empty string
    "author": "editor-slug", // null value means post has no author
    "created_by": "username",
    "categories": ["array", "of", "category", "slugs"]
  }
  ```

  - `400 Bad Request` — if post ID or language is missing or invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.
  - `404 Not Found` — if post translataion not found. Note that it does not mean that requested `{id}` does not exist. Returns a plain-text message.

### `POST /cms/posts/{lang}`

Create a post along with translation. Note that subsequent calls to this route event with existing post ID and different translation results a new post.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:**

  ```json
  {
    "title": "string", // optional
    "description": "string", // optional
    "content": "Rich text (usually)", // optional
    "keywords": "comma, separated, strings", // optional
    "slug": "string",
    "published_at": 2141453433, // optional, Unix timestamp
    "preview": "https://domain.com/preview/image.webp", // optional
    "media": "https://domain.com/media/post_1.webp", // optional
    "media_source": "string", // optional
    "media_type": "image|video", // optional, may hold empty string
    "author": "editor-slug", // optional, null value means post has no author
    "categories": ["array", "of", "category", "slugs"] // optional
  }
  ```

- **Request limit:** 2 MiB.
- **Response:**
  - `201 Created` — if post was created successfully.
  
  ```json
  {
    "id": 1,  // assigned id
    "location": "/cms/posts/{lang}/{id}"
  }
  ```

  - `400 Bad Request` — if request body, parameters or language are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.

### `POST /cms/posts/{lang}/search`

Search posts. Search is performed on ids, titles, slugs, keywords and descriptions.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** Omitted/null fields are defaulted to zero-values unless stated otherwise.

  ```json
  {
    "query": "string",
    "sort_by": "column:asc|desc",  // sortable columns: post_id, title, author, published_at, created_by, views
    "page": 1,  // defaults to 1, clamped to 1
    "count": 15,  // defaults to 1, clamped to [1, 100]
    "categories": ["array", "of", "category", "slugs"], // might contain hidden categories, ignored if empty or omitted
    "exclude_categories": ["array", "of", "category", "slugs"], // might contain hidden categories, ignored if empty or omitted
    "authors": ["array", "of", "editor", "slugs"], // might contain soft-deleted editors, ignored if empty or omitted
    "created_by": "username", // ignored if omitted
    "published": boolean, // ignored if omitted
    "published_after": 2141453433,  // Unix timestamp, ignored if omitted
    "published_before": 2141453433,  // Unix timestamp, ignored if omitted
    "created_after": 2141453433,  // Unix timestamp, ignored if omitted
    "created_before": 2141453433,  // Unix timestamp, ignored if omitted
  }
  ```

- **Request limit:** 2 MiB.
- **Response:**
  - `200 OK` — if search completed successfully

  ```json
  {
    "results": [
      {
        "post_id": 1,
        "title": "string",
        "slug": "string",
        "keywords": "comma, separated, strings",
        "language": "en",
        "created_at": 2141453433, // Unix timestamp
        "published_at": 2141453433, // Unix timestamp
        "views": 2341,
        "preview": "https://domain.com/preview/image.webp",
        "author": "editor-slug",
        "created_by": "username",
        "categories": ["array", "of", "category", "slugs"]
      }
    ],
    "last_page": 1,
    "total": 1
  }
  ```

  - `400 Bad Request` — if request body, parameters or language are invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity. Returns a plain-text message.

### `PATCH /cms/posts/{lang}/{id}`

Update target post. Note that if provided language is new, this will create translation for that post. Creator may always update his post

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `id` — target post, positive integer value
  - `lang` - target language, in format of ISO 639-1 language code.
- **Request body:** Omitted/null fields are ignored at update unless stated otherwise.

```json
  {
    "title": "string",
    "description": "string",
    "content": "Rich text (usually)",
    "slug": "string",
    "keywords": "comma, separated, strings",
    "published_at": 2141453433, // Unix timestamp, may be earlier than created_at, null value means post is not published, unpublishes post if omitted or null
    "preview": "https://domain.com/preview/image.webp",
    "media": "https://domain.com/media/post_1.webp",
    "media_source": "string",
    "media_type": "image|video", // or empty string
    "author": "editor-slug", // null value means post has no author, removes author if omitted or null
    "categories": ["array", "of", "category", "slugs"] // might contain hidden categories
  }
```

- **Request limit:** 2 MiB.
- **Response:**
  - `204 No Content` — if post update succeeds. Returns empty body.
  - `400 Bad Request` — if request body, parameters, post ID or language is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target post ID not found. Returns a plain-text message.

### `PUT /cms/posts/views/{id}`

Set target post views. Requires special permissions. Notice that creator may not override views on their own posts

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
  - `Content-Type` — `application/json`
- **Path parameters:**
  - `id` — target post, positive integer value
- **Request body:** Omitted/null fields are ignored at update.

```json
  {
    "views": 31243
  }
```

- **Request limit:** 2 MiB.
- **Response:**
  - `204 No Content` — if update succeeds. Returns empty body.
  - `400 Bad Request` — if request body, parameters or post ID is invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — if target post ID not found. Returns a plain-text message.

### `DELETE /cms/posts/{id}`

Deletes target post ID. Note that this deletes all of the post's translations. Creator may always delete his own posts.

- **Headers:**
  - `Authorization` — `Bearer <sessionID>` (required).
- **Path parameters:**
  - `id` — target post ID, positive integer value
- **Request body:** None.
- **Request limit:** 2 MiB.
- **Response:**
  - `204 No Content` — if post deleted successfully. Returns empty body.
  - `400 Bad Request` — if post ID is missing or invalid. Returns a plain-text message.
  - `401 Unauthorized` — if session is invalid, missing or expired identity or if session has insufficient permissions. Returns a plain-text message.
  - `404 Not Found` — target post ID was not found. Returns a plain-text message.
