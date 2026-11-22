# Admin API

These routes cover all required actions to control CMS state, such as administrating users, creating and editing posts, categories and etc.

## Authentication

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/profile` | [Get user profile](auth.md#get-cmsprofile) |
| GET | `/cms/permissions` | [Get list of defined permissions](auth.md#get-cmspermissions) |
| POST | `/cms/login` | [Authenticate user](auth.md#post-cmslogin) |
| POST | `/cms/logout` | [Logout from CMS](auth.md#post-cmslogout) |
| PATCH | `/cms/password` | [Change user's password](auth.md#patch-cmspassword) |

## User management

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/users` | [List active users](users.md#get-cmsusers) |
| GET | `/cms/users/{username}` | [Get a user](users.md#get-cmsusersusername) |
| POST | `/cms/users` | [Create a user](users.md#post-cmsusers) |
| POST | `/cms/users/search` | [Search users](users.md#post-cmsuserssearch) |
| PATCH | `/cms/users/{username}/email` | [Change email](users.md#patch-cmsusersusernameemail) |
| PATCH | `/cms/users/{username}/password` | [Reset a password](users.md#patch-cmsusersusernamepassword) |
| PATCH | `/cms/users/{username}/permissions` | [Change permissions](users.md#patch-cmsusersusernamepermissions) |
| DELETE | `/cms/users/{username}` | [Suspend a user](users.md#delete-cmsusersusername) |

## Editor management

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/editors/{lang}` | [List editors](editors.md#get-cmseditorslang) |
| GET | `/cms/editors/{lang}/{slug}` | [Get editor](editors.md#get-cmseditorslangslug) |
| POST | `/cms/editors/{lang}` | [Create editor](editors.md#post-cmseditorslang) |
| POST | `/cms/editors/{lang}/search` | [Search editors](editors.md#post-cmseditorslangsearch) |
| PATCH | `/cms/editors/{lang}/{slug}` | [Update editor](editors.md#patch-cmseditorslangslug) |
| DELETE | `/cms/editors/{slug}` | [Delete editor](editors.md#delete-cmseditorsslug) |

## Language management

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/languages` | [List languages](languages.md#get-cmslanguages) |
| POST | `/cms/languages` | [Create a language](languages.md#post-cmslanguages) |
| DELETE | `/cms/languages/{code}` | [Delete a language](languages.md#delete-cmslanguagescode) |

## Post management

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/posts/{lang}/{id}` | [Get a post translation](posts.md#get-cmspostslangid) |
| POST | `/cms/posts/{lang}` | [Create a post translation](posts.md#post-cmspostslang) |
| POST | `/cms/posts/{lang}/search` | [Search posts](posts.md#post-cmspostslangsearch) |
| PATCH | `/cms/posts/{lang}/{id}` | [Update a post translation](posts.md#patch-cmspostslangid) |
| PUT | `/cms/posts/views/{id}` | [Set post views](posts.md#put-cmspostsviewsid) |
| DELETE | `/cms/posts/{id}` | [Delete a post and its translations](posts.md#delete-cmspostsid) |

## Category management

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/categories` | [List categories](categories.md#get-cmscategories) |
| GET | `/cms/categories/{slug}` | [Get a category](categories.md#get-cmscategoriesslug) |
| POST | `/cms/categories` | [Create a category](categories.md#post-cmscategories) |
| POST | `/cms/categories/search` | [Search categories](categories.md#post-cmscategoriessearch) |
| PATCH | `/cms/categories/{slug}` | [Update a category](categories.md#patch-cmscategoriesslug) |
| DELETE | `/cms/categories/{slug}` | [Delete a category](categories.md#delete-cmscategoriesslug) |

## Advertisement management

| Method | Route | Reference |
| --- | --- | --- |
| GET | `/cms/ads/{id}` | [Get an advertisement](advertisements.md#get-cmsadsid) |
| GET | `/cms/ads/groups` | [List advertisement groups](advertisements.md#get-cmsadsgroups) |
| POST | `/cms/ads` | [Create an advertisement](advertisements.md#post-cmsads) |
| POST | `/cms/ads/search` | [Search advertisements](advertisements.md#post-cmsadssearch) |
| PATCH | `/cms/ads/{id}` | [Update an advertisement](advertisements.md#patch-cmsadsid) |
| DELETE | `/cms/ads/{id}` | [Delete an advertisement](advertisements.md#delete-cmsadsid) |
