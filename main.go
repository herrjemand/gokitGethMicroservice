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
    "sync"
    // "strings"
    "strconv"
    "github.com/go-kit/kit/endpoint"
    "github.com/gorilla/mux"
    httptransport "github.com/go-kit/kit/transport/http"
)

const gethUrl string = "http://localhost:8545"
var wg sync.WaitGroup

type EthService interface {
    GetSyncStatus() (interface{}, error)
    GetTransactions(string) (interface{}, error)
}

/* ----- INTERFACE IMPLEMENTORS ----- */
type EthRPCRequest struct {
    Jsonrpc string `json:"jsonrpc"`
    Method string `json:"method"`
    Params []string `json:"params"`
    Id int32 `json:"id"`
}

type BlockTransactionCount struct {
    Jsonrpc string `json:"jsonrpc"`
    Result string `json:"result"`
    Id int32 `json:"id"`
}

type Transaction struct {
    BlockHash string `json:"blockHash"`
    BlockNumber string `json:"blockNumber"`
    From string `json:"from"`
    Gas string `json:"gas"`
    GasPrice string `json:"gasPrice"`
    Hash string `json:"hash"`
    Input string `json:"input"`
    nonce string `json:"nonce"`
    To string `json:"to"`
    TransactionIndex string `json:"transactionIndex"`
    Value string `json:"value"`
    V string `json:"v"`
    R string `json:"r"`
    S string `json:"s"`
}

type TransactionResult struct {
    Jsonrpc string `json:"jsonrpc"`
    Result Transaction `json:"result"`
    Id int32 `json:"id"`
}

type TransactionResultsResponse struct {
    Transactions []Transaction `json:"transactions"`
}

type TxChannelResult struct {
    Tx Transaction
    Error error
}


func (ethreq *EthRPCRequest) constructGetSyncingRequest() {
    ethreq.Jsonrpc = "2.0"
    ethreq.Method = "eth_syncing"
    ethreq.Params = []string{}
    ethreq.Id = 0x01
}

func (ethreq *EthRPCRequest) constructGetBlockTransactionCountByHashRequest(blockHash string) {
    ethreq.Jsonrpc = "2.0"
    ethreq.Method = "eth_getBlockTransactionCountByHash"
    ethreq.Params = []string{blockHash}
    ethreq.Id = 0x01    
}

func (ethreq *EthRPCRequest) constructGetTransactionByBlockHashAndIndexRequest(blockHash string, transactionIndex int64) {
    var trInx string = "0x" + strconv.FormatInt(transactionIndex, 16)
    ethreq.Jsonrpc = "2.0"
    ethreq.Method = "eth_getTransactionByBlockHashAndIndex"
    ethreq.Params = []string{blockHash, trInx}
    ethreq.Id = 0x01
}

func callGethRPC (rpcStruct EthRPCRequest) (interface{}, error) {
    var jsonData []byte
    jsonData, err := json.Marshal(rpcStruct)

    fmt.Println("JSON Request", string(jsonData))
    resp, err := http.Post(gethUrl, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, ErrConnectingToGeth
    }

    respBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, ErrReadingGethResponse
    }
    resp.Body.Close()

    fmt.Println("JSON Response", string(respBytes))
    return respBytes, nil
}

// ethServiceImp is a concrete implementation of StringService
type ethServiceImp struct{}

var ErrConnectingToGeth = errors.New("Error connecting to geth!")
var ErrReadingGethResponse = errors.New("Error reading response from Geth!")
var ErrParsingJSON = errors.New("Error while parsing JSON!")
var ErrParsingInt = errors.New("Error while parsing JSON!")

func (ethServiceImp) GetSyncStatus() (interface{}, error) {
    rpcReq := EthRPCRequest{}

    rpcReq.constructGetSyncingRequest()

    resp, err := callGethRPC(rpcReq)
    if err != nil {
        return nil, err
    }
    
    return resp, nil
}

func (ethServiceImp) GetTransactions(blockHash string) (interface{}, error) {
    rpcReq := EthRPCRequest{}

    // Getting transaction count
    rpcReq.constructGetBlockTransactionCountByHashRequest(blockHash)
    transactionCountResp, err := callGethRPC(rpcReq)
    if err != nil {
        return nil, err
    }

    var txCountStruct BlockTransactionCount;
    err = json.Unmarshal(transactionCountResp.([]byte), &txCountStruct)
    if err != nil {
        return nil, ErrParsingJSON
    }

    fmt.Println(txCountStruct.Result)
    
    txCount, _ := strconv.ParseInt(txCountStruct.Result, 0, 16)
    if err != nil {
        return nil, ErrParsingInt
    }

    txChannel := make(chan TxChannelResult)
    for i := 0; i < int(txCount); i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            rpcReq := EthRPCRequest{}

            // Getting transaction count
            rpcReq.constructGetTransactionByBlockHashAndIndexRequest(blockHash, int64(i))
            txBlockCountResp, err := callGethRPC(rpcReq)
            if err != nil {
                txChannel <- TxChannelResult{Transaction{}, err}
            }

            var txBlockCountRespStruct TransactionResult;
            err = json.Unmarshal(txBlockCountResp.([]byte), &txBlockCountRespStruct)
            if err != nil {
                txChannel <- TxChannelResult{Transaction{}, err}
            }

            txChannel <- TxChannelResult{txBlockCountRespStruct.Result, nil}
        }()
    }

    var txs []Transaction = []Transaction{}
    for txResult := range txChannel {
        if txResult.Error != nil {
            return nil, txResult.Error
        }
        txs = append(txs, txResult.Tx)
    }

    wg.Wait()
    close(txChannel)

    txResponse := TransactionResultsResponse{txs}

    var jsonData []byte
    jsonData, err = json.Marshal(txResponse)
    if err != nil {
        return nil, ErrParsingJSON
    }

    return jsonData, nil
}

type addressResponse struct {
    A string `json:"address"`
}



/* ----- INTERFACE IMPLEMENTORS ENDS ----- */


func getSyncStatusConstr(svc EthService) endpoint.Endpoint {
    return func(_ context.Context, request interface{}) (interface{}, error) {
        resp, err := svc.GetSyncStatus()
        if err != nil {
            return nil, err
        }

        return resp, nil
    }
}


func getTransactionsConstr(svc EthService) endpoint.Endpoint {
    return func(_ context.Context, request interface{}) (interface{}, error) {
        resp, err := svc.GetTransactions(request.(string))
        if err != nil {
            return nil, err
        }

        return resp, nil
    }
}

// Transports expose the service to the network. In this first example we utilize JSON over HTTP.
func main() {
    svc := ethServiceImp{}

    addressHandler := httptransport.NewServer(
        getTransactionsConstr(svc),
        decodeAddressRequest,
        encodeTransactionsResponse,
    )

    getSyncHandler := httptransport.NewServer(
        getSyncStatusConstr(svc),
        decodeGetSyncRequest,
        encodeResponseGetSync,
    )

    router := mux.NewRouter()
    router.Methods("GET").PathPrefix("/getBlockAddressTransactions/{blockHash}").Handler(addressHandler)
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
    fmt.Println("BlockHash: ", vars["blockHash"])
    return vars["blockHash"], nil
}

func decodeGetSyncRequest(_ context.Context, r *http.Request) (interface{}, error){
    return true, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
    return json.NewEncoder(w).Encode(response)
}

func encodeTransactionsResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
    _, err := w.Write(response.([]byte))
    return err
}

func encodeResponseGetSync(_ context.Context, w http.ResponseWriter, response interface{}) error {
    _, err := w.Write(response.([]byte))
    return err
}
