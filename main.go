package main

import (
    "context"
    "encoding/json"
    "errors"
    "log"
    "time"
    "fmt"
    "bytes"
    "io/ioutil"
    "net/http"
    // "strings"
    // "strconv"
    "github.com/go-kit/kit/endpoint"
    "github.com/gorilla/mux"
    httptransport "github.com/go-kit/kit/transport/http"
)

const gethUrl string = "http://localhost:8545"
// StringService provides operations on strings.
type EthService interface {
    // Uppercase(string) (string, error)
    // Count(string) int
    GetSyncStatus() (interface{}, error)
}

/* ----- INTERFACE IMPLEMENTORS ----- */

type EthRPCRequest struct {
    Jsonrpc string `json:"jsonrpc"`
    Method string `json:"method"`
    Params []string `json:"params"`
    Id int32 `json:"id"`
}

func (ethreq *EthRPCRequest) constructGetSyncingRequest() {
    ethreq.Jsonrpc = "2.0"
    ethreq.Method = "eth_syncing"
    ethreq.Params = []string{}
    ethreq.Id = 0x01
}

// func (ethreq *EthRPCRequest) constructGetBlockTransactionCountByHashRequest(blockHash string) {
//     ethreq.Jsonrpc = "2.0"
//     ethreq.Method = ""
//     ethreq.Params = []string{blockHash}
//     ethreq.Id = 0x01    
// }

// func (ethreq *EthRPCRequest) constructGetTransactionByBlockHashAndIndexRequest(blockHash string, transactionIndex int) {
//     trInx := '0x' + strconv.FormatInt(transactionIndex, 16)
//     ethreq.Jsonrpc = "2.0"
//     ethreq.Method = ""
//     ethreq.Params = []string{blockHash, ethreq}
//     ethreq.Id = 0x01
// }

// ethServiceImp is a concrete implementation of StringService
type ethServiceImp struct{}

var ErrConnectingToGeth = errors.New("Error connecting to geth!")
var ErrReadingGethResponse = errors.New("Error reading response from Geth!")

func (ethServiceImp) GetSyncStatus() (interface{}, error) {
    rpcReq := EthRPCRequest{}
    fmt.Println(rpcReq)

    rpcReq.constructGetSyncingRequest()

    var jsonData []byte
    jsonData, err := json.Marshal(rpcReq)

    fmt.Println(rpcReq)
    fmt.Println(string(jsonData))
    resp, err := http.Post(gethUrl, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, ErrConnectingToGeth
    }

    respBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, ErrReadingGethResponse
    }

    // fmt.Println("response Body:", resp.RawText())
    return respBytes, nil
}


type addressResponse struct {
    A string `json:"address"`
}



/* ----- INTERFACE IMPLEMENTORS ENDS ----- */

// func makeCountEndpoint(svc StringService) endpoint.Endpoint {
//     return func(_ context.Context, request interface{}) (interface{}, error) {
//         req := request.(countRequest)
//         v := svc.Count(req.S)
//         return countResponse{v}, nil
//     }
// }


// func getAddressBack(svc EthService) endpoint.Endpoint {
//     return func(_ context.Context, request interface{}) (interface{}, error) {
//         req := request.(string)
//         // v := svc.GerAddressBack(req)
//         return addressResponse{v}, nil
//     }
// }

func getSyncStatusConstr(svc EthService) endpoint.Endpoint {
    return func(_ context.Context, request interface{}) (interface{}, error) {
        resp, err := svc.GetSyncStatus()
        if err != nil {
            return nil, err
        }

        return resp, nil
    }
}

// Transports expose the service to the network. In this first example we utilize JSON over HTTP.
func main() {
    svc := ethServiceImp{}

    // addressHandler := httptransport.NewServer(
    //     getAddressBack(svc),
    //     decodeAddressRequest,
    //     encodeResponse,
    // )

    getSyncHandler := httptransport.NewServer(
        getSyncStatusConstr(svc),
        decodeGetSyncRequest,
        encodeResponseGetSync,
    )

    router := mux.NewRouter()
    // router.Methods("GET").PathPrefix("/getAddressTransactions/{address}").Handler(addressHandler)
    router.Methods("GET").PathPrefix("/getSyncStatus/").Handler(getSyncHandler)
        

    server := &http.Server{
        Handler:      router,
        Addr:         "127.0.0.1:8000",

        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    log.Fatal(server.ListenAndServe())
}

func decodeAddressRequest(_ context.Context, r *http.Request) (interface{}, error){
    vars := mux.Vars(r)
    fmt.Println("Category: %v\n", vars["address"])
    return vars["address"], nil
}

func decodeGetSyncRequest(_ context.Context, r *http.Request) (interface{}, error){
    return true, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}


func encodeResponseGetSync(_ context.Context, w http.ResponseWriter, response interface{}) error {
    _, err := w.Write(response.([]byte))
    return err
}
