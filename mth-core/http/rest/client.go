package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	httplog "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/http/log"
	webcontext "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/web/context"
	h "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/web/header"
)

// Types

// Client is an HTTP Request builder and sender.
type Client struct {
	// http Client for doing requests
	httpClient *http.Client
	// raw url string for requests
	baseURL string
}

// Endpoint is HTTP endpoint
type Endpoint struct {
	// Client is an HTTP Client used to make request to this endpoint
	Client *Client
	// raw url string for requests
	rawURL string
	// HTTP method (GET, POST, etc.)
	method string
	// stores key-values pairs to add to request's Headers
	header http.Header
	// url tagged raw query string
	rawQuery *string
	// body
	body interface{}
	//formData
	formData url.Values
	// request context
	context context.Context
}

// NewClient returns a new Client with an http DefaultClient.
func NewClient(apiBaseURL string) *Client {
	return &Client{
		httpClient: http.DefaultClient,
		baseURL:    apiBaseURL,
	}
}

// NewEndpoint returns a new Endpoint with default 'GET' method
func (client *Client) NewEndpoint(ctx context.Context) *Endpoint {
	e := &Endpoint{
		Client:  client,
		rawURL:  client.baseURL,
		header:  make(http.Header),
		method:  "GET",
		context: ctx,
	}

	// Set correlation ID header field if it exists in context.
	if cID := webcontext.CorrelationID(e.context); len(cID) > 0 {
		e.WithHeader(h.HeaderKeyCorrelationID, cID)
	}

	return e
}

// WithClient sets the http Client used to do requests. If a nil client is given,
// the http.DefaultClient will be used.
func (client *Client) WithClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		client.httpClient = http.DefaultClient
	} else {
		client.httpClient = httpClient
	}

	return client
}

// Path extends the Endpoint rawURL with the given path by resolving the reference to
// an absolute URL. If parsing errors occur, the rawURL is left unmodified.
func (e *Endpoint) Path(URI string) *Endpoint {
	parsedURL, parsedURLError := url.Parse(e.rawURL)
	parsedURI, parsedURIError := url.Parse(URI)

	if parsedURLError != nil || parsedURIError != nil {
		return e
	}

	parsedURL.Path = path.Join(parsedURL.Path, parsedURI.Path)
	parsedURL.RawQuery = parsedURI.RawQuery
	parsedURL.Fragment = parsedURI.Fragment

	e.rawURL = parsedURL.String()

	return e
}

// Header

// WithHeader sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalized.
func (e *Endpoint) WithHeader(key, value string) *Endpoint {
	e.header.Set(key, value)
	return e
}

// WithAuth sets the Authorization header to use provided Authentication string
func (e *Endpoint) WithAuth(authToken *string) *Endpoint {
	if authToken != nil {
		return e.WithHeader("Authorization", *authToken)
	}

	return e
}

// WithTypedAuth sets the Authorization header to use provided Authentication type
// with the provided token.
func (e *Endpoint) WithTypedAuth(authType, token string) *Endpoint {
	return e.WithHeader("Authorization", fmt.Sprintf("%s %s", authType, token))
}

// WithBearerAuth sets Authorization header to use HTTP Bearer Authentication
// with the provided JWT token.
func (e *Endpoint) WithBearerAuth(token string) *Endpoint {
	return e.WithTypedAuth("Bearer", token)
}

// Body

// WithBody sets the Endpoints's body. The body value will be set as the Body on new
// requests.
func (e *Endpoint) WithBody(body interface{}) *Endpoint {
	e.body = body
	return e
}

// WithFormDataBody sets the Endpoints's form data.
func (e *Endpoint) WithFormDataBody(formData url.Values) *Endpoint {
	e.formData = formData
	return e
}

// URL

// WithQuery sets the Endpoints's raw query string. The rawQuery value will be set as the url.RawQuery
// on new requests.
func (e *Endpoint) WithQuery(rawQuery string) *Endpoint {
	e.rawQuery = &rawQuery
	return e
}

// Methods

// Get sets the Endpoint method to GET and sets the given pathURL.
func (e *Endpoint) Get(pathURL string) *Endpoint {
	e.method = "GET"
	return e.Path(pathURL)
}

// Post sets the Endpoint method to POST and sets the given pathURL.
func (e *Endpoint) Post(pathURL string) *Endpoint {
	e.method = "POST"
	return e.Path(pathURL)
}

