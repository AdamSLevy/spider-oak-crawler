package main

import (
	"context"
	"fmt"
	"net"

	pb "github.com/AdamSLevy/spider-oak-crawler/internal/crawl"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func StartServer(ctx context.Context,
	flags *Flags) (*errgroup.Group, context.Context, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", flags.Port))
	if err != nil {
		return nil, nil,
			fmt.Errorf("Failed to listen: %w", err)
	}
	var opts []grpc.ServerOption
	if flags.TLS {
		creds, err := credentials.NewServerTLSFromFile(
			flags.CertFile, flags.KeyFile)
		if err != nil {
			return nil, nil,
				fmt.Errorf("Failed to generate credentials: %w", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterCrawlServer(grpcServer, &crawlServer{})
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return grpcServer.Serve(lis)
	})
	g.Go(func() error {
		<-ctx.Done()
		fmt.Println("Shutting down...")
		grpcServer.GracefulStop()
		return nil
	})
	return g, ctx, nil
}

type crawlServer struct {
	pb.UnimplementedCrawlServer
}

func (s *crawlServer) List(ctx context.Context, _ *pb.Empty) (*pb.ListResponse, error) {
	fmt.Println("list")
	var res pb.ListResponse
	return &res, nil
}
func (s *crawlServer) Start(ctx context.Context, url *pb.URL) (*pb.Status, error) {
	fmt.Println("start", url.Url)
	var res pb.Status
	return &res, nil
}
func (s *crawlServer) Stop(ctx context.Context, url *pb.URL) (*pb.Status, error) {
	fmt.Println("stop", url.Url)
	var res pb.Status
	return &res, nil
}
