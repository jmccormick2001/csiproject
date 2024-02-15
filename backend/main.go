package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"example.com/csiproject/backend/api"
	"example.com/csiproject/backend/db"
	"example.com/csiproject/backend/model"
)

func main() {
	// Allow user to specify listen port on command line
	var port int
	flag.IntVar(&port, "port", 10000, "port to listen on")
	flag.Parse()

	// Create in-memory database and add a couple of test volumes
	db := db.NewMemoryDatabase()
	db.AddVolume(model.Volume{ID: "1", Name: "volume-1", Hostport: "192.168.0.107:4400", Size: "1G"})
	db.AddVolume(model.Volume{ID: "2", Name: "volume-2", Hostport: "192.168.0.107:5400", Size: "2G"})

	// Create server and wire up database
	server := api.NewServer(db, log.Default())

	log.Printf("listening on http://localhost:%d", port)
	http.ListenAndServe(":"+strconv.Itoa(port), server)
}
