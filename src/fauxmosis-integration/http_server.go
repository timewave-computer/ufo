package fauxmosisintegration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
)

// RPCHTTPServer serves HTTP requests for the RPC server
type RPCHTTPServer struct {
	client  *RPCClient
	server  *http.Server
	address string
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// NewRPCHTTPServer creates a new HTTP server for RPC
func NewRPCHTTPServer(client *RPCClient, address string) *RPCHTTPServer {
	return &RPCHTTPServer{
		client:  client,
		address: address,
	}
}

// Start starts the HTTP server
func (s *RPCHTTPServer) Start() error {
	router := http.NewServeMux()
	router.HandleFunc("/", s.handleRequest)

	s.server = &http.Server{
		Addr:    s.address,
		Handler: router,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	fmt.Printf("RPC HTTP server started on %s\n", s.address)
	return nil
}

// Stop stops the HTTP server
func (s *RPCHTTPServer) Stop() error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(context.Background())
}

// handleRequest handles a JSON-RPC request
func (s *RPCHTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" && r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var (
		req    RPCRequest
		params []interface{}
		method string
	)

	// Handle both GET and POST requests
	if r.Method == "POST" {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse the JSON-RPC request
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, fmt.Sprintf("Error parsing JSON-RPC request: %v", err), http.StatusBadRequest)
			return
		}

		method = req.Method
		params = req.Params
	} else { // GET
		// Parse the URL query parameters
		method = r.URL.Query().Get("method")
		if method == "" {
			http.Error(w, "Missing method parameter", http.StatusBadRequest)
			return
		}

		// Parse parameters based on method
		params = parseGetParams(r, method)
	}

	// Call the appropriate method on the RPC client
	result, err := s.callRPCMethod(method, params)
	if err != nil {
		resp := RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32603, // Internal error
				Message: err.Error(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Marshal the result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		resp := RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32603, // Internal error
				Message: fmt.Sprintf("Error marshaling result: %v", err),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Send the response
	resp := RPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  resultJSON,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// callRPCMethod calls a method on the RPC client
func (s *RPCHTTPServer) callRPCMethod(method string, params []interface{}) (interface{}, error) {
	// Get the method by name
	clientVal := reflect.ValueOf(s.client)
	methodVal := clientVal.MethodByName(method)
	if !methodVal.IsValid() {
		return nil, fmt.Errorf("unknown method: %s", method)
	}

	// Get the method type
	methodType := methodVal.Type()
	numIn := methodType.NumIn()
	if numIn > 0 && !methodType.In(0).AssignableTo(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return nil, fmt.Errorf("method %s does not accept context as first parameter", method)
	}

	// Prepare input parameters
	args := make([]reflect.Value, numIn)
	argIndex := 0

	// First parameter should be context
	if numIn > 0 {
		args[argIndex] = reflect.ValueOf(context.Background())
		argIndex++
	}

	// Handle remaining parameters
	for i := 0; argIndex < numIn && i < len(params); i++ {
		paramType := methodType.In(argIndex)
		param := params[i]

		// Attempt to convert the parameter to the expected type
		paramVal, err := convertParam(param, paramType)
		if err != nil {
			return nil, fmt.Errorf("error converting parameter %d: %v", i, err)
		}

		args[argIndex] = paramVal
		argIndex++
	}

	// Call the method
	results := methodVal.Call(args)
	if len(results) != 2 {
		return nil, fmt.Errorf("method %s does not return (result, error)", method)
	}

	// Handle error
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	// Return the result
	return results[0].Interface(), nil
}

// convertParam converts a parameter to the expected type
func convertParam(param interface{}, paramType reflect.Type) (reflect.Value, error) {
	paramVal := reflect.ValueOf(param)

	// If the parameter is nil, check if the expected type is a pointer or interface
	if param == nil {
		if paramType.Kind() == reflect.Ptr || paramType.Kind() == reflect.Interface {
			return reflect.Zero(paramType), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot use nil for non-pointer type %s", paramType)
	}

	// If the parameter is already assignable to the expected type, use it directly
	if paramVal.Type().AssignableTo(paramType) {
		return paramVal, nil
	}

	// For numeric types, try to convert
	switch paramType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Try to convert to int64
		var intVal int64
		switch v := param.(type) {
		case float64:
			intVal = int64(v)
		case float32:
			intVal = int64(v)
		case int:
			intVal = int64(v)
		case int8:
			intVal = int64(v)
		case int16:
			intVal = int64(v)
		case int32:
			intVal = int64(v)
		case int64:
			intVal = v
		case string:
			var err error
			intVal, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert string to int: %v", err)
			}
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to int", param)
		}
		return reflect.ValueOf(intVal).Convert(paramType), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Try to convert to uint64
		var uintVal uint64
		switch v := param.(type) {
		case float64:
			uintVal = uint64(v)
		case float32:
			uintVal = uint64(v)
		case uint:
			uintVal = uint64(v)
		case uint8:
			uintVal = uint64(v)
		case uint16:
			uintVal = uint64(v)
		case uint32:
			uintVal = uint64(v)
		case uint64:
			uintVal = v
		case string:
			var err error
			uintVal, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert string to uint: %v", err)
			}
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to uint", param)
		}
		return reflect.ValueOf(uintVal).Convert(paramType), nil

	case reflect.Float32, reflect.Float64:
		// Try to convert to float64
		var floatVal float64
		switch v := param.(type) {
		case float64:
			floatVal = v
		case float32:
			floatVal = float64(v)
		case int:
			floatVal = float64(v)
		case int64:
			floatVal = float64(v)
		case string:
			var err error
			floatVal, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert string to float: %v", err)
			}
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to float", param)
		}
		return reflect.ValueOf(floatVal).Convert(paramType), nil

	case reflect.Bool:
		// Try to convert to bool
		var boolVal bool
		switch v := param.(type) {
		case bool:
			boolVal = v
		case string:
			var err error
			boolVal, err = strconv.ParseBool(v)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert string to bool: %v", err)
			}
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", param)
		}
		return reflect.ValueOf(boolVal), nil

	case reflect.String:
		// Try to convert to string
		var strVal string
		switch v := param.(type) {
		case string:
			strVal = v
		case []byte:
			strVal = string(v)
		default:
			// Try using fmt.Sprint for other types
			strVal = fmt.Sprint(v)
		}
		return reflect.ValueOf(strVal), nil

	case reflect.Slice:
		// If the expected type is []byte and the param is a string, convert
		if paramType.Elem().Kind() == reflect.Uint8 && paramVal.Kind() == reflect.String {
			strVal := param.(string)
			return reflect.ValueOf([]byte(strVal)), nil
		}
		// For other slice types, we need to convert each element
		if paramVal.Kind() == reflect.Slice {
			// Create a new slice of the expected type
			newSlice := reflect.MakeSlice(paramType, paramVal.Len(), paramVal.Len())
			for i := 0; i < paramVal.Len(); i++ {
				// Convert each element
				elemVal, err := convertParam(paramVal.Index(i).Interface(), paramType.Elem())
				if err != nil {
					return reflect.Value{}, fmt.Errorf("cannot convert slice element %d: %v", i, err)
				}
				newSlice.Index(i).Set(elemVal)
			}
			return newSlice, nil
		}
		return reflect.Value{}, fmt.Errorf("cannot convert %T to slice", param)

	case reflect.Ptr:
		// If the expected type is a pointer, convert the param to the pointed type
		// and then take its address
		elemVal, err := convertParam(param, paramType.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		// Create a new pointer and set it to the converted value
		ptrVal := reflect.New(paramType.Elem())
		ptrVal.Elem().Set(elemVal)
		return ptrVal, nil

	default:
		return reflect.Value{}, fmt.Errorf("unsupported parameter type: %s", paramType)
	}
}

