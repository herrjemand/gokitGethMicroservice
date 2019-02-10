package router

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "github.com/go-kit/kit/endpoint"
    "github.com/gorilla/mux"
    httptransport "github.com/go-kit/kit/transport/http"
)

func generateErrorResponse(errorMessage string) ([]byte) {
    return []byte("{\"status\":\"error\", \"errorMessage\":\"" + errorMessage + "\"}")
}

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

func GenerateHTTPRouter() interface{} {
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

    return router
}