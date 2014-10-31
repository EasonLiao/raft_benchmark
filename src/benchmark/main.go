package main

import (
  "flag"
  "fmt"
  "github.com/goraft/raft"
  "log"
  "os"
  "math/rand"
  "time"
)

var host string
var port int
var join string
var numTxns int
var txnSize int
var numPeers int
var showInterval int
var snapshot int64
var stateMemory int64

func init() {
  flag.StringVar(&host, "h", "localhost", "hostname")
  flag.StringVar(&join, "join", "", "host:port of leader to join")
  flag.IntVar(&port, "p", 4001, "port")
  flag.IntVar(&numTxns, "n", 100000, "number of transactions")
  flag.IntVar(&txnSize, "s", 128, "transaction size(bytes)")
  flag.IntVar(&numPeers, "np", 1, "number of peers in cluster")
  flag.IntVar(&showInterval, "int", 3, "the intervals for showing the perf")
  flag.Int64Var(&snapshot, "snapshot", -1, "the threshold for taking snapshot")
  flag.Int64Var(&stateMemory, "memory", 100000000, "the memory size of state machine")
}

func main() {
  flag.Parse()

  // Go uses pseudo random, need to init seed.
  rand.Seed(time.Now().UnixNano())

  path := flag.Arg(0)
  fmt.Println(path)
  if err := os.MkdirAll(path, 0744); err != nil {
    log.Fatalf("Unable to create path: %v", err)
  }
  raft.RegisterCommand(&PutCommand{})
  server := New(path, host, port, numTxns, txnSize, numPeers, showInterval,
                snapshot, stateMemory)
  server.Run(join)
}