// parseGetParams parses parameters from GET request query for known methods
func parseGetParams(r *http.Request, method string) []interface{} {
	q := r.URL.Query()

	switch method {
	case "ABCIInfo":
		return []interface{}{}

	case "ABCIQuery":
		path := q.Get("path")
		data := q.Get("data")
		heightStr := q.Get("height")
		proveStr := q.Get("prove")

		var height int64
		if heightStr != "" {
			height, _ = strconv.ParseInt(heightStr, 10, 64)
		}

		var prove bool
		if proveStr != "" {
			prove, _ = strconv.ParseBool(proveStr)
		}

		return []interface{}{path, []byte(data), height, prove}

	case "BroadcastTxAsync", "BroadcastTxSync", "BroadcastTxCommit":
		tx := q.Get("tx")
		return []interface{}{[]byte(tx)}

	case "Block":
		heightStr := q.Get("height")
		if heightStr == "" {
			return []interface{}{nil}
		}
		height, _ := strconv.ParseInt(heightStr, 10, 64)
		return []interface{}{&height}

	case "BlockResults":
		heightStr := q.Get("height")
		if heightStr == "" {
			return []interface{}{nil}
		}
		height, _ := strconv.ParseInt(heightStr, 10, 64)
		return []interface{}{&height}

	case "Commit":
		heightStr := q.Get("height")
		if heightStr == "" {
			return []interface{}{nil}
		}
		height, _ := strconv.ParseInt(heightStr, 10, 64)
		return []interface{}{&height}

	case "Genesis", "Health", "NetInfo", "Status":
		return []interface{}{}

	case "Validators":
		heightStr := q.Get("height")
		pageStr := q.Get("page")
		perPageStr := q.Get("per_page")

		var height *int64
		if heightStr != "" {
			h, _ := strconv.ParseInt(heightStr, 10, 64)
			height = &h
		}

		var page *int
		if pageStr != "" {
			p, _ := strconv.Atoi(pageStr)
			page = &p
		}

		var perPage *int
		if perPageStr != "" {
			pp, _ := strconv.Atoi(perPageStr)
			perPage = &pp
		}

		return []interface{}{height, page, perPage}

	case "Simulate":
		tx := q.Get("tx")
		return []interface{}{[]byte(tx)}

	default:
		// For unknown methods, return an empty param array
		return []interface{}{}
	}
}
