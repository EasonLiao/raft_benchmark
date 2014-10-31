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

func init() {
  flag.StringVar(&host, "h", "localhost", "hostname")
  flag.StringVar(&join, "join", "", "host:port of leader to join")
  flag.IntVar(&port, "p", 4001, "port")
  flag.IntVar(&numTxns, "n", 100000, "number of transactions")
  flag.IntVar(&txnSize, "s", 128, "transaction size(bytes)")
}

func main() {
  flag.Parse()
  raft.SetLogLevel(raft.Debug)

  // Go uses pseudo random, need to init seed.
  rand.Seed(time.Now().UnixNano())

  path := flag.Arg(0)
  fmt.Println(path)
  if err := os.MkdirAll(path, 0744); err != nil {
    log.Fatalf("Unable to create path: %v", err)
  }
  raft.RegisterCommand(&PutCommand{})
  server := New(path, host, port, numTxns, txnSize)
  server.Run(join)
}
