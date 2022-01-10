package main

import "net/http"

func main() {
  http.HandleFunc("/", index)

  http.ListenAndServe(":8090", nil)
}
