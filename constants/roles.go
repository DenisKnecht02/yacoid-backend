package constants

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleUser      Role = "user"
)

func (role *Role) String() string {
	return string(*role)
}

func (role *Role) StringAddress() *string {
	str := string(*role)
	return &str
}

func RoleArrayToStringAdressArray(roles []Role) []*string {

	stringArray := []*string{}
	for _, role := range roles {
		stringArray = append(stringArray, role.StringAddress())
	}

	return stringArray

}
