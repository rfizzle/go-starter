package entity

type Entity interface {
	ID() string                           // ID returns the ID of the entity
	Type() string                         // Type returns the type of the entity
	Role() string                         // Role returns the role of the entity
	Permissions() []string                // Permissions returns the permissions of the entity
	IsType(entityType string) bool        // IsType returns true if the entity is of the given type
	HasRole(role string) bool             // HasRole returns true if the entity has the given role
	HasPermission(permission string) bool // HasPermission returns true if the entity has the given permission
}

type AuthEntity struct {
	id          string
	authType    string
	role        string
	permissions []string
}

func NewAuthEntity(id string, opts ...Option) *AuthEntity {
	return &AuthEntity{
		id: id,
	}
}

type Option func(a *AuthEntity)

func WithType(authType string) Option {
	return func(a *AuthEntity) {
		a.authType = authType
	}
}

func WithRole(role string) Option {
	return func(a *AuthEntity) {
		a.role = role
	}
}

func WithPermissions(permissions []string) Option {
	return func(a *AuthEntity) {
		a.permissions = permissions
	}
}

func (e *AuthEntity) ID() string {
	return e.id
}

func (e *AuthEntity) Type() string {
	return e.authType
}

func (e *AuthEntity) Role() string {
	return e.role
}

func (e *AuthEntity) Permissions() []string {
	return e.permissions
}

func (e *AuthEntity) IsType(entityType string) bool {
	return e.authType == entityType
}

func (e *AuthEntity) HasRole(role string) bool {
	return e.role == role
}

func (e *AuthEntity) HasPermission(permission string) bool {
	for _, p := range e.permissions {
		if p == permission {
			return true
		}
	}
	return false
}
