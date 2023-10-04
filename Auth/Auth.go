package Auth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"server/DB"
	"server/Models"
	"server/Utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var SESS = make(map[string]Session)            // session map [apiKey] = Session
var LINU = make(map[primitive.ObjectID]string) // logged-in users [userID] = apiKey
var TOKEN_FIELD string = "token"
var USER_FIELD string = "user"

type Session struct {
	USERID  primitive.ObjectID
	APIKEY  string
	LASTREQ time.Time
}

func IsSessionExist(apiKey string) (bool, Session) {
	session, ok := SESS[apiKey]
	if ok {
		return true, session
	}
	return false, Session{}
}

func IsLoggedIn(userId primitive.ObjectID) (bool, Session) {
	apiKey, ok := LINU[userId]
	if ok {
		return true, SESS[apiKey]
	}
	return false, Session{}
}

func IsSessionExpired(session Session) bool {
	if time.Now().Sub(session.LASTREQ).Minutes() > 480 || !session.GetStatus() {
		return true
	}
	return false
}

func CreateSession(userId primitive.ObjectID, apiKey string, employee Models.User) Session {
	SESS[apiKey] = Session{
		USERID:  userId,
		APIKEY:  apiKey,
		LASTREQ: time.Now(),
	}
	LINU[userId] = apiKey
	return SESS[apiKey]
}

func UpdateSessionLastReqTime(session Session) Session {
	SESS[session.APIKEY] = Session{
		USERID:  session.USERID,
		APIKEY:  session.APIKEY,
		LASTREQ: time.Now(),
	}
	return SESS[session.APIKEY]
}

func DeleteSession(apiKey string) {
	session, ok := SESS[apiKey]
	if ok {
		delete(SESS, session.APIKEY)
		delete(LINU, session.USERID)
	}
}

func GetAuthID(c *fiber.Ctx) primitive.ObjectID {
	userAPIKey := string(c.Request().Header.Peek(TOKEN_FIELD))
	session := SESS[userAPIKey]
	return session.USERID
}

type LoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var self LoginInfo
	c.BodyParser(&self)
	self.Password = Utils.HashPassword(self.Password)
	res, err := Utils.FindByFilterProjected(DB.Collections.User,
		bson.M{"username": self.Username, "password": self.Password},
		bson.M{"password": 0})
	if err != nil {
		return c.Status(500).SendString("Internal Server Error")
	}
	if len(res) == 0 {
		return c.Status(401).SendString("incorrect credentials")
	}
	var user Models.User
	bsonBytes, _ := bson.Marshal(res[0])
	bson.Unmarshal(bsonBytes, &user)
	if !user.Status {
		return c.Status(401).SendString("Your status is inactive! Ask admin for permission")
	}
	// check if user is already logged in
	isloggedIn, session := IsLoggedIn(user.ID)
	if IsSessionExpired(session) {
		DeleteSession(session.APIKEY)
		isloggedIn = false
	}
	if !isloggedIn {
		current := time.Now()
		apiString := user.Password + current.String()
		apiSum256 := sha256.Sum256([]byte(apiString))
		apiHash := fmt.Sprintf("%X", apiSum256)
		session = CreateSession(user.ID, apiHash, user)
	} else {
		session = UpdateSessionLastReqTime(session)
	}
	return c.Status(200).JSON(fiber.Map{
		TOKEN_FIELD: session.APIKEY,
		USER_FIELD:  res[0],
	})
}

func Logout(c *fiber.Ctx) error {
	userAPIKey := string(c.Request().Header.Peek(TOKEN_FIELD))
	DeleteSession(userAPIKey)
	return c.Status(200).SendString("logged out")
}

func User(c *fiber.Ctx) Session {
	userAPIKey := string(c.Request().Header.Peek(TOKEN_FIELD))
	session := SESS[userAPIKey]
	return session
}

func SeedAdmin() error {
	collection := DB.Collections.User
	result, err := Utils.FindByFilterProjected(collection, bson.M{"role": Models.ROLE_ADMIN}, bson.M{"_id": 1})
	if err != nil {
		return err
	}
	if len(result) > 0 {
		return nil
	}
	var admin Models.User = Models.User{
		Name:     "Admin",
		Username: "admin",
		Email:    "admin@admin.com",
		Password: Utils.HashPassword("admin"),
		Status:   true,
		Role:     Models.ROLE_ADMIN,
	}
	_, err = collection.InsertOne(context.Background(), admin)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) GetRole() string {
	user, err := Utils.FindByFilterProjected(DB.Collections.User, bson.M{"_id": s.USERID}, bson.M{"role": 1})
	if err != nil {
		return ""
	}
	if len(user) == 0 {
		return ""
	}
	var userObj Models.User
	bsonBytes, _ := bson.Marshal(user[0])
	bson.Unmarshal(bsonBytes, &userObj)
	return userObj.Role
}

func (s *Session) GetStatus() bool {
	user, err := Utils.FindByFilterProjected(DB.Collections.User, bson.M{"_id": s.USERID}, bson.M{"status": 1})
	if err != nil {
		return false
	}
	if len(user) == 0 {
		return false
	}
	var userObj Models.User
	bsonBytes, _ := bson.Marshal(user[0])
	bson.Unmarshal(bsonBytes, &userObj)
	return userObj.Status
}

func GetSessionByID(ID primitive.ObjectID) Session {
	return SESS[LINU[ID]]
}

func GetRoleByID(ID primitive.ObjectID) string {
	user, err := Utils.FindByFilterProjected(DB.Collections.User, bson.M{"_id": ID}, bson.M{"role": 1})
	if err != nil {
		return ""
	}
	if len(user) == 0 {
		return ""
	}
	var userObj Models.User
	bsonBytes, _ := bson.Marshal(user[0])
	bson.Unmarshal(bsonBytes, &userObj)
	return userObj.Role
}

func IsAdmin(ID primitive.ObjectID) bool {
	role := GetRoleByID(ID)
	if role == Models.ROLE_ADMIN {
		return true
	}
	return false
}
