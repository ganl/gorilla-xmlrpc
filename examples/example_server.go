package main

import (
	"log"
	"net/http"

	"github.com/ganl/gorilla-xmlrpc/xml"
	"github.com/gorilla/rpc"
)

type HelloService struct{}

func (h *HelloService) Say(r *http.Request, args *struct{ Who string }, reply *struct{ Message string }) error {
	log.Println("Say", args.Who)
	reply.Message = "Hello, " + args.Who + "!"
	return nil
}

func main() {
	RPC := rpc.NewServer()
	xmlrpcCodec := xml.NewCodec(xml.WithSnakeTrans(true))
	xmlrpcCodec.RegisterAlias(`rpc.hello`, `HelloService.Say`)
	RPC.RegisterCodec(xmlrpcCodec, "text/xml")
	RPC.RegisterService(new(HelloService), "")
	RPC.RegisterService(new(HelloService), "test")
	http.Handle("/RPC2", RPC)

	log.Println("Starting XML-RPC server on localhost:1234/RPC2")
	log.Fatal(http.ListenAndServe(":1234", nil))
}
