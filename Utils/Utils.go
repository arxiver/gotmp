package Utils

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var EmailREGEX = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,5}$`)
var BaseStaticPath = "./Resources" // used at main.go
var BaseImagesPath = "./Resources/Images/"

func HashPassword(password string) string {
	return fmt.Sprintf("%X", sha256.Sum256([]byte(password)))
}
func CheckIfObjExistingByObjId(collection *mongo.Collection, objID primitive.ObjectID) error {
	filter := bson.M{"_id": objID}

	var results []bson.M
	cur, err := collection.Find(context.Background(), filter)
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	cur.All(context.Background(), &results)
	fmt.Println("Count : ", len(results))

	if len(results) == 0 {
		return errors.New("obj not found")
	}

	return nil
}

func AdaptCurrentTimeByUnit(unit string, period int) time.Time {
	now := time.Now()
	if unit == "Month" {
		now = now.AddDate(0, period, 0)
	} else if unit == "Week" {
		now = now.AddDate(0, 0, period*7)
	} else if unit == "Day" {
		now = now.AddDate(0, 0, period)
	} else if unit == "Year" {
		now = now.AddDate(period, 0, 0)
	}
	return now
}

func AdaptRefernceTimeByUnit(refernceTime time.Time, unit string, period int) time.Time {
	if unit == "Month" {
		refernceTime = refernceTime.AddDate(0, period, 0)
	} else if unit == "Week" {
		refernceTime = refernceTime.AddDate(0, 0, period*7)
	} else if unit == "Day" {
		refernceTime = refernceTime.AddDate(0, 0, period)
	} else if unit == "Year" {
		refernceTime = refernceTime.AddDate(period, 0, 0)
	}
	return refernceTime
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func UploadImage(c *fiber.Ctx) string {
	file, err := c.FormFile("image")
	if err != nil {
		fmt.Println("Failed in saving Image")
		c.Status(500).Send([]byte("Invalid data sent for uploading"))
		return "Error"
	}

	// Save file to root directory
	var filePath = fmt.Sprintf("Resources/Images/img_%d_%d.png", rand.Intn(1024), MakeTimestamp())
	saveing_err := c.SaveFile(file, "./"+filePath)
	if saveing_err != nil {
		c.Status(500).Send([]byte("Failed to save the uploaded image"))
		return "Error"
	} else {
		c.Status(200).Send([]byte("Saved Successfully"))
		return filePath
	}
}

func FindByFilter(collection *mongo.Collection, filter bson.M) (bool, []bson.M) {
	results := []bson.M{}

	cur, err := collection.Find(context.Background(), filter)
	if err != nil {
		return false, results
	}
	defer cur.Close(context.Background())

	cur.All(context.Background(), &results)

	return true, results
}

func Contains(arr []primitive.ObjectID, elem primitive.ObjectID) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
}

func ContainsString(arr []string, elem string) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
}

func Unique(inSlice []primitive.ObjectID) []primitive.ObjectID {
	keys := make(map[string]bool)
	var list []primitive.ObjectID
	for _, entry := range inSlice {
		if _, value := keys[entry.Hex()]; !value {
			keys[entry.Hex()] = true
			list = append(list, entry)
		}
	}
	return list
}

func ArrayStringContains(arr []string, elem string) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
}

func DecodeArrData(inStructArr, outStructArr interface{}) error {
	in := struct{ Data interface{} }{Data: inStructArr}
	inStructArrData, err := bson.Marshal(in)
	if err != nil {
		return err
	}
	var out struct{ Data bson.Raw }
	if err := bson.Unmarshal(inStructArrData, &out); err != nil {
		return err
	}
	return bson.Unmarshal(out.Data, &outStructArr)
}

func SendTextResponseAsJSON(c *fiber.Ctx, msg string) {
	response, _ := json.Marshal(
		bson.M{"result": msg},
	)
	c.Set("Content-Type", "application/json")
	c.Status(200).Send(response)
}

func DateToJulianDay() string {
	yearDay := time.Now().YearDay()
	lastYearTwoDigits := strconv.Itoa(time.Now().Year())[2:4]
	return lastYearTwoDigits + strconv.Itoa(yearDay)
}

func UploadImageBase64(stringBase64, imageDocType, prefix, baseFolderName string) (string, error) {
	i := strings.Index(stringBase64, ",")
	if stringBase64 == "" || imageDocType == "" {
		return "", errors.New("Invalid data sent to be saved")
	}
	file, _ := base64.StdEncoding.DecodeString(stringBase64[i+1:])
	// make sure base path exists
	if _, err := os.Stat(BaseImagesPath); os.IsNotExist(err) {
		os.MkdirAll(BaseImagesPath, 0755)
	}
	// make sure base folder exists
	if _, err := os.Stat(path.Join(BaseImagesPath, baseFolderName)); os.IsNotExist(err) {
		os.MkdirAll(path.Join(BaseImagesPath, baseFolderName), 0755)
	}

	var filePath = path.Join(BaseImagesPath, baseFolderName, fmt.Sprintf(prefix+"_%d_%d.%s", rand.Intn(1024), MakeTimestamp(), imageDocType))

	f, err := os.Create("./" + filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(file); err != nil {
		return "", err
	}
	f.Sync()
	return filePath, nil
}

func GetModifcationBSONObj(obj interface{}, invalidNames []string) bson.M {
	self := bson.M{}
	valueOfObj := reflect.ValueOf(obj)
	typeOfObj := valueOfObj.Type()
	invalidFieldNames := append([]string{"ID"}, invalidNames...)

	for i := 0; i < valueOfObj.NumField(); i++ {
		if ArrayStringContains(invalidFieldNames, typeOfObj.Field(i).Name) {
			continue
		}
		self[strings.ToLower(typeOfObj.Field(i).Name)] = valueOfObj.Field(i).Interface()
	}
	return self
}

func RegexBSONSearch(s string) bson.D {
	regexPattern := fmt.Sprintf(".*%s.*", s)
	return bson.D{{"$regex", primitive.Regex{Pattern: regexPattern, Options: "i"}}}
}

func FindByFilterProjected(collection *mongo.Collection, filter bson.M, fields bson.M) ([]bson.M, error) {
	var results []bson.M
	opts := options.FindOptions{Projection: fields}
	cur, err := collection.Find(context.Background(), filter, &opts)
	if err != nil {
		return results, err
	}
	defer cur.Close(context.Background())

	cur.All(context.Background(), &results)

	return results, err
}

func GetDateTimeNow() primitive.DateTime {
	return primitive.NewDateTimeFromTime(time.Now())
}
