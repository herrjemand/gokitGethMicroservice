package router

import (
    "context"

    gt "github.com/go-kit/kit/transport/grpc"

    "grpc-go-kit-example/product-service/pb"
)


// func EncodeProductResponse(_ context.Context, r interface{}) (interface{}, error) {
//     res := r.(ProductResponse)
//     return &pb.ProductResponse{
//         res.ID,
//         res.Name,
//     }, nil
// }

// func DecodeGetProductRequest(_ context.Context, r interface{}) (interface{}, error) {
//     req := r.(*pb.GetProductRequest)
//     return GetProductRequest{req.Id}, nil
// }

// func DecodeCreateProductRequest(_ context.Context, r interface{}) (interface{}, error) {
//     req := r.(*pb.CreateProductRequest)
//     return CreateProductRequest{req.Name}, nil
// }

// type GRPCServer struct {
//     getSync       gt.Handler
//     createProduct gt.Handler
// }

// func (s *GRPCServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
//     _, resp, err := s.getProduct.ServeGRPC(ctx, req)
//     if err != nil {
//         return nil, err
//     }
//     return resp.(*pb.ProductResponse), nil
// }

// func (s *GRPCServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
//     _, resp, err := s.createProduct.ServeGRPC(ctx, req)
//     if err != nil {
//         return nil, err
//     }
//     return resp.(*pb.ProductResponse), nil
// }

func constructGetSyncEndpointGPRC(svc EthService) endpoint.Endpoint {
    return func(_ context.Context, request interface{}) (interface{}, error) {
        gr := req.(GetProductRequest)
        product, _ := svc.GetProduct(gr.ID)
        return ProductResponse{product.ID, product.Name}, nil
    }
}

func constructGetSyncStatusEndpointHTTP(svc EthService) endpoint.Endpoint {
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

func GetGethGRPCEndpoints(_ context.Context, ethService EthService) EthGRPCServer {
    return &GRPCServer{
        getSync: gt.NewServer(
            constructGetSyncEndpointGPRC(ethService),
            DecodeGetProductRequest,
            EncodeProductResponse,
        ),
        createProduct: gt.NewServer(
            product.MakeCreateProductEndpoint(ethService),
            DecodeCreateProductRequest,
            EncodeProductResponse,
        ),
    }
}