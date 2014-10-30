package main

import (
  "flag"
  "fmt"
  "github.com/goraft/raft"
  "log"
  "os"
)

var host string
var port int
var join string

func init() {
  flag.StringVar(&host, "h", "localhost", "hostname")
  flag.StringVar(&join, "join", "", "host:port of leader to join")
  flag.IntVar(&port, "p", 4001, "port")
}

func main() {
  flag.Parse()
  raft.SetLogLevel(raft.Debug)

  path := flag.Arg(0)
  fmt.Println(path)
  if err := os.MkdirAll(path, 0744); err != nil {
    log.Fatalf("Unable to create path: %v", err)
  }
  server := New(path, host, port)
  server.Run(join)
}
