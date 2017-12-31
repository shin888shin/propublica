package main

// go run propublica.go
// curl -XPOST -d'{"s":"aaa, bb"}' localhost:9090/count     // {"v":7}
// curl -XPOST -d'{"s":"aaa, bb"}' localhost:9090/uppercase // {"v":"AAA, BB"}
// curl -XPOST -d'{"search":"ecoviva"}' localhost:9090/fetch
// curl -XPOST -d'{"search":"oakland"}' localhost:9090/fetch
// browser: http://localhost:9090/jesus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	// "github.com/davecgh/go-spew/spew"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type PropublicaService interface {
	FetchByName(string) (PResponse, error)
}

// StringService provides operations on strings.
type StringService interface {
	Uppercase(string) (string, error)
	Count(string) int
	Concat(string) (string, error)
}

type stringService struct{}
type pService struct{}

func (pService) FetchByName(s string) (PResponse, error) {
	var pr PResponse
	resp, err := http.Get("https://projects.propublica.org/nonprofits/api/v2/search.json?q=" + s)
	if err != nil {
		fmt.Println("Error:", err)
		// os.Exit(1)
		return pr, err
	}
	// fmt.Printf("+++> FetchByName resp: %T - %v\n", resp, resp)
	err = json.NewDecoder(resp.Body).Decode(&pr)
	// fmt.Printf("+++> FetchByName pr: %T - %v\n", pr, pr)
	return pr, nil
}

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
	return strings.Replace(s, " ", "", -1), nil
}

func (stringService) Count(s string) int {
	return len(s)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n\n", r.URL.Path[1:])
	fmt.Fprintf(w, "%s\n\n", r.UserAgent())

	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "%s\n\n", name)

	addrs, _ := net.LookupIP(name)
	// fmt.Fprintf(w, "\n")
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Fprintf(w, "%s\n", ipv4)
		}
	}

}

func main() {
	svc := stringService{}
	psvc := pService{}

	fetchByNameHandler := httptransport.NewServer(
		makeFetchByNameEndpoint(psvc),
		decodeFetchByNameRequest,
		encodeFetchByNameResponse,
	)

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
	http.HandleFunc("/", handler)
	http.Handle("/concat", concatHandler)
	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	http.Handle("/fetch", fetchByNameHandler)
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func makeFetchByNameEndpoint(psvc PropublicaService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(fetchByNameRequest) // must cast interface{}
		presp, err := psvc.FetchByName(req.Name)
		// fmt.Printf("\n\n\n-x-> presp: %T - %v\n\n", presp, presp)
		// fmt.Printf("\n\n\n---> err: %T - %v\n\n", err, err)
		if err != nil {
			return presp, err
		}
		return presp, nil
	}
}

func makeUppercaseEndpoint(svc StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// spew.Dump(ctx)
		// fmt.Printf("---> request: %T - %v\n", request, request)
		req := request.(uppercaseRequest)
		// fmt.Printf("---> req: %T - %v\n", req, req)
		v, err := svc.Uppercase(req.S)
		// fmt.Printf("---> v: %T - %v\n", v, v)
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

func decodeFetchByNameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request fetchByNameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	// fmt.Printf("ooo> decodeFetchByNameRequest: %T - %v\n", request, request)
	return request, nil
}

func encodeFetchByNameResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	// fmt.Printf("ooo> encodeFetchByNameResponse: %T - %v\n", response, response)
	return json.NewEncoder(w).Encode(response)
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	// fmt.Printf("ooo> encodeResponse: %T - %v\n", response, response)
	return json.NewEncoder(w).Encode(response)
}

type fetchByNameRequest struct {
	Name string `json:"search"`
}

type fetchByNameResponse struct {
	V   string // `json:"total_results"`
	Err string `json:"err,omitempty"` // errors don't define JSON marshaling
}

type (
	POrganization struct {
		Ein          int     `json:"ein"`
		Strein       string  `json:"strein"`
		Name         string  `json:"name"`
		SubName      string  `json:"sub_name"`
		City         string  `json:"city"`
		State        string  `json:"state"`
		NteeCode     string  `json:"ntee_code"`
		RawNteeCode  string  `json:"raw_ntee_code"`
		Subseccd     int     `json:"subseccd"`
		HasSubseccd  bool    `json:"has_subseccd"`
		HaveFilings  bool    `json:"have_filings"`
		HaveExtracts bool    `json:"have_extracts"`
		HavePdfs     bool    `json:"have_pdfs"`
		Score        float32 `json:"score"`
	}
	PResponse struct {
		Organizations []POrganization `json:organizations`
		TotalResults  int             `json:"total_results"`
		Error         string          `json:"err,omitempty"`
	}
	PRequest struct {
		Name string `json:"search"`
	}
)

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
