package router

import (
    "encoding/json"
    "log"
    "bytes"
    "io/ioutil"
    "net/http"
    "sync"
    "strconv"
)
var wg sync.WaitGroup
const gethUrl string = "http://localhost:8545"

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

func callGethRPC (rpcStruct EthRPCRequest) (interface{}, error) {
    var jsonData []byte
    jsonData, err := json.Marshal(rpcStruct)

    log.Println("Sending GethRPC request on address \"" + gethUrl + "\" with payload: " + string(jsonData))

    resp, err := http.Post(gethUrl, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, ErrConnectingToGeth
    }

    respBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, ErrReadingGethResponse
    }
    resp.Body.Close()

    log.Println("Got response from GethRPC with JSON payload: ", string(respBytes))
    return respBytes, nil
}

type ethServiceImp struct{}

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
