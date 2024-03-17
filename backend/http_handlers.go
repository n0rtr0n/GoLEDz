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

func updatePatternHandler(w http.ResponseWriter, r *http.Request, patterns map[string]Pattern, currentPattern *Pattern) {
	fmt.Println("handling pattern update request")
	patternName := r.PathValue("pattern")
	pattern, ok := patterns[patternName]
	if !ok {
		fmt.Println("error fetching pattern")
		return
	}

	parameters := pattern.GetPatternUpdateRequest()

	err := json.NewDecoder(r.Body).Decode(&parameters)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	err = pattern.UpdateParameters(parameters.GetParameters())
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("new pattern", patternName)
	*currentPattern = pattern
}
