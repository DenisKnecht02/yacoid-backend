package constants

import (
	"strings"
)

type Role string

type roleList struct {
	Unknown   Role
	Admin     Role
	Moderator Role
	User      Role
}

var EnumRole = &roleList{
	Unknown:   "unknown",
	Admin:     "admin",
	Moderator: "moderator",
	User:      "user",
}

var roleMap = map[string]Role{
	"unknown":   EnumRole.Unknown,
	"admin":     EnumRole.Admin,
	"moderator": EnumRole.Moderator,
	"user":      EnumRole.User,
}

func ParseStringToRole(str string) (Role, error) {
	role, ok := roleMap[strings.ToLower(str)]
	if ok {
		return role, nil
	} else {
		return role, ErrorInvalidEnum
	}
}

func (role Role) String() string {
	switch role {
	case EnumRole.Admin:
		return "admin"
	case EnumRole.Moderator:
		return "moderator"
	case EnumRole.User:
		return "user"
	}
	return "unknown"
}

func (role *Role) StringAddress() *string {
	str := role.String()
	return &str
}

func RoleArrayToStringAdressArray(roles []Role) []*string {

	stringArray := []*string{}
	for _, role := range roles {
		stringArray = append(stringArray, role.StringAddress())
	}

	return stringArray

}

func StringArrayToRoleArray(stringRoles []string) (*[]Role, error) {

	roles := []Role{}
	for _, stringRole := range stringRoles {
		role, err := ParseStringToRole(stringRole)

		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return &roles, nil

}
