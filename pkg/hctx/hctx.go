package hctx

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"

	"github.com/idproxy/httpserver/internal/pathsegment"
	"github.com/idproxy/httpserver/pkg/params"
	"github.com/idproxy/httpserver/pkg/render"
)

// used to abort the handler processing
const abortIndex int8 = math.MaxInt8 >> 1

type Context interface {
	/************ CONTEXT INIT **************/
	Init(*Config)
	/************ CONTEXT PROCESSING ********/
	UseRawPath() bool
	GetStatus() int
	SetStatus(code int)
	GetMessage() string
	SetMessage(string)
	GetMethod() string
	GetRequest() *http.Request
	GetRequestPath() string
	GetRawRequestPath() string
	GetURLPath() string
	GetParams() params.Params
	GetRawQuery() string
	SetPathSegments(pathsegment.PathSegments)
	SetPathSegmentIndex(int)
	IncrementPathSegmentIndex()
	GetPathSegments() pathsegment.PathSegments
	GetPathSegmentIndex() int
	SetHandlers(HandlerChain)
	GetHandlers() HandlerChain
	ClientIP() string
	RemoteIP() string
	Next()
	Abort()
	AbortWithStatus(code int)

	/************ RENDER RESPONSE ***********/
	Status(code int)
	String(code int, format string, values ...any)
	JSON(code int, obj any)
	Render(code int, r render.Render)
	Writer() http.ResponseWriter
}

func NewContext() Context {
	return &context{}
}

type context struct {
	// set during init
	w                  http.ResponseWriter
	r                  *http.Request
	useRawPath         bool
	unescapePathValues bool

	// dynamic updated during processing
	errs     error
	status   int
	message  string
	index    int8
	handlers HandlerChain
	params   params.Params
	// dynamic set based on useRawPath in Init method
	urlPath string
	// dynamic set based on useRawPath and unescapePathValues in Init method
	unescape bool
	// dynamic context updated during http request processing
	pathSegments   pathsegment.PathSegments
	pathSegmentIdx int
}

/************ CONTEXT INIT ********/

type Config struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  params.Params
	// UseRawPath if enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool
	// Used to unescape the urlraw path
	UnescapePathValues bool
}

func (c *context) Init(cfg *Config) {
	c.w = cfg.Writer
	c.r = cfg.Request
	c.params = cfg.Params
	c.useRawPath = cfg.UseRawPath

	// need to reinitialize the
	c.index = 0
	c.handlers = New()
	c.status = http.StatusOK

	c.urlPath = c.GetRequestPath()
	if c.UseRawPath() && len(c.GetRawRequestPath()) > 0 {
		c.urlPath = c.GetRawRequestPath()
		c.unescape = c.unescapePathValues
	}
}

/************ CONTEXT PROCESSING ********/

func (c *context) GetStatus() int {
	return c.status
}

func (c *context) SetStatus(code int) {
	c.status = code
}

func (c *context) GetMessage() string {
	return c.message
}

func (c *context) SetMessage(s string) {
	c.message = s
}

func (c *context) UseRawPath() bool {
	return c.useRawPath
}

func (c *context) GetMethod() string {
	return c.r.Method
}

func (c *context) GetRequest() *http.Request {
	return c.r
}

// GetRequestPath return the url path from the http request
func (c *context) GetRequestPath() string {
	return c.r.URL.Path
}

// GetRawRequestPath return the raw url path from the http request
func (c *context) GetRawRequestPath() string {
	return c.r.URL.RawPath
}

// GetURLPath return the processed path based on the http request
// and the use Raw config parameters
func (c *context) GetURLPath() string {
	return c.urlPath
}

func (c *context) GetParams() params.Params {
	return c.params
}

func (c *context) GetRawQuery() string {
	return c.r.URL.RawQuery
}

// SetPathSegments sets the path segment for the
func (c *context) SetPathSegments(ps pathsegment.PathSegments) {
	c.pathSegments = ps
}

func (c *context) SetPathSegmentIndex(i int) {
	c.pathSegmentIdx = i
}

func (c *context) IncrementPathSegmentIndex() {
	c.pathSegmentIdx++
}

func (c *context) GetPathSegments() pathsegment.PathSegments {
	return c.pathSegments
}

func (c *context) GetPathSegmentIndex() int {
	return c.pathSegmentIdx
}

func (c *context) SetHandlers(h HandlerChain) {
	c.handlers = h
}

func (c *context) GetHandlers() HandlerChain {
	return c.handlers
}

// ClientIP parses the remote IP and returns the
// TODO proxies
func (c *context) ClientIP() string {
	remoteIP := c.RemoteIP()
	if c.RemoteIP() == "" {
		return ""
	}
	clientIP := net.ParseIP(remoteIP)
	if clientIP == nil {
		return ""
	}
	// TODO trusted proxies
	return clientIP.String()
}

// RemoteIP parses the IP from Request.RemoteAddr, returns the IP (without the port).
// When an error occurs an empty string is returned
func (c *context) RemoteIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.r.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in GitHub.
func (c *context) Next() {
	fmt.Printf("index: %d handler size: %d\n", c.index, c.handlers.Size())
	for c.index < int8(c.handlers.Size()) {
		// call the handler
		fmt.Println(c.handlers.Get(int(c.index)))
		c.handlers.Get(int(c.index))(c)
		c.index++
	}
}

// Abort prevents pending handlers from being called. Note that this will not stop the current handler.
// Let's say you have an authorization middleware that validates that the current request is authorized.
// If the authorization fails (ex: the password does not match), call Abort to ensure the remaining handlers
// for this request are not called.
func (c *context) Abort() {
	c.index = abortIndex
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// For example, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (c *context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

/************ RENDER RESPONSE ***********/

// Status sets the HTTP response code.
func (c *context) Status(code int) {
	c.w.WriteHeader(code)
}

// String writes the given string into the response body.
func (c *context) String(code int, format string, values ...any) {
	c.Render(code, render.String{Format: format, Data: values})
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *context) JSON(code int, obj any) {
	c.Render(code, render.JSON{Data: obj})
}

// Render writes the response headers and calls render.Render to render data.
func (c *context) Render(code int, r render.Render) {
	c.Status(code)

	if err := r.Render(c.w); err != nil {
		c.errs = errors.Join(c.errs, err)
		c.Abort()
	}
}

func (c *context) Writer() http.ResponseWriter {
	return c.w
}
