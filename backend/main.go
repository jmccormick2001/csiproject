package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"example.com/csiproject/backend/api"
	"example.com/csiproject/backend/db"
)

func main() {
	// Allow user to specify listen port on command line
	var port int
	flag.IntVar(&port, "port", 10000, "port to listen on")
	flag.Parse()

	// Create in-memory database and add a couple of test volumes
	db := db.NewMemoryDatabase()
	db.AddVolume(api.Volume{ID: "1", Name: "volume-1", Hostport: "192.168.0.107:4400", Size: "1G"})
	db.AddVolume(api.Volume{ID: "2", Name: "volume-2", Hostport: "192.168.0.107:5400", Size: "2G"})

	// Create server and wire up database
	server := NewServer(db, log.Default())

	log.Printf("listening on http://localhost:%d", port)
	http.ListenAndServe(":"+strconv.Itoa(port), server)
}

// Server is the volume HTTP server.
type Server struct {
	db  db.Database
	log *log.Logger
}

const (
	ErrorAlreadyExists    = "already-exists"
	ErrorDatabase         = "database"
	ErrorInternal         = "internal"
	ErrorMalformedJSON    = "malformed-json"
	ErrorMethodNotAllowed = "method-not-allowed"
	ErrorNotFound         = "not-found"
	ErrorValidation       = "validation"
)

// NewServer creates a new server using the given database implementation.
func NewServer(db db.Database, log *log.Logger) *Server {
	return &Server{db: db, log: log}
}

// Regex to match "/volumes/:id" (id must be one or more non-slash chars).
var reVolumesID = regexp.MustCompile(`^/volumes/([^/]+)$`)

// ServeHTTP routes the request and calls the correct handler based on the URL
// and HTTP method. It writes a 404 Not Found if the request URL is unknown,
// or 405 Method Not Allowed if the request method is invalid.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	s.log.Printf("%s %s", r.Method, path)

	var id string

	switch {
	case path == "/volumes":
		switch r.Method {
		case "GET":
			s.getVolumes(w, r)
		case "POST":
			s.addVolume(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			s.jsonError(w, http.StatusMethodNotAllowed, ErrorMethodNotAllowed, nil)
		}

	case match(path, reVolumesID, &id):
		switch r.Method {
		case "GET":
			s.getVolumeByID(w, r, id)
		case "DELETE":
			s.deleteVolumeByID(w, r, id)
		default:
			w.Header().Set("Allow", "GET")
			s.jsonError(w, http.StatusMethodNotAllowed, ErrorMethodNotAllowed, nil)
		}

	default:
		s.jsonError(w, http.StatusNotFound, ErrorNotFound, nil)
	}
}

// match returns true if path matches the regex pattern, and binds any
// capturing groups in pattern to the vars.
func match(path string, pattern *regexp.Regexp, vars ...*string) bool {
	matches := pattern.FindStringSubmatch(path)
	if len(matches) <= 0 {
		return false
	}
	for i, match := range matches[1:] {
		*vars[i] = match
	}
	return true
}

func (s *Server) getVolumes(w http.ResponseWriter, r *http.Request) {
	volumes, err := s.db.GetVolumes()
	if err != nil {
		s.log.Printf("error fetching volumes: %v", err)
		s.jsonError(w, http.StatusInternalServerError, ErrorDatabase, nil)
		return
	}
	s.writeJSON(w, http.StatusOK, volumes)
}

func (s *Server) addVolume(w http.ResponseWriter, r *http.Request) {
	var volume api.Volume
	if !s.readJSON(w, r, &volume) {
		return
	}

	// Validate the input and build a map of validation issues
	type validationIssue struct {
		Error   string `json:"error"`
		Message string `json:"message,omitempty"`
	}
	issues := make(map[string]interface{})
	if volume.ID == "" {
		issues["id"] = validationIssue{"required", ""}
	}
	if volume.Hostport == "" {
		issues["hostport"] = validationIssue{"required", ""}
	}
	if volume.Size == "" {
		issues["size"] = validationIssue{"required", ""}
	}
	if volume.Name == "" {
		issues["name"] = validationIssue{"required", ""}
	}
	if len(issues) > 0 {
		s.jsonError(w, http.StatusBadRequest, ErrorValidation, issues)
		return
	}

	err := s.db.AddVolume(volume)
	if errors.Is(err, db.ErrAlreadyExists) {
		s.jsonError(w, http.StatusConflict, ErrorAlreadyExists, nil)
		return
	} else if err != nil {
		s.log.Printf("error adding volume ID %q: %v", volume.ID, err)
		s.jsonError(w, http.StatusInternalServerError, ErrorDatabase, nil)
		return
	}

	s.writeJSON(w, http.StatusCreated, volume)
}

func (s *Server) getVolumeByID(w http.ResponseWriter, r *http.Request, id string) {
	volume, err := s.db.GetVolumeByID(id)
	if errors.Is(err, db.ErrDoesNotExist) {
		s.jsonError(w, http.StatusNotFound, ErrorNotFound, nil)
		return
	} else if err != nil {
		s.log.Printf("error fetching volume ID %q: %v", id, err)
		s.jsonError(w, http.StatusInternalServerError, ErrorDatabase, nil)
		return
	}
	s.writeJSON(w, http.StatusOK, volume)
}
func (s *Server) deleteVolumeByID(w http.ResponseWriter, r *http.Request, id string) {
	deleteResponse, err := s.db.DeleteVolumeByID(id)
	if errors.Is(err, db.ErrDoesNotExist) {
		s.jsonError(w, http.StatusNotFound, ErrorNotFound, nil)
		return
	} else if err != nil {
		s.log.Printf("error fetching volume ID %q: %v", id, err)
		s.jsonError(w, http.StatusInternalServerError, ErrorDatabase, nil)
		return
	}
	s.writeJSON(w, http.StatusOK, deleteResponse)
}

// writeJSON marshals v to JSON and writes it to the response, handling
// errors as appropriate. It also sets the Content-Type header to
// "application/json".
func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		s.log.Printf("error marshaling JSON: %v", err)
		http.Error(w, `{"error":"`+ErrorInternal+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	_, err = w.Write(b)
	if err != nil {
		// Very unlikely to happen, but log any error (not much more we can do)
		s.log.Printf("error writing JSON: %v", err)
	}
}

// jsonError writes a structured error as JSON to the response, with
// optional structured data in the "data" field.
func (s *Server) jsonError(w http.ResponseWriter, status int, error string, data map[string]interface{}) {
	response := struct {
		Status int                    `json:"status"`
		Error  string                 `json:"error"`
		Data   map[string]interface{} `json:"data,omitempty"`
	}{
		Status: status,
		Error:  error,
		Data:   data,
	}
	s.writeJSON(w, status, response)
}

// readJSON reads the request body and unmarshals it from JSON, handling
// errors as appropriate. It returns true on success; the caller should
// return from the handler early if it returns false.
func (s *Server) readJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Printf("error reading JSON body: %v", err)
		s.jsonError(w, http.StatusInternalServerError, ErrorInternal, nil)
		return false
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		data := map[string]interface{}{"message": err.Error()}
		s.jsonError(w, http.StatusBadRequest, ErrorMalformedJSON, data)
		return false
	}
	return true
}
