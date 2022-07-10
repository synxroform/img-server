package main

import (
  "os"
//  "io"
  "fmt"
  "log"
//  "time"
//  "math"
//  "strings"
//  "errors"
  "net/http"
//  "encoding/json"
  bolt "go.etcd.io/bbolt" 
  "github.com/go-chi/chi/v5"
)


type MappingJSON struct {
  ProductTag string `json:"product_tag"`
  ImageURL string `json:"image_url"`
}

var users_db *bolt.DB
var images_db *bolt.DB

func fatal_assert(e error, msg string) {
  if e != nil {
    log.Fatal(e)
    panic(msg)
  }
}


func init() {
  var e error
  users_db, e = bolt.Open("users.db", 0600, nil)
  fatal_assert(e, "failed to open users.db")
  
  images_db, e = bolt.Open("images.db", 0600, nil)
  fatal_assert(e, "failed to open images.db")
    
  defer users_db.Close()
  defer images_db.Close()
}


func BodyGuard(next http.Handler) http.Handler {
  return http.HandlerFunc( func (w http.ResponseWriter, r *http.Request) {
      if r.Body == http.NoBody {
        http.Error(w, "empty body", http.StatusBadRequest)
        return
      } else {
        next.ServeHTTP(w, r)
      }
  })
}


func main() {
  R := chi.NewRouter()
  
  R.Group(func (r chi.Router) {
    r.Use(BodyGuard)
    r.Use(AuthorizationGuard)
    r.Post("/user/add", UserAdd)
   // r.Post("/mapping/add", MappingAdd)
  })
  
  http.Handle("/", R)
  fmt.Printf("running img-server on port %s\n", os.Args[1])
  http.ListenAndServe(os.Args[1], nil)
}
