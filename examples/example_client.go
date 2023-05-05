package main

import (
	"bytes"
	"log"
	"net/http"

	"github.com/ganl/gorilla-xmlrpc/xml"
)

func XmlRpcCall(method string, args struct{ Who string }) (reply struct{ Message string }, err error) {
	buf, _ := xml.EncodeClientRequest(method, &args)

	resp, err := http.Post("http://localhost:1234/RPC2", "text/xml", bytes.NewBuffer(buf))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = xml.DecodeClientResponse(resp.Body, &reply)
	return
}

func main() {
	// normal call
	reply, err := XmlRpcCall("HelloService.Say", struct{ Who string }{"User 1"})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response: %s\n", reply.Message)

	// Snake to Camel
	reply2, err := XmlRpcCall("test.say", struct{ Who string }{"User 2"})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response: %s\n", reply2.Message)

	// Register Alias
	reply3, err := XmlRpcCall("rpc.hello", struct{ Who string }{"User 3"})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response: %s\n", reply3.Message)
}
