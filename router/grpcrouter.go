package router

import (
    "context"

    gt "github.com/go-kit/kit/transport/grpc"
    "github.com/herrjemand/gethGoKitRPCMicroService/proto"
    "github.com/go-kit/kit/endpoint"
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
            return GetSyncResponse{"failed", err.Error(), BlockSyncProgress{}}, nil
        }

        return GetSyncResponse{"ok", "", result.(BlockSyncProgress)}, nil
    }
}

func decodeGetSyncRequestGRPC(_ context.Context, _ interface{}) (interface{}, error) {
    return true, nil
}

func encodeGetSyncResponseGPRC(_ context.Context, result interface{}) (interface{}, error) {
    res := result.(GetSyncResponse)
    syncInfo := &proto.SyncInfo{
        StartingBlock: res.SyncInfo.StartingBlock,
        CurrentBlock: res.SyncInfo.CurrentBlock,
        HighestBlock: res.SyncInfo.HighestBlock,
    }

    return &proto.GetSyncResponse{
        Status: res.Status,
        ErrorMessage: res.ErrorMessage,
        SyncInfo: syncInfo,
    }, nil
}

func constructGetBlockHashTxsEndpointGRPC(svc EthService) endpoint.Endpoint {
    return func(_ context.Context, request interface{}) (interface{}, error) {
        result, err := svc.GetTransactions(request.(string))
        if err != nil {
            return GetBlockHashTxsResponse{"failed", err.Error(), []Transaction{}}, nil
        }

        return GetBlockHashTxsResponse{"ok", "", result.(TransactionResultsResponse).Transactions}, nil
    }
}

func decodeGetBlockHashTxsRequestGPRC(_ context.Context, r interface{}) (interface{}, error) {
    req := r.(*proto.GetTxsForBlockHashRequest)
    return req.BlockHash, nil
}

func encodeGetBlockHashTxsResponseGPRC(_ context.Context, result interface{}) (interface{}, error) {
    res := result.(GetBlockHashTxsResponse)
    protoTxs := []*proto.Transaction{}

    for _, transaction := range res.Txs {
        protoTx := &proto.Transaction{
                BlockHash:        transaction.BlockHash,
                BlockNumber:      transaction.BlockNumber,
                From:             transaction.From,
                Gas:              transaction.Gas,
                GasPrice:         transaction.GasPrice,
                Hash:             transaction.Hash,
                Input:            transaction.Input,
                Nonce:            transaction.Nonce,
                To:               transaction.To,
                TransactionIndex: transaction.TransactionIndex,
                Value:            transaction.Value,
                V:                transaction.V,
                R:                transaction.R,
                S:                transaction.S,
        }
        protoTxs = append(protoTxs, protoTx)
    }
    return &proto.GetTxsForBlockHashResponse{
        Status:       res.Status,
        ErrorMessage: res.ErrorMessage,
        Transactions: protoTxs,
    }, nil
}


type GRPCServer struct {
    getSync            gt.Handler
    getTxsForBlockHash gt.Handler
}

func (s *GRPCServer) GetTxsForBlockHash(ctx context.Context, req *proto.GetTxsForBlockHashRequest) (*proto.GetTxsForBlockHashResponse, error) {
    _, resp, err := s.getSync.ServeGRPC(ctx, req)
    if err != nil {
        return nil, err
    }
    return resp.(*proto.GetTxsForBlockHashResponse), nil
}

func (s *GRPCServer) GetSync(ctx context.Context, req *proto.GetSyncRequest) (*proto.GetSyncResponse, error) {
    _, resp, err := s.getTxsForBlockHash.ServeGRPC(ctx, req)
    if err != nil {
        return nil, err
    }
    return resp.(*proto.GetSyncResponse), nil
}

func GetGethGRPCEndpoints(_ context.Context, ethService EthService) proto.EthGRPCServer {
    return &GRPCServer{
        getSync: gt.NewServer(
            constructGetSyncEndpointGPRC(ethService),
            decodeGetSyncRequestGRPC,
            encodeGetSyncResponseGPRC,
        ),
        getTxsForBlockHash: gt.NewServer(
            constructGetBlockHashTxsEndpointGRPC(ethService),
            decodeGetBlockHashTxsRequestGPRC,
            encodeGetBlockHashTxsResponseGPRC,
        ),
    }
}