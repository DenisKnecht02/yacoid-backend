package constants

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (role *Role) String() string {
	return string(*role)
}

func (role *Role) StringAddress() *string {
	str := string(*role)
	return &str
}
