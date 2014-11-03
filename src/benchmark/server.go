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
  numTxns     int
  txnSize     int
  numPeers    int
  chStart     chan string
  showInterv  int
  snapshot    int64
  stateMemory int64
}

func New(path, host string, port, numTxns, txnSize, numPeers,
         showInterv int, snapshot, stateMemory int64) *Server {
  s := &Server {
        host:         host,
        port:         port,
        path:         path,
        router:       mux.NewRouter(),
        db:           NewDB(),
        numTxns:      numTxns,
        txnSize:      txnSize,
        numPeers:     numPeers,
        chStart:      make(chan string),
        showInterv:   showInterv,
        snapshot:     snapshot,
        stateMemory:  stateMemory,
  }
  s.db.server = s
  return s
}

func (s *Server) connectionString() string {
  return fmt.Sprintf("http://%s:%d", s.host, s.port)
}

func (s *Server) Run(leader string) error {
  var err error
  var isLeader bool = false
  s.name = fmt.Sprintf("%07x", rand.Int())[0:7]
  log.Printf("Initialize benchmark server: %s", s.name)
  if err = ioutil.WriteFile(filepath.Join(s.path, "name"), []byte(s.name), 0644); err != nil {
      panic(err)
  }
  // Initialize and start Raft server.
  transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
  s.raftServer, err = raft.NewServer(s.name, s.path, transporter, s.db, s.db, "")
  //s.raftServer.SetHeartbeatInterval(500 * time.Millisecond)

  if err != nil {
    log.Fatal(err)
  }
  transporter.Install(s.raftServer, s)
  s.raftServer.Start()

  if leader != "" {
    log.Println("Attempting to join leader:", leader)

    if !s.raftServer.IsLogEmpty() {
      log.Fatal("Cannot join with an existing log")
    }

    if err := s.join(leader); err != nil {
      log.Fatal(err)
      panic(err)
    }
  } else if s.raftServer.IsLogEmpty() {
    log.Println("Initializing new cluster : ", s.raftServer.Name(), s.connectionString())
    _, err := s.raftServer.Do(&raft.DefaultJoinCommand{
      Name:             s.raftServer.Name(),
      ConnectionString: s.connectionString(),
    })
    if err != nil {
      log.Fatal(err)
    }
    isLeader = true
  } else {
    log.Println("Recovered from log")
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

  s.initStateMachine()

  // Pre-fill the state machine to make it has specified memory usage.
  if isLeader {
    go s.runBenchmark()
  }
  return s.httpServer.ListenAndServe()
}

func (s *Server) join(leader string) error {
  command := &raft.DefaultJoinCommand{
    Name:             s.raftServer.Name(),
    ConnectionString: s.connectionString(),
  }
  var b bytes.Buffer
  json.NewEncoder(&b).Encode(command)
  fmt.Println("BEF JOIN!!")
  resp, err := http.Post(fmt.Sprintf("http://%s/join", leader), "application/json", &b)
  fmt.Println("AFT JOIN!!")
  if err != nil {
    return err
  }
  resp.Body.Close()
  fmt.Println("SUC!!")
  return nil
}

func (s *Server) readHandler(w http.ResponseWriter, req *http.Request) {
}

func (s *Server) joinHandler(w http.ResponseWriter, req *http.Request) {
  command := &raft.DefaultJoinCommand{}
  if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if _, err := s.raftServer.Do(command); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  numMembers := s.raftServer.MemberCount()
  if numMembers == s.numPeers {
    // Gets enough servers join in.
    s.chStart <- "start"
  }
  log.Printf("New server joined in, now cluster has %d servers.", numMembers)
}

func (s* Server) runBenchmark() {
  if s.numPeers > 1 {
    log.Printf("Waits for cluster size changes to %d", s.numPeers)
    // Waits for start message.
    <-s.chStart
  }

  ticker := time.NewTicker(time.Second * time.Duration(s.showInterv))
  go s.showPerf(ticker)

  log.Println("Starts benchmark:")
  // Execute the command against the Raft server.
  st := time.Now()

  for i:=0; i < 1000; i++ {
    go s.doPuts()
  }
  s.doPuts()

  duration := float32(time.Since(st)) / 1000000000
  fmt.Printf("Duration : %f, throughput : %f\n", duration, float32(s.numTxns) / duration)
  ticker.Stop()
}

func (s* Server) doPuts() {
  for i:=0; i < s.numTxns; i++ {
    _, err := s.raftServer.Do(NewPutCommand(i, string(make([]byte, s.txnSize, s.txnSize))))
    if err != nil {
      fmt.Println("Error in raft", err)
    }
  }
}

// This is a hack around Gorilla mux not providing the correct net/http
// HandleFunc() interface.
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
  s.router.HandleFunc(pattern, handler)
}

func (s *Server) asyncPuts(num int) {

}

func (s *Server) showPerf(ticker *time.Ticker) {
  lastPuts := s.db.puts
  lastDelays :=  s.db.delays
  for {
    <-ticker.C
    curPuts := s.db.puts
    curDelays := s.db.delays
    diff := curPuts - lastPuts
    intvThroughput := diff / s.showInterv
    if diff == 0 {
      diff = 1
    }
    fmt.Printf("throughput %d, delay %d ms\n", intvThroughput,
               int(curDelays - lastDelays) / diff)
    lastPuts = curPuts
    lastDelays = curDelays
  }
}

func (s *Server) initStateMachine() {
  numKeys := int(s.stateMemory / int64(s.txnSize))
  s.db.fill(numKeys, s.txnSize)
}
