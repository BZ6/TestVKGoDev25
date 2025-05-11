package main

import (
    "context"
    "log"
    "net"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/grpc/reflection"

    pb "GoEventHub/proto"
    "GoEventHub/subpub"
)

type server struct {
    pb.UnimplementedPubSubServer
    subPub subpub.SubPub
}

func (s *server) Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) error {
	log.Printf("New subscription for key: %s", req.Key)
    sub, err := s.subPub.Subscribe(req.Key, func(msg interface{}) {
        if event, ok := msg.(string); ok {
            stream.Send(&pb.Event{Data: event})
        }
    })
    if err != nil {
		log.Printf("Failed to subscribe: %v", err)
        return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
    }
    defer sub.Unsubscribe()

    <-stream.Context().Done()
	log.Printf("Subscription for key %s closed", req.Key)
    return stream.Context().Err()
}

func (s *server) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	log.Printf("Publishing event to key: %s, data: %s", req.Key, req.Data)
    err := s.subPub.Publish(req.Key, req.Data)
    if err != nil {
		log.Printf("Failed to publish: %v", err)
        return nil, status.Errorf(codes.Internal, "failed to publish: %v", err)
    }
    return &emptypb.Empty{}, nil
}

func main() {
	config := loadConfig()

    subPub := subpub.NewSubPub()
    defer subPub.Close(context.Background())

    grpcServer := grpc.NewServer()
    pb.RegisterPubSubServer(grpcServer, &server{subPub: subPub})

	reflection.Register(grpcServer)

    listener, err := net.Listen("tcp", ":" + config.Server.Port)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    log.Println("Server is running on port :", config.Server.Port)
    if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
