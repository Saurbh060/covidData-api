package main

import (
	"net/http"
	"fmt"
	"encoding/json"
    "io/ioutil"
    "os"
	"time"
	"github.com/labstack/echo/v4"
	// "go.mongodb.org/mongo-driver/bson"

	"context"
    "log"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoFields struct {
	State string `bson:"state"`
	TotalCases float64 `bson:"totalCases"`
}

type StateData struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	StateCases []MongoFields `bson:"stateCases" json:"stateCases"`
}

func Connect( mongoFiled StateData) {
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

	// data := StateData{
	// 	StateCases: []MongoFields{
	// 		{
	// 			State:"AN",
	// 			TotalCases : 23124,
	// 		},
	// 		{
	// 			State:"AN",
	// 			TotalCases : 23124,
	// 		},
	// 	},
	// }
	// fmt.Println("printing data", data)
	insertManyResult, err := stateCollection.InsertOne(ctx,mongoFiled)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted multiple documents: ", insertManyResult.InsertedID)
	fmt.Println("Connection to MongoDB closed.")
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

	docs:= []MongoFields{} 
	doc:= StateData{} 
	
	for key := range result {		
		n := MongoFields{State: key, TotalCases: result[key].(map[string]interface{})["total"].(map[string]interface{})["confirmed"].(float64)}
        docs = append(docs,n)
	}
	doc.StateCases = docs
	fmt.Println(docs)

	Connect(doc )
	return c.JSON(http.StatusOK, "success")
}

func getStateName(c echo.Context) error {
	// User ID from path `users/:id`
	state := c.Param("state")
  return c.String(http.StatusOK, state)
}

var FILE_PATH = "./output-stats.json"


func main() {
	e := echo.New()
	e.GET("/cases/:state", getStateName)
	e.GET("/saveCases", saveCovidData)
	e.Logger.Fatal(e.Start(":1323"))
}

