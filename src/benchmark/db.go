package main

import (
  "bytes"
  "encoding/gob"
  "log"
  "sync"
)

// The key-value database.
type DB struct {
  data    map[int]string
  puts    int
  delays  int64
  distribution map[int]int
  lock    sync.Mutex
}

// Creates a new database.
func NewDB() *DB {
  return &DB{
    data: make(map[int]string),
    distribution: make(map[int]int),
  }
}

// Retrieves the value for a given key.
func (db *DB) Get(key int) string {
  return db.data[key]
}

// Sets the value for a given key.
func (db *DB) Put(key int, value string, timeStamp int64) {
  db.lock.Lock()
  defer db.lock.Unlock()

  db.data[key % len(db.data)] = value
  db.puts++
  delay := GetTimeMs() - timeStamp
  db.delays += delay
  db.distribution[int(delay) % 10] += 1
}

func (db *DB) Save() ([]byte, error) {
  db.lock.Lock()
  defer db.lock.Unlock()

  log.Println("Start Snapshot")
  b := new(bytes.Buffer)
  e := gob.NewEncoder(b)
  // Encoding the map
  err := e.Encode(db.data)
  if err != nil {
    panic(err)
  }
  log.Println("Return Snapshot")
  return b.Bytes(), nil
}

func (db *DB) Recovery([]byte) error {
  return nil
}

// Pre-fill the state machine.
func (db *DB) fill(numKeys int, txnSize int) {
  for i := 0; i < numKeys; i++ {
    value := string(make([]byte, txnSize, txnSize))
    db.data[i] = value
  }
  log.Println("After fill : ", len(db.data))
}
