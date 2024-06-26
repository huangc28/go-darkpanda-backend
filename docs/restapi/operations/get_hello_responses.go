// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// GetHelloOKCode is the HTTP code returned for type GetHelloOK
const GetHelloOKCode int = 200

/*GetHelloOK get hello o k

swagger:response getHelloOK
*/
type GetHelloOK struct {
}

// NewGetHelloOK creates GetHelloOK with default headers values
func NewGetHelloOK() *GetHelloOK {

	return &GetHelloOK{}
}

// WriteResponse to the client
func (o *GetHelloOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}
