package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// die Namen werden spaeter nach die Finalprompt benannt
type Data struct {
	HttpCode int    `json:"http_status_code"`
	Message  string `json:"status_message"`
	IssueTag string `json:"issue_tag"`
}

type Root struct {
	Status     Data `json:"status"`
	Validation Data `json:"validation"`
}

// Array-Dummies for storing the data before access to database
var ValidRequests_correctData []Data
var ValidRequests_incorrectData []Data
var InvalidRequests []Data

func jsonReader(validationFile string) {
	// >>read JSON data from a file
	// >>the file contains a list of Validation objects
	// >>each object has an HttpCode, Message, and IssueTag
	fileContents, err := os.ReadFile(validationFile)
	if err != nil {
		panic(err)
	}
	// >>In the future, replace this file reading logic with code that reads user input from the CLI,
	// >>for example using bufio.Scanner or the flag package to accept input from the user.
	var data Data
	err = json.Unmarshal(fileContents, &data)
	if err != nil {
		panic(err)
	}

	var root Root
	err = json.Unmarshal(fileContents, &root)
	if err != nil {
		panic(err)
	}

	// validate the data

	validate(root.Status.HttpCode, root.Status.Message, root.Validation.IssueTag)

	fmt.Println("")
	fmt.Println("Valid correct data:", ValidRequests_correctData)
	fmt.Println("Valid incorrect data:", ValidRequests_incorrectData)
	fmt.Println("Invalid data:", InvalidRequests)

	// >>save the data in the database
	//saveInDatabase()

}

func validate(httpCode int, message string, issueTag string) {

	// read JSON data and divide it into 3 categories
	//1. Valid correct data
	//2. Valid incorrect data
	//3. Invalid data
	// Output a corresponding message for each category

	if httpCode == 200 {

		if issueTag == "" {
			ValidRequests_correctData = append(ValidRequests_correctData, Data{HttpCode: httpCode, Message: message, IssueTag: issueTag})
			fmt.Println("Valid correct data:", message)
		} else {
			ValidRequests_incorrectData = append(ValidRequests_incorrectData, Data{HttpCode: httpCode, Message: message, IssueTag: issueTag})
			fmt.Println("Valid incorrect data:", message, issueTag)
		}
	} else {
		InvalidRequests = append(InvalidRequests, Data{HttpCode: httpCode, Message: message, IssueTag: issueTag})
		fmt.Println("Invalid data:", message)
	}
}

func saveInDatabase() {
	// save the data in the database

}
