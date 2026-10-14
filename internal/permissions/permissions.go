package permissions

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

func Has(entity string, action string, level uint32) bool {
	if offset, exists := entites[entity]; exists {
		mask := actions[action]
		if mask != 0 {
			return (mask & (level >> offset)) == mask
		}
	}
	return false
}

func OnEntityC(entity string) uint32 {
	if offset, exists := entites[entity]; exists {
		mask := actions["create"]
		return mask << offset
	}
	return 0
}

func OnEntityE(entity string) uint32 {
	if offset, exists := entites[entity]; exists {
		mask := actions["edit"]
		return mask << offset
	}
	return 0
}

func OnEntityD(entity string) uint32 {
	if offset, exists := entites[entity]; exists {
		mask := actions["delete"]
		return mask << offset
	}
	return 0
}

func OnEntityCE(entity string) uint32 {
	if offset, exists := entites[entity]; exists {
		create := actions["create"]
		edit := actions["edit"]
		return (create | edit) << offset
	}
	return 0
}

func OnEntityCD(entity string) uint32 {
	if offset, exists := entites[entity]; exists {
		create := actions["create"]
		delete := actions["delete"]
		return (create | delete) << offset
	}
	return 0
}

func OnEntityDE(entity string) uint32 {
	if offset, exists := entites[entity]; exists {
		edit := actions["edit"]
		delete := actions["delete"]
		return (edit | delete) << offset
	}
	return 0
}

func OnEntityCED(entity string) uint32 {
	if offset, exists := entites[entity]; exists {
		create := actions["create"]
		edit := actions["edit"]
		delete := actions["delete"]
		return (create | edit | delete) << offset
	}
	return 0
}
