package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
)

type savedValues struct {
	StartedDate string  `json:"startedDate"`
	EndedDate   string  `json:"endedDate"`
	Tagged      string  `json:"tagged"`
	Euros       float64 `json:"euros"`
}

var cardTraderBearer = "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJjYXJkdHJhZGVyLXByb2R1Y3Rpb24iLCJzdWIiOiJhcHA6OTc0NiIsImF1ZCI6ImFwcDo5NzQ2IiwiZXhwIjo0ODY3NjMzNzI0LCJqdGkiOiI3ZjI4YjljYS0wNTlkLTQ2MDMtODcwNC02OTQzMjNkMjJmMTIiLCJpYXQiOjE3MTE5NjAxMjQsIm5hbWUiOiJQYW5hc2F2diBBcHAgMjAyNDAzMzExNDQyMzIifQ.Kp7Qmoe7tilLPhDM66QMGHMLowLkgqs-gARsR2OPZk9ryQ4boTwTxfzDDOV-f9Tnzg3pK1XqGi6M9Ua13qCC7vFdSIDzYRVeS4Wt0rk28NGJA-M3EUkz6Pn2PqWsIHX5YL33u_X_EVAVWM68eJwTRhnVopdl7-oB57FVRs1lmovICEFtH0zgA_PqIk_RMhEoIpdzJUjSJPXKrtiLLCFhWadtSg1WVBGmHntjMGhjKweWx6jm5RRoLR8beNMCtC1qP2fOB0tC7vdSO_lcfu9T79tS2N6z3gStJSJSoSJWmvQ2GVkDsF2rvNeKNdXyWtPcH5mTFkS7_aSDS3lDvgjmWg"

func main() {
	// Check if start and end dates are provided as command-line arguments
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <start_date> <end_date> <tag>")
		return
	}

	// Parse start and end dates from command-line arguments
	startDate := os.Args[1]
	endDate := os.Args[2]
	tag := os.Args[3]
	if tag == "owner" {
		tag = ""
	} else if tag == "past" {
		tag = "past sleeves"
	}

	// Construct the URL with start and end dates
	url := fmt.Sprintf("https://api.cardtrader.com/api/v2/orders?sort=date.desc&from=%s&to=%s", startDate, endDate)

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Add the Authorization header with the JWT token
	req.Header.Add("Authorization", "Bearer "+cardTraderBearer)

	// Send the request
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer response.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Check if the response body is empty
	if string(responseBody) == "[]" {
		fmt.Println("Wrong Dates")
		return
	}

	// Print the response body
	//fmt.Println(string(responseBody))
	jsonName := fmt.Sprintf("order(%s---%s).json", startDate, endDate)
	// You can also store the response in a file if needed
	err = os.WriteFile(jsonName, responseBody, 0644)
	if err != nil {
		fmt.Println("Error writing response to file:", err)
		return
	}
	fmt.Println("Response saved to " + jsonName)

	total, err := ProcessOrders(jsonName, tag)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	totalPostFee := (total - (total * 0.07)) / 100
	fmt.Println("Total Euros", total/100)
	fmt.Println("Total Euros after the 7 percent fee of cardtrader", roundFloat(totalPostFee, 2))
	savedValuesPastSleeves := savedValues{
		StartedDate: startDate,
		EndedDate:   endDate,
		Tagged:      tag,
		Euros:       roundFloat(totalPostFee, 2),
	}

	if totalPostFee == float64(0) {
		fmt.Println("Either the seller with tag " + tag + " didn't sell anything or wrong tag")
	}

	// Convert the struct to JSON format
	jsonData, err := json.Marshal(savedValuesPastSleeves)
	if err != nil {
		fmt.Println("Error encoding struct to JSON:", err)
		return
	}
	if tag == "" {
		tag = "owner"
	}
	// Open the file for appending
	txtFileName := fmt.Sprintf("./%s.txt", tag)
	file, err := os.OpenFile(txtFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Append the JSON data to the file
	if _, err := file.WriteString("\n" + string(jsonData)); err != nil {
		fmt.Println("Error appending to file:", err)
		return
	}

	fmt.Println("Data appended to txt file successfully.")
}

// Function to process orders with buyer ID 34089
func ProcessOrders(filename string, tagGiven string) (float64, error) {
	// Read the JSON file
	jsonFile, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("error reading JSON file: %v", err)
	}

	// Declare a slice to store the JSON data
	var orders []map[string]interface{}

	// Unmarshal the JSON data into the slice
	if err := json.Unmarshal(jsonFile, &orders); err != nil {
		return 0, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// Variable to store total cents for past sleeves
	pastSleevesTotal := float64(0)

	// Iterate over each order
	for _, order := range orders {
		// Check if the buyer ID is 34089
		if buyer, ok := order["buyer"].(map[string]interface{}); ok {
			if buyerID, ok := buyer["id"].(float64); ok && buyerID == 34089 {
				// Iterate over order items
				if orderItems, ok := order["order_items"].([]interface{}); ok {
					for _, item := range orderItems {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if tag, ok := itemMap["tag"].(string); ok && tag == tagGiven {
								if sellerPrice, ok := itemMap["seller_price"].(map[string]interface{}); ok {
									if cents, ok := sellerPrice["cents"].(float64); ok {
										if quantity, ok := itemMap["quantity"].(float64); ok {
											pastSleevesTotal += float64(cents * quantity)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Return the total cents for past sleeves
	return pastSleevesTotal, nil
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
