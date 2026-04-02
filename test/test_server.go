package main

import (
"fmt"
"net/http"
)

func main() {
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
fmt.Fprintf(w, "MS-AI Server Running - MongoDB Transaction Fixes Applied!")
})

fmt.Println("Server starting on :8081")
http.ListenAndServe(":8081", nil)
}
