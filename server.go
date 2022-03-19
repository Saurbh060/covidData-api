package main

import (
	"net/http"
	"fmt"
	
	"encoding/json"
    "io/ioutil"
    "os"
	"github.com/labstack/echo/v4"
)

func parseMap(aMap map[string]interface{}) {
    for key, val := range aMap {
		// fmt.Println("------------------in map------------------")
        switch concreteVal := val.(type) {
        case map[string]interface{}:
            // fmt.Println(key)
			if key=="total" {
				fmt.Println(key, ":", concreteVal["confirmed"])
			}
            parseMap(val.(map[string]interface{}))
        case []interface{}:
            // fmt.Println(key)
            parseArray(val.([]interface{}))
        default:
			if key=="total" {
				fmt.Println(key, ":", concreteVal)
			}
        }
		
    }
}

func parseArray(anArray []interface{}) {
	fmt.Println("------------------in array------------------")
    for i, val := range anArray {
        switch concreteVal := val.(type) {
        case map[string]interface{}:
            fmt.Println("Index:", i)
            parseMap(val.(map[string]interface{}))
        case []interface{}:
            fmt.Println("Index:", i)
            parseArray(val.([]interface{}))
        default:
            fmt.Println("Index", i, ":", concreteVal)

        }
    }
}

func saveCovidData(c echo.Context) error {
	// resp, err := http.Get("https://data.covid19india.org/v4/min/data.min.json")
	// if err != nil {
	// 	return err
	// }

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

	var resp = make(map[string]float64)
	// parseMap(result)

	// fmt.Print("\nall states json data\n\n")
    // fmt.Println(result["AN"])


	// json.Unmarshal([]byte(resp), &res)

	// var y [2]string{"English", "Japanese"}

	// dummy+json= { 
	// 	"state" : "key",
	// 	"count": value
	// }

	for key := range result {
		// fmt.Println("------------------in map------------------")
        // concreteVal := val.(type)
		// fmt.Println(key, ":",result[key].(map[string]interface{})["total"].(map[string]interface{})["confirmed"])
		// fmt.Println("------------------in array------------------")
		
		resp[key] = result[key].(map[string]interface{})["total"].(map[string]interface{})["confirmed"].(float64)
		// fmt.Print("value  ",result[key].(map[string]interface{})["total"].(map[string]interface{})["confirmed"].(float64))
    }

	jsonStr, err := json.Marshal(resp)

	if err != nil {
        fmt.Printf("Error: %s", err.Error())
    } else {
        fmt.Println(string(jsonStr))
    }

		

	
	return c.JSON(http.StatusOK, string(jsonStr))
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

