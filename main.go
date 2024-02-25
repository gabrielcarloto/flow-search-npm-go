package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"github.com/pkg/browser"
)

type GetPackagesResponse struct {
	Results []struct {
		Package struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Links       struct {
				Npm string `json:"npm"`
			} `json:"links"`
		} `json:"package"`
	} `json:"results"`
}

type JsonRPCAction struct {
	Method string   `json:"method,omitempty"`
	Params []string `json:"parameters,omitempty"`
}

type Result struct {
	Title         string `json:"Title"`
	Subtitle      string `json:"Subtitle,omitempty"`
	JsonRPCAction `json:"JsonRPCAction"`
	IcoPath       string `json:"IcoPath"`
}

type JsonRPCResponse struct {
	Result []Result `json:"result"`
}

const (
	Query = "query"
	Open  = "open"
)

const API_URL = "https://api.npms.io/v2/search"

func main() {
	jsonrpc := os.Args[len(os.Args)-1]

	var request JsonRPCAction
	err := json.Unmarshal([]byte(jsonrpc), &request)
	check(err, "Error parsing request")

	switch request.Method {
	case Query:
		methodQuery(request.Params[0])
  case Open:
    methodOpen(request.Params[0])
	}
}

func methodQuery(query string) {
	if query == "" {
		sendResult(Result{Title: "Waiting for query...", Subtitle: "Hello from go!"})
    return
	}

	packages, err := queryPackage(query)
	check(err, "Error querying package")

	results := mapPackagesToResults(packages)
  sendResults(results)
}

func methodOpen(url string) {
  browser.OpenURL(url)
}

func queryPackage(query string) (*GetPackagesResponse, error) {
	res, err := http.Get(API_URL + "?q=" + query)
	check(err, "Error querying packages")

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("The API responded with status code %v", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	check(err, "Error reading API response body")

	var packages GetPackagesResponse
	err = json.Unmarshal(resBody, &packages)
	check(err, "Error parsing API response body")

	return &packages, nil
}

func mapPackagesToResults(packs *GetPackagesResponse) []Result {
	var mapped = make([]Result, len(packs.Results))
	for i, v := range packs.Results {
		var action JsonRPCAction
		action.Method = Open
		action.Params = append(action.Params, v.Package.Links.Npm)

		mapped[i] = Result{
			Title:         v.Package.Name + " | v" + v.Package.Version,
			Subtitle:      v.Package.Description,
			JsonRPCAction: action,
      IcoPath: "app.png",
    }
	}

  return mapped
}

func sendResult(result Result) {
	var r = []Result{result}
	sendResults(r)
}

func sendResults(results []Result) {
	str, err := json.Marshal(JsonRPCResponse{Result: results})
	check(err, "Error sending results")
	fmt.Println(string(str))
}

func check(err error, msg string) {
	if err != nil {
		fmt.Println(msg + ": " + err.Error())
		os.Exit(1)
	}
}
