package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"diektronics.com/carter/dl/dl"
	"diektronics.com/carter/dl/hook"
	"diektronics.com/carter/dl/types"
	"github.com/gorilla/mux"
)

const DownPrefix = "/down"
const HookPrefix = "/hook"

type Server struct {
	d *dl.Downloader
}

func New(d *dl.Downloader) *Server {
	return &Server{d: d}
}

func (s *Server) Run() {
	// 1. Register paths
	r := mux.NewRouter()
	s1 := r.PathPrefix(DownPrefix).Subrouter()
	s1.HandleFunc("/", errorHandler(s.listDowns)).Methods("GET")
	s1.HandleFunc("/", errorHandler(s.newDown)).Methods("POST")
	s1.HandleFunc("/{id}", errorHandler(s.getDown)).Methods("GET")
	s1.HandleFunc("/{id}", errorHandler(s.letDown)).Methods("DELETE")

	s2 := r.PathPrefix(HookPrefix).Subrouter()
	s2.HandleFunc("/", errorHandler(s.listHooks)).Methods("GET")

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("src/diektronics.com/carter/dl/server/static"))))

	http.Handle("/", r)
	// 2. Run server
	go http.ListenAndServe(":4444", nil)
}

// badRequest is handled by setting the status code in the reply to StatusBadRequest.
type badRequest struct{ error }

// notFound is handled by setting the status code in the reply to StatusNotFound.
type notFound struct{ error }

// errorHandler wraps a function returning an error by handling the error and returning a http.Handler.
// If the error is of the one of the types defined above, it is handled as described for every type.
// If the error is of another type, it is considered as an internal error and its message is logged.
func errorHandler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}
		switch err.(type) {
		case badRequest:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case notFound:
			http.Error(w, "task not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "oops", http.StatusInternalServerError)
		}
	}
}

func (s *Server) listDowns(w http.ResponseWriter, r *http.Request) error {
	downs, err := s.d.Db().GetAll(nil)
	if err != nil {
		return err
	}
	res := struct{ Downs []*types.Download }{downs}
	return json.NewEncoder(w).Encode(res)
}

func (s *Server) newDown(w http.ResponseWriter, r *http.Request) error {
	req := struct {
		Name  string
		Links string
		Hooks map[string]bool
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return badRequest{err}
	}
	log.Println(req)

	hooks := []string{}
	for h := range req.Hooks {
		hooks = append(hooks, h)
	}
	links := []*types.Link{}
	for _, url := range strings.Split(req.Links, "\n") {
		l := &types.Link{URL: strings.TrimSpace(url)}
		links = append(links, l)
	}
	// FIXME(diek): UNRAR hook has to happen BEFORE REMOVE hook. sort inversely...
	sort.Sort(sort.Reverse(sort.StringSlice(hooks)))
	down := &types.Download{Name: req.Name, Posthook: strings.Join(hooks, ","), Links: links}
	if len(down.Name) == 0 || len(down.Links) == 0 {
		return badRequest{errors.New("please provide a name and links to download")}
	}
	if err := s.d.Download(down); err != nil {
		return err
	}

	return nil
}

func (s *Server) getDown(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return badRequest{err}
	}
	down, err := s.d.Db().Get(id)
	if err != nil {
		return notFound{}
	}
	return json.NewEncoder(w).Encode(down)
}

func (s *Server) letDown(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return badRequest{err}
	}
	down, err := s.d.Db().Get(id)
	if err != nil {
		return notFound{}
	}
	if err := s.d.Db().Del(down); err != nil {
		return notFound{}
	}
	return nil
}

func parseID(r *http.Request) (int64, error) {
	txt, ok := mux.Vars(r)["id"]
	if !ok {
		return 0, fmt.Errorf("down id not found")
	}
	return strconv.ParseInt(txt, 10, 0)
}

func (s *Server) listHooks(w http.ResponseWriter, r *http.Request) error {
	res := struct{ Hooks []string }{hook.Names()}
	return json.NewEncoder(w).Encode(res)
}
