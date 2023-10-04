package Controllers

import (
	"context"
	"server/DB"
	"server/Models"
	"server/Utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserGetAll - Get all users
func UserGetAll(c *fiber.Ctx) error {
	var search Models.UserSearch
	if err := c.BodyParser(&search); err != nil {
		return err
	}
	users, err := userGetAll(&search, false)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return c.JSON([]interface{}{})
	}
	return c.JSON(users)
}

// UserGetById - Get user by id
func UserGetById(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return err
	}
	user, err := userGetById(id, false)
	if err != nil {
		return err
	}
	return c.JSON(user)
}

// UserNew - Create new user
func UserNew(c *fiber.Ctx) error {
	var user Models.User
	if err := c.BodyParser(&user); err != nil {
		return err
	}
	if err := user.Validate(); err != nil {
		return err
	}
	if ok := IsUniqueField("username", user.Username, user.ID); !ok {
		return c.Status(fiber.StatusConflict).SendString("username: already exists")
	}
	user.Password = Utils.HashPassword(user.Password)
	if err := userNew(user); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).SendString("User created")
}

// UserModifyById - Edit user by id
func UserModifyById(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return err
	}
	var user Models.User
	if err := c.BodyParser(&user); err != nil {
		return err
	}
	if ok := IsUniqueField("username", user.Username, id); !ok {
		return c.Status(fiber.StatusConflict).SendString("username: already exists")
	}
	self, err := userGetById(id, true)
	if err != nil {
		return err
	}
	user.ID = id
	if err = user.Validate(); err != nil {
		return err
	}
	if user.Password != "" {
		user.Password = Utils.HashPassword(user.Password)
	} else {
		user.Password = self.Password
	}
	if err := userModify(user); err != nil {
		return err
	}
	return c.JSON(user)
}

func userGetAll(search *Models.UserSearch, safe bool) ([]Models.User, error) {
	var users []Models.User
	var projection = bson.M{}
	if !safe {
		projection = bson.M{"password": 0}
	}
	cursor, err := DB.Collections.User.Find(context.Background(), search.ToBSON(), &options.FindOptions{Projection: projection})
	if err != nil {
		return users, err
	}
	cursor.All(nil, &users)
	return users, err
}

func userGetById(id primitive.ObjectID, safe bool) (Models.User, error) {
	var user Models.User
	err := DB.Collections.User.FindOne(nil, bson.M{"_id": id}).Decode(&user)
	if !safe {
		user.Password = ""
	}
	return user, err
}

func userNew(user Models.User) error {
	_, err := DB.Collections.User.InsertOne(nil, user)
	return err
}

func userModify(user Models.User) error {
	user.ModifiedAt = Utils.GetDateTimeNow()
	_, err := DB.Collections.User.UpdateOne(nil, bson.M{"_id": user.ID},
		bson.M{"$set": Utils.GetModifcationBSONObj(user, []string{})})
	return err
}

func IsUniqueField(field, value string, id primitive.ObjectID) bool {
	var user Models.User
	err := DB.Collections.User.FindOne(nil, bson.M{field: value, "_id": bson.M{"$ne": id}}).Decode(&user)
	return err != nil
}
