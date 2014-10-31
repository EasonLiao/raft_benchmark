package main

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
