package main

import (
  "time"
)

// Gets current timestamp in milliseconds
func GetTimeMs() int64 {
  return time.Now().UnixNano() / int64(time.Millisecond)
}
