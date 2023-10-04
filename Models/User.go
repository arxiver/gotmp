package Models

import (
	"server/Utils"

	validation "github.com/go-ozzo/ozzo-validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string             `json:"name,omitempty"`
	Username   string             `json:"username,omitempty"`
	Status     bool               `json:"status"`
	Email      string             `json:"email,omitempty"`
	Password   string             `json:"password,omitempty"`
	Role       string             `json:"role,omitempty"`
	ModifiedAt primitive.DateTime `json:"modifiedat" bson:"modifiedat"`
}
type UserView struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string             `json:"name,omitempty"`
	Username   string             `json:"username,omitempty"`
	Status     bool               `json:"status"`
	Email      string             `json:"email,omitempty"`
	Role       string             `json:"role,omitempty"`
	ModifiedAt primitive.DateTime `json:"modifiedat" bson:"modifiedat"`
}

type UserSearch struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	IDIsUsed       bool               `json:"idisused,omitempty"`
	Name           string             `json:"name,omitempty"`
	NameIsUsed     bool               `json:"nameisused,omitempty"`
	Username       string             `json:"username,omitempty"`
	UsernameIsUsed bool               `json:"usernameisused,omitempty"`
	Status         bool               `json:"status,omitempty"`
	StatusIsUsed   bool               `json:"statusisused,omitempty"`
	Email          string             `json:"email,omitempty"`
	EmailIsUsed    bool               `json:"emailisused,omitempty"`
	Role           string             `json:"role,omitempty"`
	RoleIsUsed     bool               `json:"roleisused,omitempty"`
}

func (obj *UserSearch) ToBSON() bson.M {
	self := bson.M{}
	if obj.IDIsUsed {
		self["_id"] = obj.ID
	}
	if obj.NameIsUsed {
		self["name"] = Utils.RegexBSONSearch(obj.Name)
	}
	if obj.UsernameIsUsed {
		self["username"] = Utils.RegexBSONSearch(obj.Username)
	}
	if obj.StatusIsUsed {
		self["status"] = obj.Status
	}
	if obj.EmailIsUsed {
		self["email"] = Utils.RegexBSONSearch(obj.Email)
	}
	if obj.RoleIsUsed {
		self["role"] = obj.Role
	}
	return self
}

func (obj *User) Validate() error {
	return validation.ValidateStruct(obj,
		validation.Field(&obj.Name, validation.Required),
		validation.Field(&obj.Username, validation.Required),
		validation.Field(&obj.Email, validation.Required),
		validation.Field(&obj.Email, validation.Match(Utils.EmailREGEX)),
		validation.Field(&obj.Role, validation.Required),
		validation.Field(&obj.Role, validSelect(ROLE_VALID)),
		validation.Field(&obj.Password, validation.Length(6, 20)),
	)
}

var validSelect = func(states map[string]bool) *validation.StringRule {
	return validation.NewStringRule(func(str string) bool {
		return states[str]
	}, "Not a valid selection")
}

func (self *User) CloneFromView(other *UserView) {
	self.ID = other.ID
	self.Name = other.Name
	self.Username = other.Username
	self.Status = other.Status
	self.Email = other.Email
	self.Role = other.Role
	self.ModifiedAt = other.ModifiedAt
}

func (other *UserView) ToUser() User {
	var self User
	self.ID = other.ID
	self.Name = other.Name
	self.Username = other.Username
	self.Status = other.Status
	self.Email = other.Email
	self.Role = other.Role
	return self
}

const (
	STATE_ONLINE  = "online"
	STATE_OFFLINE = "offline"
	STATE_BREAK   = "break"
)

var STATE_VALID map[string]bool = map[string]bool{
	STATE_ONLINE:  true,
	STATE_OFFLINE: true,
	STATE_BREAK:   true,
}

const (
	ROLE_ADMIN = "admin"
	ROLE_USER  = "user"
)

var ROLE_VALID map[string]bool = map[string]bool{
	ROLE_ADMIN: true,
	ROLE_USER:  true,
}
