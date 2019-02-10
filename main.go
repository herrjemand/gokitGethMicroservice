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
    "strconv"
    "github.com/go-kit/kit/endpoint"
    "github.com/gorilla/mux"
    httptransport "github.com/go-kit/kit/transport/http"
)

const gethUrl string = "http://localhost:8545"
const serverAddress string = "127.0.0.1"
const serverPort string = "8000"


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
    Nonce string `json:"nonce"`
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

type BlockSyncProgress struct {
    StartingBlock string `json:"startingBlock"`
    CurrentBlock string `json:"currentBlock"`
    HighestBlock string `json:"highestBlock"`
}

type GetSyncResult struct {
    Jsonrpc string `json:"jsonrpc"`
    Result BlockSyncProgress `json:"result"`
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

func generateErrorResponse(errorMessage string) ([]byte) {
    return []byte("{\"status\":\"error\", \"errorMessage\":\"" + errorMessage + "\"}")
}

func callGethRPC (rpcStruct EthRPCRequest) (interface{}, error) {
    var jsonData []byte
    jsonData, err := json.Marshal(rpcStruct)

    log.Println("Sending GethRPC request on address \"" + gethUrl + "\" with payload: " + string(jsonData))

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

    fmt.Println("GOt response from GethRPC with JSON payload: ", string(respBytes))
    return respBytes, nil
}

// ethServiceImp is a concrete implementation of StringService
type ethServiceImp struct{}

var ErrConnectingToGeth = errors.New("Error connecting to geth!")
var ErrReadingGethResponse = errors.New("Error reading response from Geth!")
var ErrParsingJSON = errors.New("Error while parsing JSON!")
var ErrEncodingJSON = errors.New("Error while encoding JSON!")
var ErrParsingInt = errors.New("Error while parsing Int!")
var ErrNullResult = errors.New("Error! Geth returned NULL result!")

func (ethServiceImp) GetSyncStatus() (interface{}, error) {
    rpcReq := EthRPCRequest{}

    rpcReq.constructGetSyncingRequest()

    resp, err := callGethRPC(rpcReq)
    if err != nil {
        return nil, err
    }

    var getSyncResp GetSyncResult;
    err = json.Unmarshal(resp.([]byte), &getSyncResp)
    if err != nil {
        return nil, ErrParsingJSON
    }

    
    return getSyncResp.Result, nil
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

    if txCountStruct.Result == "" {
        return nil, ErrNullResult
    }
    
    var txCount int64;
    txCount, err = strconv.ParseInt(txCountStruct.Result, 0, 16)
    if err != nil {
        return nil, ErrParsingInt
    }

    txChannel := make(chan TxChannelResult, txCount)
    for i := 0; i < int(txCount); i++ {
        inji := i
        wg.Add(1)
        go func() {
            defer wg.Done()
            rpcReq := EthRPCRequest{}

            rpcReq.constructGetTransactionByBlockHashAndIndexRequest(blockHash, int64(inji))
            txBlockCountResp, err := callGethRPC(rpcReq)
            if err != nil {
                txChannel <- TxChannelResult{Transaction{}, err}
            }

            var txBlockCountRespStruct TransactionResult;
            err = json.Unmarshal(txBlockCountResp.([]byte), &txBlockCountRespStruct)
            if err != nil {
                txChannel <- TxChannelResult{Transaction{}, err}
                return
            }

            if txBlockCountRespStruct.Result == (Transaction{}) {
                txChannel <- TxChannelResult{Transaction{}, ErrNullResult}
                return
            }

            txChannel <- TxChannelResult{txBlockCountRespStruct.Result, nil}
        }()
    }

    wg.Wait()
    close(txChannel)

    var txs []Transaction = []Transaction{}
    for txResult := range txChannel {
        if txResult.Error != nil {
            return nil, txResult.Error
        }
        txs = append(txs, txResult.Tx)
    }

    txResponse := TransactionResultsResponse{txs}

    return txResponse, nil
}

type addressResponse struct {
    A string `json:"address"`
}
/* ----- INTERFACE IMPLEMENTORS ENDS ----- */

func constructGetBlockHashTxsEndpoint(svc EthService) endpoint.Endpoint {
    return func(_ context.Context, request interface{}) (interface{}, error) {
        result, err := svc.GetTransactions(request.(string))
        if err != nil {
            return generateErrorResponse(err.Error()), nil
        }

        var jsonData []byte
        jsonData, err = json.Marshal(result.(TransactionResultsResponse))
        if err != nil {
            return generateErrorResponse(ErrEncodingJSON.Error()), nil
        }

        return jsonData, nil
    }
}

func decodeBlockHashTxsRequest(_ context.Context, r *http.Request) (interface{}, error){
    vars := mux.Vars(r)
    log.Println("Receiving GetBlockHashTxs Request for Hash: " + vars["blockHash"])
    return vars["blockHash"], nil
}

func decodeBlockHashTxsResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
    log.Println("Sending GetBlockHashTxs Response: " + string(response.([]byte)))
    _, err := w.Write(response.([]byte))
    return err
}

func constructGetSyncStatusEndpoint(svc EthService) endpoint.Endpoint {
    return func(_ context.Context, request interface{}) (interface{}, error) {
        result, err := svc.GetSyncStatus()
        if err != nil {
            return generateErrorResponse(err.Error()), nil
        }

        var jsonData []byte
        jsonData, err = json.Marshal(result.(BlockSyncProgress))
        if err != nil {
            return generateErrorResponse(ErrEncodingJSON.Error()), nil
        }

        return jsonData, nil
    }
}

func decodeGetSyncRequest(_ context.Context, r *http.Request) (interface{}, error){
    log.Println("Receiving GetSync Request")
    return true, nil
}

func encodeGetSyncResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
    log.Println("Sending GetSync Response: " + string(response.([]byte)))
    _, err := w.Write(response.([]byte))
    return err
}

// Transports expose the service to the network. In this first example we utilize JSON over HTTP.
func main() {
    svc := ethServiceImp{}

    addressHandler := httptransport.NewServer(
        constructGetBlockHashTxsEndpoint(svc),
        decodeBlockHashTxsRequest,
        decodeBlockHashTxsResponse,
    )

    getSyncHandler := httptransport.NewServer(
        constructGetSyncStatusEndpoint(svc),
        decodeGetSyncRequest,
        encodeGetSyncResponse,
    )

    router := mux.NewRouter()
    router.Methods("GET").PathPrefix("/getBlockHashTransactions/{blockHash}").Handler(addressHandler)
    router.Methods("GET").PathPrefix("/getSyncStatus/").Handler(getSyncHandler)
        
    server := &http.Server{
        Handler:      router,
        Addr:         serverAddress + ":" + serverPort,

        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    log.Println("Starting HTTP server at " + serverAddress + ":" + serverPort + "...")
    log.Fatal(server.ListenAndServe())
}
