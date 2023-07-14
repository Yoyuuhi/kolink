package main

import (
	"fmt"
	"os"
	"yoyuuhi/kolink/graph"
	"yoyuuhi/kolink/request"
	"yoyuuhi/kolink/state"
)

func main() {
	requestDefs, err := request.GetRequestDefs()
	if err != nil {
		fmt.Printf("error %s", err)
		os.Exit(1)
	}

	repositoryName := requestDefs.RepositoryName
	for _, requestDef := range requestDefs.RequestDefs {
		callerFileStateMap, err := state.GenerateStateMap(repositoryName, requestDef)
		if err != nil {
			fmt.Printf("error %s", err)
			os.Exit(1)
		}
		if err := graph.DrawGraphs(callerFileStateMap, requestDef); err != nil {
			fmt.Println("Failed to draw graph:", err)
			continue
		}
	}
}
