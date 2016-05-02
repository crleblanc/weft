/*
weft helps with web applications.
*/
package weft

import (
	"bytes"
	"github.com/GeoNet/mtr/mtrapp"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

// Return pointers to these as required.
var (
	StatusOK         = Result{Ok: true, Code: http.StatusOK, Msg: ""}
	MethodNotAllowed = Result{Ok: false, Code: http.StatusMethodNotAllowed, Msg: "method not allowed"}
	NotFound         = Result{Ok: false, Code: http.StatusNotFound, Msg: "not found"}
	NotAcceptable    = Result{Ok: false, Code: http.StatusNotAcceptable, Msg: "specify accept"}
)

type Result struct {
	Ok   bool   // set true to indicate success
	Code int    // http status code for writing back to the client e.g., http.StatusOK for success.
	Msg  string // any error message for logging or to send to the client.
}

type RequestHandler func(r *http.Request, h http.Header, b *bytes.Buffer) *Result

func InternalServerError(err error) *Result {
	return &Result{Ok: false, Code: http.StatusInternalServerError, Msg: err.Error()}
}

func ServiceUnavailableError(err error) *Result {
	return &Result{Ok: false, Code: http.StatusServiceUnavailable, Msg: err.Error()}
}

func BadRequest(message string) *Result {
	return &Result{Ok: false, Code: http.StatusBadRequest, Msg: message}
}

/*
CheckQuery inspects r and makes sure all required query parameters
are present and that no more than the required and optional parameters
are present.
*/
func CheckQuery(r *http.Request, required, optional []string) *Result {
	if strings.Contains(r.URL.Path, ";") {
		return BadRequest("cache buster")
	}

	v := r.URL.Query()

	if len(required) == 0 && len(optional) == 0 {
		if len(v) == 0 {
			return &StatusOK
		} else {
			return BadRequest("found unexpected query parameters")
		}
	}

	var missing []string

	for _, k := range required {
		if v.Get(k) == "" {
			missing = append(missing, k)
		} else {
			v.Del(k)
		}
	}

	switch len(missing) {
	case 0:
	case 1:
		return BadRequest("missing required query parameter: " + missing[0])
	default:
		return BadRequest("missing required query parameters: " + strings.Join(missing, ", "))
	}

	for _, k := range optional {
		v.Del(k)
	}

	if len(v) > 0 {
		return BadRequest("found additional query parameters")
	}

	return &StatusOK
}

// name finds the name of the function f
func name(f RequestHandler) string {
	var n string
	// Find the name of the function f to use as the timer id
	fn := runtime.FuncForPC(reflect.ValueOf(f).Pointer())
	if fn != nil {
		n = fn.Name()
		i := strings.LastIndex(n, ".")
		if i > 0 && i+1 < len(n) {
			n = n[i+1 : len(n)]
		}
	}
	return n
}

// count increments mtr counters for Result.
func (r *Result) Count() {
	if r != nil && r.Code != 0 {
		mtrapp.Requests.Inc()

		switch r.Code {
		case http.StatusOK:
			mtrapp.StatusOK.Inc()
		case http.StatusBadRequest:
			mtrapp.StatusBadRequest.Inc()
		case http.StatusUnauthorized:
			mtrapp.StatusUnauthorized.Inc()
		case http.StatusNotFound:
			mtrapp.StatusNotFound.Inc()
		case http.StatusInternalServerError:
			mtrapp.StatusInternalServerError.Inc()
		case http.StatusServiceUnavailable:
			mtrapp.StatusServiceUnavailable.Inc()
		}
	}
}
