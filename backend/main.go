package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"example.com/csiproject/backend/api"
)

func main() {
	mux := http.NewServeMux()
	storageH := &api.StorageHandler{
		Store: &api.Datastore{
			M: map[string]api.Storage{
				"1": {ID: "1", Name: "bob"},
				"2": {ID: "2", Name: "junk"},
			},
			RWMutex: &sync.RWMutex{},
		},
	}
	mux.Handle("/users", storageH)
	mux.Handle("/users/", storageH)
	mux.Handle("/storageget/", api.GetHandler("2", storageH))
	mux.Handle("/storage/", api.ListHandler(storageH))
	mux.Handle("/time/", api.TimeHandler(time.RFC1123))

	fmt.Println("listening...")
	log.Fatal(http.ListenAndServe(":10000", mux))
}
