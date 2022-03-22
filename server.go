package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/jasonwinn/geocoder"

	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoFields struct {
	State      string  `bson:"state"`
	TotalCases float64 `bson:"totalCases"`
}

type StateData struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	StateCases []MongoFields      `bson:"stateCases" json:"stateCases"`
}

type Location struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

func Connect(mongoFiled StateData) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to MongoDB!")
	defer client.Disconnect(ctx)

	database := client.Database("covidCases")
	stateCollection := database.Collection("stateData")

	fmt.Println("printing mongo data")
	fmt.Println(mongoFiled)

	insertManyResult, err := stateCollection.InsertOne(ctx, mongoFiled)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted multiple documents: ", insertManyResult.InsertedID)
	fmt.Println("Connection to MongoDB closed.")
}

func ConnectAndGet(stateName string) MongoFields {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to MongoDB!")
	defer client.Disconnect(ctx)

	database := client.Database("covidCases")
	stateCollection := database.Collection("stateData")

	var podcast StateData
	if err = stateCollection.FindOne(ctx, bson.M{}).Decode(&podcast); err != nil {
		log.Fatal(err)
	}

	var result MongoFields
	for i := 0; i < len(podcast.StateCases); i++ {
		// fmt.Println("key", podcast.StateCases[i])
		data := podcast.StateCases[i]
		// fmt.Println("state:", data.State)
		if data.State == stateName {
			result.State = data.State
			result.TotalCases = data.TotalCases
			break
		}
	}
	fmt.Println("Connection to MongoDB closed.")
	return result
}

func saveCovidData(c echo.Context) error {

	// Open our jsonFile
	jsonFile, err := os.Open("data.min.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print("\n\nSuccessfully Opened users.json\n\n")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	docs := []MongoFields{}
	doc := StateData{}

	for key := range result {
		n := MongoFields{State: key, TotalCases: result[key].(map[string]interface{})["total"].(map[string]interface{})["confirmed"].(float64)}
		docs = append(docs, n)
	}
	doc.StateCases = docs
	fmt.Println(docs)

	Connect(doc)
	return c.JSON(http.StatusOK, "success")
}

func getStateName(c echo.Context) error {

	var location Location

	defer c.Request().Body.Close()

	b, err := ioutil.ReadAll(c.Request().Body)

	if err != nil {
		log.Printf("Failed reading the request body %s", err)
		return c.String(http.StatusInternalServerError, "")
	}
	err = json.Unmarshal(b, &location)
	if err != nil {
		log.Printf("Failed unmarshal  %s", err)
		return c.String(http.StatusInternalServerError, "")
	}

	fmt.Println(location)
	geocoder.SetAPIKey("Fmjtd%7Cluub256alu%2C7s%3Do5-9u82ur")
	address, err := geocoder.ReverseGeocode(location.Lat, location.Long)
	if err != nil {
		panic("THERE WAS SOME ERROR!!!!!")
	}

	fmt.Println(address.State)

	stateMapping := map[string]string{
		"Andaman and Nicobar Islands": "AN",
		"Andhra Pradesh":              "AP",
		"Arunachal Pradesh":           "AR",
		"Assam":                       "AS",
		"Bihar":                       "BR",
		"Chandigarh":                  "CH",
		"Chhattisgarh":                "CT",
		"Dadra and Nagar Haveli":      "DN",
		"Daman and Diu":               "DD",
		"Delhi":                       "DL",
		"Goa":                         "GA",
		"Gujarat":                     "GJ",
		"Haryana":                     "HR",
		"Himachal Pradesh":            "HP",
		"Jammu and Kashmir":           "JK",
		"Jharkhand":                   "JH",
		"Karnataka":                   "KA",
		"Kerala":                      "KL",
		"Lakshadweep":                 "LD",
		"Madhya Pradesh":              "MP",
		"Maharashtra":                 "MH",
		"Manipur":                     "MN",
		"Meghalaya":                   "ML",
		"Mizoram":                     "MZ",
		"Nagaland":                    "NL",
		"Odisha":                      "OR",
		"Orissa":                      "OD",
		"cherry":                      "PY",
		"Punjab":                      "PB",
		"Rajasthan":                   "RJ",
		"Sikkim":                      "SK",
		"Tamil Nadu":                  "TN",
		"Telangana":                   "TG",
		"Tripura":                     "TR",
		"Uttar Pradesh":               "UP",
		"Uttarakhand":                 "UT",
		"West Bengal":                 "WB",
	}

	if v, found := stateMapping[address.State]; found {

		fmt.Println("state code: ", v)
		fetchedData := ConnectAndGet(v)

		responseData := &MongoFields{
			State:      fetchedData.State,
			TotalCases: fetchedData.TotalCases,
		}
		fmt.Println("response: ", *responseData)
		return c.JSON(http.StatusOK, responseData)
	} else {
		fmt.Println("Coordinates are not belongs to India")
	}

	return c.JSON(http.StatusOK, "Coordinates are not belongs to India")
}

func main() {
	e := echo.New()
	e.POST("/stateCases", getStateName)
	e.GET("/saveCases", saveCovidData)
	e.Logger.Fatal(e.Start(":1323"))
}
