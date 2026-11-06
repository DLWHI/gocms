package domain

import "fmt"

const (
	ActionCreate = "create"
	ActionEdit   = "edit"
	ActionDelete = "delete"
)

var actions = map[string]uint32{
	ActionDelete: 0x04,
	ActionEdit:   0x02,
	ActionCreate: 0x01,
}

var entites = map[string]uint32{
	EntityName(&Post{}):          0,
	EntityName(&Category{}):      3,
	EntityName(&Advertisement{}): 6,
	"views":                      9,
	EntityName(&User{}):          12,
	EntityName(&Editor{}):        15,
	"languages":                  18,
}

const (
	RoleEditor     = uint32(0x00000019) // posts:create, categories:create;edit
	RoleTranslator = uint32(0x00000033) // posts:create, posts:edit, categories;edit
	RoleLeadEditor = uint32(0x001F8FFB) // posts:create;edit, categories:all, ads:all, views:all, editors:all, languages:all
	RoleAdmin      = uint32(0x001FF3F6) // posts:edit;delete, categories:all, ads:all, users:all, editors:all, languages:all
	RoleRoot       = uint32(0xFFFFFFFF) // everything possible, even future
)

func EntityName(entity any) string {
	switch entity.(type) {
	case Advertisement, *Advertisement:
		return "ads"
	case Category, *Category:
		return "categories"
	case Editor, *Editor:
		return "editors"
	case Post, *Post:
		return "posts"
	case User, *User:
		return "users"
	default:
		panic(fmt.Errorf("%w: unregistered entity requested", ErrInval))
	}
}

func GetEntities() map[string]uint32 {
	return entites
}

func GetActions() map[string]uint32 {
	return actions
}

func HasPermission(permissions uint32, entity string, action string) bool {
	if offset, exists := entites[entity]; exists {
		mask := actions[action]
		if mask != 0 {
			return (mask & (permissions >> offset)) == mask
		}
	}
	return false
}

func PermissionsOnEntity(entity string, acts ...string) uint32 {
	if offset, exists := entites[entity]; exists {
		var mask uint32
		for _, action := range acts {
			if val, ok := actions[action]; ok {
				mask |= val
			}
		}
		return mask << offset
	}
	return 0
}
