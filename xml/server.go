// Copyright 2013 Ivan Danyliuk
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xml

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/rpc"
)

var snakeReg = regexp.MustCompile("(^[A-Za-z]+|_[A-Za-z])")

func Snake2Camel(str string) string {
	return snakeReg.ReplaceAllStringFunc(str, func(s string) string {
		return strings.Title(strings.TrimLeft(s, "_"))
	})
}

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

type Option func(*Codec)

func WithSnakeTrans(snakeTrans bool) Option {
	return func(c *Codec) {
		c.snakeTrans = snakeTrans
	}
}

// NewCodec returns a new XML-RPC Codec.
func NewCodec(opts ...Option) *Codec {
	c := &Codec{
		aliases: make(map[string]string),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Codec creates a CodecRequest to process each request.
type Codec struct {
	aliases    map[string]string
	snakeTrans bool
}

// RegisterAlias creates a method alias
func (c *Codec) RegisterAlias(alias, method string) {
	c.aliases[alias] = method
}

// NewRequest returns a CodecRequest.
func (c *Codec) NewRequest(r *http.Request) rpc.CodecRequest {
	rawxml, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &CodecRequest{err: err}
	}
	defer r.Body.Close()

	var request ServerRequest
	if err := xml.Unmarshal(rawxml, &request); err != nil {
		return &CodecRequest{err: err}
	}
	request.rawxml = string(rawxml)
	if method, ok := c.aliases[request.Method]; ok {
		request.Method = method
	} else {
		if c.snakeTrans {
			rpcMethod := request.Method
			parts := strings.Split(rpcMethod, ".")
			if len(parts) != 2 {
				request.Method = Snake2Camel(parts[0])
			} else {
				request.Method = parts[0] + "." + Snake2Camel(parts[1])
			}
		}
	}
	return &CodecRequest{request: &request}
}

// ----------------------------------------------------------------------------
// CodecRequest
// ----------------------------------------------------------------------------

type ServerRequest struct {
	Name   xml.Name `xml:"methodCall"`
	Method string   `xml:"methodName"`
	rawxml string
}

// CodecRequest decodes and encodes a single request.
type CodecRequest struct {
	request *ServerRequest
	err     error
}

// Method returns the RPC method for the current request.
//
// The method uses a dotted notation as in "Service.Method".
func (c *CodecRequest) Method() (string, error) {
	if c.err == nil {
		return c.request.Method, nil
	}
	return "", c.err
}

// ReadRequest fills the request object for the RPC method.
//
// args is the pointer to the Service.Args structure
// it gets populated from temporary XML structure
func (c *CodecRequest) ReadRequest(args interface{}) error {
	c.err = xml2RPC(c.request.rawxml, args)
	return nil
}

// WriteResponse encodes the response and writes it to the ResponseWriter.
//
// response is the pointer to the Service.Response structure
// it gets encoded into the XML-RPC xml string
func (c *CodecRequest) WriteResponse(w http.ResponseWriter, response interface{}, methodErr error) error {
	var xmlstr string
	if c.err != nil {
		var fault Fault
		switch c.err.(type) {
		case Fault:
			fault = c.err.(Fault)
		default:
			fault = FaultApplicationError
			fault.String += fmt.Sprintf(": %v", c.err)
		}
		xmlstr = fault2XML(fault)
	} else {
		xmlstr, _ = rpcResponse2XML(response)
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write([]byte(xmlstr))
	return nil
}
