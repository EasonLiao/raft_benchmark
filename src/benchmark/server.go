package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "log"
  "github.com/goraft/raft"
  "github.com/gorilla/mux"
  "net/http"
)

type Server struct {
  name        string
  host        string
  port        int
  path        string
  router      *mux.Router
  httpServer  *http.Server
}

func New(path string, host string, port int) *Server {
  s := &Server {
        host:   host,
        port:   port,
        path:   path,
        router: mux.NewRouter(),
  }
  fmt.Println("NewServer : ", s.connectionString())
  return s
}

func (s *Server) connectionString() string {
  return fmt.Sprintf("http://%s:%d", s.host, s.port)
}

func (s *Server) Run(leader string) error {
  log.Println("Initialize benchmark server: %s", s.path)

  if leader != "" {
    log.Println("Attempting to join leader:", leader)
    if err := s.join(leader); err != nil {
      log.Fatal(err)
    }
  }

  log.Println("Initialize http server.")
  // Initialize and start HTTP server.
  s.httpServer = &http.Server{
    Addr:    fmt.Sprintf(":%d", s.port),
    Handler: s.router,
  }
  s.router.HandleFunc("/", s.readHandler).Methods("GET")
  s.router.HandleFunc("/join", s.joinHandler).Methods("POST")
  log.Println("Listening at:", s.connectionString())
  return s.httpServer.ListenAndServe()
}

func (s *Server) join(leader string) error {
  command := &raft.DefaultJoinCommand{
    //Name:             s.raftServer.Name(),
    Name:             s.connectionString(),
    ConnectionString: s.connectionString(),
  }
  var b bytes.Buffer
  json.NewEncoder(&b).Encode(command)
  resp, err := http.Post(fmt.Sprintf("http://%s/join", leader), "application/json", &b)
  if err != nil {
    return err
  }
  resp.Body.Close()
  return nil
}

func (s *Server) readHandler(w http.ResponseWriter, req *http.Request) {
  fmt.Println("read")
}

func (s *Server) joinHandler(w http.ResponseWriter, req *http.Request) {
  fmt.Println("join")
}
