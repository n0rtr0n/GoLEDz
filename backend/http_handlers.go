package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "GoLEDz web server")
}

func getPatternsHandler(w http.ResponseWriter, r *http.Request, patterns *map[string]Pattern) {
	type AllPatternsRequest struct {
		Patterns Patterns `json:"patterns"`
	}

	patternsReq := AllPatternsRequest{
		Patterns: *patterns,
	}

	jsonData, err := json.Marshal(patternsReq)
	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		return
	}

	fmt.Fprint(w, string(jsonData))
}
