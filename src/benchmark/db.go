package main

import (
  "bytes"
  "encoding/gob"
  "log"
)

// The key-value database.
type DB struct {
  data  map[int]string
  puts  int
}

// Creates a new database.
func NewDB() *DB {
  return &DB{
    data: make(map[int]string),
  }
}

// Retrieves the value for a given key.
func (db *DB) Get(key int) string {
  return db.data[key]
}

// Sets the value for a given key.
func (db *DB) Put(key int, value string) {
  db.data[key] = value
  db.puts++
}

func (db *DB) Save() ([]byte, error) {
  b := new(bytes.Buffer)
  e := gob.NewEncoder(b)
  // Encoding the map
  err := e.Encode(db.data)
  if err != nil {
    panic(err)
  }
  log.Println("Return Snapshot!")
  return b.Bytes(), nil
}

func (db *DB) Recovery([]byte) error {
  return nil
}
