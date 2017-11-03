package main

// curl -XPOST -d'{"s":"aaa, bb"}' localhost:8080/count
// {"v":7}
// curl -XPOST -d'{"s":"aaa, bb"}' localhost:8080/uppercase
// {"v":"AAA, BB"}

import (
  "context"
  "encoding/json"
  "errors"
  "log"
  "net/http"
  "strings"
  "fmt"

  "github.com/go-kit/kit/endpoint"
  httptransport "github.com/go-kit/kit/transport/http"

  "github.com/davecgh/go-spew/spew"
)

// StringService provides operations on strings.
type StringService interface {
  Uppercase(string) (string, error)
  Count(string) int
  Concat(string) (string, error)
}

type stringService struct{}

func (stringService) Uppercase(s string) (string, error) {
  if s == "" {
    return "", ErrEmpty
  }
  return strings.ToUpper(s), nil
}

func (stringService) Concat(s string) (string, error) {
  if s == "" {
    return "", ErrEmpty
  }
  return strings.Replace(s," ", "", -1), nil
}

func (stringService) Count(s string) int {
  return len(s)
}

func main() {
  svc := stringService{}

  uppercaseHandler := httptransport.NewServer(
    makeUppercaseEndpoint(svc),
    decodeUppercaseRequest,
    encodeResponse,
  )

  concatHandler := httptransport.NewServer(
    makeConcatEndpoint(svc),
    decodeConcatRequest,
    encodeResponse,
  )

  countHandler := httptransport.NewServer(
    makeCountEndpoint(svc),
    decodeCountRequest,
    encodeResponse,
  )
  http.Handle("/concat", concatHandler)
  http.Handle("/uppercase", uppercaseHandler)
  http.Handle("/count", countHandler)
  log.Fatal(http.ListenAndServe(":8080", nil))
}

func makeUppercaseEndpoint(svc StringService) endpoint.Endpoint {
  return func(ctx context.Context, request interface{}) (interface{}, error) {
    spew.Dump(ctx)
    fmt.Println("---> 1")
    fmt.Printf("---> 2 %T\n", ctx)
    req := request.(uppercaseRequest)
    v, err := svc.Uppercase(req.S)
    if err != nil {
      return uppercaseResponse{v, err.Error()}, nil
    }
    return uppercaseResponse{v, ""}, nil
  }
}

func makeConcatEndpoint(svc StringService) endpoint.Endpoint {
  return func(ctx context.Context, request interface{}) (interface{}, error) {
    req := request.(concatRequest)
    v, err := svc.Concat(req.S)
    if err != nil {
      return concatResponse{v, err.Error()}, nil
    }
    return concatResponse{v, ""}, nil
  }
}

func makeCountEndpoint(svc StringService) endpoint.Endpoint {
  return func(ctx context.Context, request interface{}) (interface{}, error) {
    req := request.(countRequest)
    v := svc.Count(req.S)
    return countResponse{v}, nil
  }
}

func decodeUppercaseRequest(_ context.Context, r *http.Request) (interface{}, error) {
  var request uppercaseRequest
  if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
    return nil, err
  }
  return request, nil
}

func decodeConcatRequest(_ context.Context, r *http.Request) (interface{}, error) {
  var request concatRequest
  if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
    return nil, err
  }
  return request, nil
}

func decodeCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
  var request countRequest
  if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
    return nil, err
  }
  return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
  return json.NewEncoder(w).Encode(response)
}

type uppercaseRequest struct {
  S string `json:"s"`
}

type uppercaseResponse struct {
  V   string `json:"v"`
  Err string `json:"err,omitempty"` // errors don't define JSON marshaling
}

type concatRequest struct {
  S string `json:"s"`
}

type concatResponse struct {
  V   string `json:"v"`
  Err string `json:"err,omitempty"` // errors don't define JSON marshaling
}

type countRequest struct {
  S string `json:"s"`
}

type countResponse struct {
  V int `json:"v"`
}

// ErrEmpty is returned when an input string is empty.
var ErrEmpty = errors.New("empty string")