package main

import (
  "github.com/goraft/raft"
)

// This command writes a value to a key.
type PutCommand struct {
  Key     int `json:"key"`
  Value   string `json:"value"`
  TimeSt  int64 `json: "timest"`
}

// Creates a new write command.
func NewPutCommand(key int, value string) *PutCommand {
  return &PutCommand{
    Key:    key,
    Value:  value,
    TimeSt: GetTimeMs(),
  }
}

// The name of the command in the log.
func (c *PutCommand) CommandName() string {
  return "write"
}

// Writes a value to a key.
func (c *PutCommand) Apply(server raft.Server) (interface{}, error) {
  db := server.Context().(*DB)
  db.Put(c.Key, c.Value, c.TimeSt)
  return nil, nil
}
