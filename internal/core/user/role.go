// ----- START OF FILE: backend/MS-AI/internal/core/user/role.go -----
// internal/core/user/role.go
package user

// Role represents a user role as a typed string
type Role string

// Role constants
const (
	RoleUser      Role = "user"      // Regular user, granted automatically on signup
	RoleAdmin     Role = "admin"     // System administrator, has full access
	RoleModerator Role = "moderator" // Content moderator, has content management access
)

// Roles represents a collection of user roles
type Roles []Role

// Has checks if the roles collection contains the given role
func (r Roles) Has(role Role) bool {
	for _, existing := range r {
		if existing == role {
			return true
		}
	}
	return false
}

// IsAdmin checks if the roles collection contains admin role
func (r Roles) IsAdmin() bool {
	return r.Has(RoleAdmin)
}

// IsModerator checks if the roles collection contains moderator role
func (r Roles) IsModerator() bool {
	return r.Has(RoleModerator)
}

// ToStrings converts Roles to []string for storage/compatibility
func (r Roles) ToStrings() []string {
	strings := make([]string, len(r))
	for i, role := range r {
		strings[i] = string(role)
	}
	return strings
}

// FromStrings creates Roles from []string (used when reading from storage)
func FromStrings(strings []string) Roles {
	roles := make(Roles, len(strings))
	for i, s := range strings {
		roles[i] = Role(s)
	}
	return roles
}

// ----- END OF FILE: backend/MS-AI/internal/core/user/role.go -----
