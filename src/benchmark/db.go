package main


// The key-value database.
type DB struct {
  data  map[string]string
}

// Creates a new database.
func NewDB() *DB {
  return &DB{
    data: make(map[string]string),
  }
}

// Retrieves the value for a given key.
func (db *DB) Get(key string) string {
  return db.data[key]
}

// Sets the value for a given key.
func (db *DB) Put(key string, value string) {
  db.data[key] = value
}
