package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"diektronics.com/carter/dl/cfg"
	dlpb "diektronics.com/carter/dl/protos/dl"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	downPrefix    = "/down"
	hookPrefix    = "/hook"
	staticContent = "src/diektronics.com/carter/dl/frontend/web/server/static"
)

type Server struct {
	port    int
	backend string
}

func New(c *cfg.Configuration) *Server {
	return &Server{
		port:    c.HTTPPort,
		backend: fmt.Sprintf("localhost:%v", c.BackendPort),
	}
}

func (s *Server) Run() {
	// 1. Register paths
	r := mux.NewRouter()
	s1 := r.PathPrefix(downPrefix).Subrouter()
	s1.HandleFunc("/", errorHandler(s.listDowns)).Methods("GET")
	s1.HandleFunc("/", errorHandler(s.newDown)).Methods("POST")
	s1.HandleFunc("/{id}", errorHandler(s.getDown)).Methods("GET")
	s1.HandleFunc("/{id}", errorHandler(s.letDown)).Methods("DELETE")

	s2 := r.PathPrefix(hookPrefix).Subrouter()
	s2.HandleFunc("/", errorHandler(s.listHooks)).Methods("GET")

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(staticContent))))

	http.Handle("/", r)
	// 2. Run server
	go http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
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
	conn, err := grpc.Dial(s.backend)
	if err != nil {
		log.Println("dialing:", err)
		return err
	}
	defer conn.Close()
	client := dlpb.NewDlClient(conn)
	reply, err := client.GetAll(context.Background(), &dlpb.GetAllRequest{})
	if err != nil {
		return err
	}
	res := struct{ Downs []*dlpb.Down }{reply.Downs}
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
	for h, ok := range req.Hooks {
		if ok {
			hooks = append(hooks, h)
		}
	}
	links := []*dlpb.Link{}
	for _, url := range strings.Split(req.Links, "\n") {
		url = strings.TrimSpace(url)
		if len(url) > 0 {
			l := &dlpb.Link{Url: url}
			links = append(links, l)
		}
	}
	down := &dlpb.Down{Name: req.Name, Posthook: hooks, Links: links}
	if len(down.Name) == 0 || len(down.Links) == 0 {
		return badRequest{errors.New("please provide a name and links to download")}
	}
	conn, err := grpc.Dial(s.backend)
	if err != nil {
		log.Println("dialing:", err)
		return err
	}
	defer conn.Close()
	client := dlpb.NewDlClient(conn)
	if _, err := client.Download(context.Background(), &dlpb.DownloadRequest{Down: down}); err != nil {
		return err
	}

	return nil
}

func (s *Server) getDown(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return badRequest{err}
	}
	conn, err := grpc.Dial(s.backend)
	if err != nil {
		log.Println("dialing:", err)
		return err
	}
	defer conn.Close()
	client := dlpb.NewDlClient(conn)
	reply, err := client.Get(context.Background(), &dlpb.GetRequest{Id: id})
	if err != nil {
		return notFound{}
	}
	return json.NewEncoder(w).Encode(reply.Down)
}

func (s *Server) letDown(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return badRequest{err}
	}
	conn, err := grpc.Dial(s.backend)
	if err != nil {
		log.Println("dialing:", err)
		return err
	}
	defer conn.Close()
	client := dlpb.NewDlClient(conn)
	reply, err := client.Get(context.Background(), &dlpb.GetRequest{Id: id})
	if err != nil {
		return notFound{}
	}
	if _, err := client.Del(context.Background(), &dlpb.DelRequest{Down: reply.Down}); err != nil {
		return err
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
	conn, err := grpc.Dial(s.backend)
	if err != nil {
		log.Println("dialing:", err)
		return err
	}
	defer conn.Close()
	client := dlpb.NewDlClient(conn)
	reply, err := client.HookNames(context.Background(), &dlpb.HookNamesRequest{})
	if err != nil {
		return err
	}
	res := struct{ Hooks []string }{reply.Names}
	return json.NewEncoder(w).Encode(res)
}
