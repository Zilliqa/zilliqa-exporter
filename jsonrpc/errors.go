package jsonrpc

import (
	"fmt"
)

var (
	ParserError    = RPCError{Code: -32700, Message: "Invalid JSON was received by the server. An error occurred on the server while parsing the JSON text."}
	InvalidRequest = RPCError{Code: -32600, Message: "The JSON sent is not a valid Request object."}
	MethodNotFound = RPCError{Code: -32601, Message: "The method does not exist / is not available."}
	InvalidParams  = RPCError{Code: -32602, Message: "Invalid method parameter(s)."}
	InternalError  = RPCError{Code: -32603, Message: "Internal JSON-RPC error."}
	ServerError    = RPCError{Code: -32000, Message: "Reserved for implementation-defined server-errors."}
)

type RPCError struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e RPCError) Error() string {
	return fmt.Sprintf("code: %d %s", e.Code, e.Message)
}

func (e RPCError) IsRpcError(err RPCError) bool {
	if err.Code == ServerError.Code && e.Code > -32099 && e.Code < -32000 {
		return true
	}
	return e.Code == err.Code
}

// for errors.Is
func (e RPCError) Is(err error) bool {
	if rpcErr, ok := err.(RPCError); ok {
		return e.IsRpcError(rpcErr)
	}
	return false
}
