package test

import (
	"gocms/internal/search"

	"github.com/meilisearch/meilisearch-go"
)

type Meilisearch struct {
	search.MeilisearchClient

	UsersIndex      meilisearch.IndexManager
	PostsIndex      meilisearch.IndexManager
	CategoriesIndex meilisearch.IndexManager
	EditorsIndex    meilisearch.IndexManager
	AdsIndex        meilisearch.IndexManager
}

func NewMeilisearch(
	client meilisearch.ServiceManager,
	usersIndex string,
	postsIndex string,
	categoriesIndex string,
	editorsIndex string,
	adsIndex string,
) *Meilisearch {
	return &Meilisearch{
		MeilisearchClient: search.MeilisearchClient{
			Client: client,
		},
		UsersIndex:      client.Index(usersIndex),
		PostsIndex:      client.Index(postsIndex),
		CategoriesIndex: client.Index(categoriesIndex),
		EditorsIndex:    client.Index(editorsIndex),
		AdsIndex:        client.Index(adsIndex),
	}
}