// Put sets the Endpoint method to PUT and sets the given pathURL.
func (e *Endpoint) Put(pathURL string) *Endpoint {
	e.method = "PUT"
	return e.Path(pathURL)
}

// Patch sets the Endpoint method to PATCH and sets the given pathURL.
func (e *Endpoint) Patch(pathURL string) *Endpoint {
	e.method = "PATCH"
	return e.Path(pathURL)
}

// Delete sets the Sling method to DELETE and sets the given pathURL.
func (e *Endpoint) Delete(pathURL string) *Endpoint {
	e.method = "DELETE"
	return e.Path(pathURL)
}

// Requests

// Request returns a new http.Request created with the Endpoint properties.
// Returns any errors parsing the rawURL, encoding query structs, encoding
// the body, or creating the http.Request.
func (e *Endpoint) Request() (*http.Request, error) {
	reqURL, err := url.Parse(e.rawURL)
	if err != nil {
		return nil, err
	}

	if e.rawQuery != nil {
		reqURL.RawQuery = *e.rawQuery
	}

	var bodyBuf io.ReadWriter
	if e.body != nil {
		b := new(bytes.Buffer)
		err := json.NewEncoder(b).Encode(e.body)
		if err != nil {
			return nil, err
		}
		bodyBuf = b
	}

	var req *http.Request
	if e.formData != nil {
		req, err = http.NewRequest(e.method, reqURL.String(), strings.NewReader(e.formData.Encode()))
	} else {
		req, err = http.NewRequest(e.method, reqURL.String(), bodyBuf)
	}

	if err != nil {
		return nil, err
	}

	// add context with correlationId to request
	req = req.WithContext(e.context)

	e.addHeaders(req)
	return req, err
}

// addHeaders adds the key, value pairs from the given http.Header to the
// request. Values for existing keys are appended to the keys values.
func (e *Endpoint) addHeaders(req *http.Request) {
	if e.formData != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
	}

	for key, values := range e.header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

// SendAndParse creates a new HTTP request and parses response.
// Success responses (2XX or 3XX) are JSON decoded into the value pointed to by successV and
// other responses are JSON decoded into the value pointed to by failureV.
// Any error creating the request, sending it, or decoding the response is
// returned.
func (e *Endpoint) SendAndParse(successV, failureV interface{}) (*http.Response, error) {
	req, err := e.Request()
	if err != nil {
		return nil, err
	}
	return e.do(req, successV, failureV)
}

// do sends an HTTP request and returns the response.
// Success responses (2XX or 3XX) are JSON decoded into the value pointed to by successV.
// Other responses are JSON decoded into the value pointed to by failureV.
// Any error sending the request or decoding the response is returned.
func (e *Endpoint) do(req *http.Request, successV, failureV interface{}) (*http.Response, error) {
	httplog.RequestLogger(req).Info("http request")

	resp, err := e.Client.httpClient.Do(req)
	l := httplog.ResponseLogger(resp)
	if err != nil {
		l.Error("http response")
		return resp, err
	}

	// when err is nil, resp contains a non-nil resp.Body which must be closed
	defer resp.Body.Close()

	// Log error  or info
	if resp.StatusCode >= 500 {
		l.Error("http response")
	} else {
		l.Info("http response")
	}

	// Don't try to decode on 204s
	if resp.StatusCode == 204 {
		return resp, nil
	}

	// Decode from json
	if successV != nil || failureV != nil {
		err = decodeResponseJSON(resp, successV, failureV)
	}
	return resp, err
}

// decodeResponse decodes response Body into the value pointed to by successV
// if the response is a success (2XX and 3XX) or into the value pointed to by failureV
// otherwise. If the successV or failureV argument to decode into is nil,
// decoding is skipped.
// Caller is responsible for closing the resp.Body.
func decodeResponseJSON(resp *http.Response, successV, failureV interface{}) error {
	if code := resp.StatusCode; 200 <= code && code <= 399 {
		if successV != nil {
			return decodeResponseBodyJSON(resp, successV)
		}
	} else {
		if failureV != nil {
			return decodeResponseBodyJSON(resp, failureV)
		}
	}
	return nil
}

// decodeResponseBodyJSON JSON decodes a Response Body into the value pointed
// to by v.
// Caller must provide a non-nil v and close the resp.Body.
func decodeResponseBodyJSON(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}
