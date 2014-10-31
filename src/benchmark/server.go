package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "log"
  "github.com/goraft/raft"
  "github.com/gorilla/mux"
  "net/http"
  "time"
)

type Server struct {
  name        string
  host        string
  port        int
  path        string
  router      *mux.Router
  httpServer  *http.Server
  raftServer  raft.Server
  db          *DB
}

func New(path string, host string, port int) *Server {
  s := &Server {
        host:   host,
        port:   port,
        path:   path,
        router: mux.NewRouter(),
        db:     NewDB(),
  }
  fmt.Println("NewServer : ", s.connectionString())
  return s
}

func (s *Server) connectionString() string {
  return fmt.Sprintf("http://%s:%d", s.host, s.port)
}

func (s *Server) Run(leader string) error {
  var err error
  log.Println("Initialize benchmark server: %s", s.path)

  // Initialize and start Raft server.
  transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
  s.raftServer, err = raft.NewServer(s.name, s.path, transporter, nil, s.db, "")
  if err != nil {
    log.Fatal(err)
  }
  transporter.Install(s.raftServer, s)
  s.raftServer.Start()

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

  go s.runBenchmark()

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

func (s* Server) runBenchmark() {
  fmt.Println("Run benchmark!!!")
  value := string("12345")
  // Execute the command against the Raft server.
  s.raftServer.Do(NewPutCommand(0, value))
}

// This is a hack around Gorilla mux not providing the correct net/http
// HandleFunc() interface.
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
  s.router.HandleFunc(pattern, handler)
}
