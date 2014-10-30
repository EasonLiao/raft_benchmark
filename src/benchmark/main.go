package main

import (
  "fmt"
  "github.com/goraft/raft"
)

func main() {
  raft.SetLogLevel(raft.Debug)
  fmt.Println("Hello World")
}
