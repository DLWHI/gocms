package domain

var actions = map[string]uint32{
	"delete": 0x04,
	"edit":   0x02,
	"create": 0x01,
}

var entites = map[string]uint32{
	"posts":      0,
	"categories": 3,
	"ads":        6,
	"views":      9,
	"users":      12,
	"editors":    15,
	"locales":    18,
}

const (
	RoleEditor     = uint32(0x00000019) // posts:create, categories:create;edit
	RoleTranslator = uint32(0x00000033) // posts:create, posts:edit, categories;edit
	RoleLeadEditor = uint32(0x001F8FFB) // posts:create;edit, categories:all, ads:all, views:all, editors:all, locales:all
	RoleAdmin      = uint32(0x001FF3F6) // posts:edit;delete, categories:all, ads:all, users:all, editors:all, locales:all
	RoleRoot       = uint32(0xFFFFFFFF) // everything possible, even future
)

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
