package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

var (
	listStorageRe   = regexp.MustCompile(`^\/users[\/]*$`)
	getStorageRe    = regexp.MustCompile(`^\/storageget\/(\d+)$`)
	createStorageRe = regexp.MustCompile(`^\/users[\/]*$`)
)

type Storage struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Datastore struct {
	M map[string]Storage
	*sync.RWMutex
}

type StorageHandler struct {
	Store *Datastore
}

func (h *StorageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	header := r.Header.Values("x-api-key")
	fmt.Printf("jeff your header values for x-api-key are %v\n", header)
	switch {
	case r.Method == http.MethodGet && listStorageRe.MatchString(r.URL.Path):
		h.List(w, r)
		return
	case r.Method == http.MethodGet && getStorageRe.MatchString(r.URL.Path):
		h.Get(w, r)
		return
	case r.Method == http.MethodPost && createStorageRe.MatchString(r.URL.Path):
		h.Create(w, r)
		return
	default:
		notFound(w, r)
		return
	}
}

func (h *StorageHandler) List(w http.ResponseWriter, r *http.Request) {
	h.Store.RLock()
	users := make([]Storage, 0, len(h.Store.M))
	for _, v := range h.Store.M {
		users = append(users, v)
	}
	h.Store.RUnlock()
	jsonBytes, err := json.Marshal(users)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func ListHandler(h *StorageHandler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		h.Store.RLock()
		users := make([]Storage, 0, len(h.Store.M))
		for _, v := range h.Store.M {
			users = append(users, v)
		}
		h.Store.RUnlock()
		jsonBytes, err := json.Marshal(users)
		if err != nil {
			internalServerError(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	}
	return http.HandlerFunc(fn)
}
func TimeHandler(format string) http.Handler {
	fmt.Println("in timehandler")
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("in timehandler closure")
		tm := time.Now().Format(format)
		w.Write([]byte("the time is " + tm))
	}
	return http.HandlerFunc(fn)
}

func GetHandler(id string, h *StorageHandler) http.Handler {
	fmt.Println("hi jeff from GetHandler")
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hi jeff from GetHandler closure")
		matches := getStorageRe.FindStringSubmatch(r.URL.Path)
		if len(matches) < 2 {
			notFound(w, r)
			return
		}
		h.Store.RLock()
		u, ok := h.Store.M[matches[1]]
		h.Store.RUnlock()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("user not found"))
			return
		}
		jsonBytes, err := json.Marshal(u)
		if err != nil {
			internalServerError(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	}
	return http.HandlerFunc(fn)
}

func (h *StorageHandler) Get(w http.ResponseWriter, r *http.Request) {
	matches := getStorageRe.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		notFound(w, r)
		return
	}
	h.Store.RLock()
	u, ok := h.Store.M[matches[1]]
	h.Store.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
		return
	}
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *StorageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var u Storage
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		internalServerError(w, r)
		return
	}
	h.Store.Lock()
	h.Store.M[u.ID] = u
	h.Store.Unlock()
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal server error"))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found here"))
}
