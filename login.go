package main

import (
//  "fmt"
  "bytes"
//  "strings"
//  "errors"
  "net/http"
//  "regexp"
  "encoding/json"
  "golang.org/x/crypto/pbkdf2"
  bolt "go.etcd.io/bbolt" 
  "crypto/sha256"
//  "crypto/hmac"
  "encoding/base64"
)

const server_key = "c2192e43-a4f2-48e4-8904-a8961d80ac6f"

func HashPassword(pass string, salt string) []byte {
  return pbkdf2.Key([]byte(pass), []byte(salt), 1000, 64, sha256.New)
}

func ComparePassword(pass string, salt string, hash []byte) bool {
  return bytes.Compare(HashPassword(pass, salt), hash) == 0
}

func B64decode(str string) ([]byte, error) {
  data, err := base64.StdEncoding.DecodeString(str)
  return data, err
}


type AuthorizationJSON struct {
  Username string `json:"username"`
  Password string `json:"password"`
}


func Authorize(auth_data AuthorizationJSON) bool {
  result := false
  err := users_db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("users"))
        if b != nil {
          user_pass := b.Get([]byte(auth_data.Username))
          if user_pass != nil {
            result = ComparePassword(auth_data.Password, server_key, user_pass)
          }
        } else {
          result = auth_data.Password == "сказочная фея"
        }
        return nil
  })
  return result && err == nil
}


func AuthorizationGuard(next http.Handler) http.Handler {
  return http.HandlerFunc( func (w http.ResponseWriter, r *http.Request) {
      var auth_data AuthorizationJSON
      auth_header, ok := r.Header["Authorization"]
      if ok {
        auth_json, e := B64decode(auth_header[0])
        if e != nil {
          http.Error(w, "failed to decode authorization string", http.StatusBadRequest);
          return
        }
        e = json.Unmarshal(auth_json, &auth_data)
        if e != nil {
          http.Error(w, "can't read authorization json", http.StatusBadRequest);
          return
        } else {
          if Authorize(auth_data) {
            next.ServeHTTP(w, r)
          } else {
            http.Error(w, "authorization failed", http.StatusUnauthorized);
            return
          }
        }
      } else {
        http.Error(w, "authorization header not set", http.StatusBadRequest);
        return
      }
  })
}


func UserAdd(w http.ResponseWriter, r *http.Request) {
    var user_data AuthorizationJSON
    e := json.NewDecoder(r.Body).Decode(&user_data)
    if e != nil {
      http.Error(w, "failed to read json object", http.StatusBadRequest)
      return 
    } else {
      e = users_db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("users"))
        if b == nil {
          b, e = tx.CreateBucket([]byte("users"))
          if e != nil { return e }
        }
        e = b.Put([]byte(user_data.Username), HashPassword(user_data.Password, server_key))
        return e
      })
      if e != nil {
        http.Error(w, "failed to add user", http.StatusInternalServerError)
        return 
      } else {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("success on add user"))
      }
    }
}

