//The MIT License (MIT)

//Copyright (c) 2014 Ali Najafizadeh

//Permission is hereby granted, free of charge, to any person obtaining a copy of
//this software and associated documentation files (the "Software"), to deal in
//the Software without restriction, including without limitation the rights to
//use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
//the Software, and to permit persons to whom the Software is furnished to do so,
//subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
//FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
//COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
//IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
//CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package relax

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var lookFor = []string{"{", "}", ":number", ":string"}
var replaceWith = []string{"(?P<", ">", "[0-9\\.]+)", "[0-9a-zA-Z_]+)"}

func compileactualURL(actualURL string) (*regexp.Regexp, string) {
	for index := range lookFor {
		actualURL = strings.Replace(actualURL, lookFor[index], replaceWith[index], -1)
	}
	actualURL = "^" + actualURL + "$"
	r, _ := regexp.Compile(actualURL)

	return r, actualURL
}

func applyPath(target string, source *regexp.Regexp) *map[string]string {
	result := make(map[string]string)
	if source.MatchString(target) {
		keys := source.SubexpNames()
		values := source.FindStringSubmatch(target)

		for index, value := range keys {
			if index != 0 {
				result[value] = values[index]
			}
		}
	}
	return &result
}

//RelaxFuncHandler is a exported func for each handler
type RelaxFuncHandler func(req RelaxRequest, res RelaxResponse)

//RelaxHandler is a structure for storing each handler for each regular expression
type RelaxHandler struct {
	method      string
	path        *regexp.Regexp
	funcHandler RelaxFuncHandler
}

//RelaxRequest is a structure that tries to map the params and send it to RelaxFuncHandler as a first argument
type RelaxRequest struct {
	params  *map[string]string
	query   url.Values
	Request *http.Request
}

//Param gets the requested param
func (rr *RelaxRequest) Param(name string) string {
	param, _ := (*rr.params)[name]
	return param
}

//Query gets the querystring value
func (rr *RelaxRequest) Query(name string) string {
	return rr.query.Get(name)
}

//RelaxResponse is a structure that wrapps http.ResponseWriter and send it to RelaxFuncHandler as a second argument
type RelaxResponse struct {
	responseWriter http.ResponseWriter
}

//Send is a method that does 2 things set http status code and write the body to the http pipe
func (rr *RelaxResponse) Send(body string, code int) {
	rr.responseWriter.WriteHeader(code)
	fmt.Fprintf(rr.responseWriter, body)
}

//SendAsJSON is extacly as Send method except, ti convert the body message into json.
func (rr *RelaxResponse) SendAsJSON(body interface{}, code int) {
	result, _ := json.Marshal(body)
	temp := string(result)
	rr.Send(temp, code)
}

//Relax structure holds all the registered relax handlers.
type Relax struct {
	relaxHandlerMap map[string]*RelaxHandler
}

func (r *Relax) mainHandler(w http.ResponseWriter, req *http.Request) {
	var statusCode = http.StatusNotFound
	var relaxHandler *RelaxHandler

	url := req.URL
	method := req.Method
	relaxResponse := RelaxResponse{w}

	for _, value := range r.relaxHandlerMap {
		if value.path.MatchString(url.Path) {
			if value.method == method {
				relaxHandler = value
				statusCode = http.StatusOK
			} else {
				statusCode = http.StatusMethodNotAllowed
			}
			break
		}
	}

	if relaxHandler != nil && statusCode == http.StatusOK {
		params := applyPath(url.Path, relaxHandler.path)
		query := url.Query()
		relaxHandler.funcHandler(RelaxRequest{params, query, req}, relaxResponse)
	} else {
		var errorMessage string
		if statusCode == http.StatusNotFound {
			errorMessage = "Not Found"
		} else if statusCode == http.StatusMethodNotAllowed {
			errorMessage = "Method Not Allowed"
		}

		relaxResponse.Send(errorMessage, statusCode)
	}
}

//RegisterHandler registers handlers for each REST API. if method + path is not unique it returns false.
func (r Relax) RegisterHandler(method string, path string, funcHandler RelaxFuncHandler) bool {
	regexpPath, actualPath := compileactualURL(path)

	if value, ok := r.relaxHandlerMap[actualPath]; ok && value.method == method {
		return false
	}

	r.relaxHandlerMap[actualPath] = &RelaxHandler{method, regexpPath, funcHandler}
	return true
}

//Listen starts listening to requested host and port.
func (r Relax) Listen(host string, port int) {
	http.HandleFunc("/", r.mainHandler)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}

//NewRelax creates Relax structure
func NewRelax() *Relax {
	relax := Relax{}

	relax.relaxHandlerMap = make(map[string]*RelaxHandler)

	return &relax
}
