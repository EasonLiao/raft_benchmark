package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "log"
  "math/rand"
  "github.com/goraft/raft"
  "github.com/gorilla/mux"
  "net/http"
  "time"
  "io/ioutil"
  "path/filepath"
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
  leader      bool
}

func New(path string, host string, port int) *Server {
  s := &Server {
        host:   host,
        port:   port,
        path:   path,
        router: mux.NewRouter(),
        db:     NewDB(),
  }
  return s
}

func (s *Server) connectionString() string {
  return fmt.Sprintf("http://%s:%d", s.host, s.port)
}

func (s *Server) Run(leader string) error {
  var err error
  s.name = fmt.Sprintf("%07x", rand.Int())[0:7]
  log.Println("Initialize benchmark server: %s", s.name)
  if err = ioutil.WriteFile(filepath.Join(s.path, "name"), []byte(s.name), 0644); err != nil {
      panic(err)
  }

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
      panic(err)
    }
    s.leader = false
  } else {
    s.leader = true
    log.Println("Initializing new cluster : ", s.raftServer.Name(), s.connectionString())
    _, err := s.raftServer.Do(&raft.DefaultJoinCommand{
      Name:             s.raftServer.Name(),
      ConnectionString: s.connectionString(),
    })
    if err != nil {
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

  res := s.httpServer.ListenAndServe()
  time.Sleep(5000 * time.Millisecond)
  return res
}

func (s *Server) join(leader string) error {
  command := &raft.DefaultJoinCommand{
    Name:             s.raftServer.Name(),
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
  fmt.Println("join", req)
  command := &raft.DefaultJoinCommand{}

  if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if _, err := s.raftServer.Do(command); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  fmt.Println("Join suc!")
}

func (s* Server) runBenchmark() {
  if s.leader == false {
    return
  }
  return
  time.Sleep(10000 * time.Millisecond)
  fmt.Println("Run benchmark!")
  // Execute the command against the Raft server.
  stt := time.Now()
  for i:=0; i < 10000; i++ {
    _, err := s.raftServer.Do(NewWriteCommand(string("1"), string("2")))
    if err != nil {
      fmt.Println("Error in raft", err)
    }
  }
  fmt.Println("Duration : ", time.Since(stt))
}

// This is a hack around Gorilla mux not providing the correct net/http
// HandleFunc() interface.
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
  s.router.HandleFunc(pattern, handler)
}
