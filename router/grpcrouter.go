package router

import (
    "context"

    gt "github.com/go-kit/kit/transport/grpc"
    "github.com/herrjemand/gethGoKitRPCMicroService/proto"
)

type GetSyncResponse struct{
    Status string
    ErrorMessage string
    SyncInfo BlockSyncProgress
}

type GetBlockHashTxsResponse struct{
    Status string
    ErrorMessage string
    Txs []Transaction
}

type GetBlockHashTxsRequest struct{
    BlockHash string
}


func constructGetSyncEndpointGPRC(svc EthService) endpoint.Endpoint {
    return func(_ context.Context, _ interface{}) (interface{}, error) {
        result, err := svc.GetSyncStatus()
        if err != nil {
            return GetSyncResponse{"failed", err.Error()}, nil
        }

        return GetSyncResponse{"ok", "", result.(BlockSyncProgress)}, nil
    }
}

func decodeGetSyncRequest(_ context.Context, _ interface{}) (interface{}, error) {
    return true, nil
}

func encodeGetSyncResponse(_ context.Context, result interface{}) (interface{}, error) {
    res := r.(GetSyncResponse)
    return &proto.GetSyncResponse{
        res.Status,
        res.ErrorMessage,
        res.SyncInfo,
    }, nil
}

func constructGetBlockHashTxsEndpointGRPC(svc EthService) endpoint.Endpoint {
    return func(_ context.Context, request interface{}) (interface{}, error) {
        result, err := svc.GetTransactions(request.(string))
        if err != nil {
            return GetBlockHashTxsResponse{"failed", err.Error()}, nil
        }

        return GetBlockHashTxsResponse{"ok", "", result.(TransactionResultsResponse).Transactions}, nil
    }
}

func decodeCGetBlockHashTxsRequest(_ context.Context, r interface{}) (interface{}, error) {
    req := r.(*proto.GetTxsForBlockHashRequest)
    return req.BlockHash, nil
}

func encodeGetBlockHashTxsResponse(_ context.Context, result interface{}) (interface{}, error) {
    res := r.(GetSyncResponse)
    return &proto.GetTxsForBlockHashResponse{
        res.Status,
        res.ErrorMessage,
        res.SyncInfo,
    }, nil
}

func GetGethGRPCEndpoints(_ context.Context, ethService EthService) EthGRPCServer {
    return &GRPCServer{
        getSync: gt.NewServer(
            constructGetSyncEndpointGPRC(ethService),
            decodeGetSyncRequest,
            encodeGetSyncResponse,
        ),
        createProduct: gt.NewServer(
            constructGetBlockHashTxsEndpointHTTP(ethService),
            decodeCreateProductRequest,
            encodeGetBlockHashTxsResponse,
        ),
    }
}