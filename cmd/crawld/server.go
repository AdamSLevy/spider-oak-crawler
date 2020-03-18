package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	pb "github.com/AdamSLevy/spider-oak-crawler/internal/crawl"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// StartServer launches the gRPC server in a errgroup.Group controlled
// goroutine. The server is gracefully stopped if ctx is cancelled. The
// returned context.Context is cancelled if Serve returns an unexpected error
// and the error is returned by g.Wait().
func StartServer(ctx context.Context, flags *Flags) (*errgroup.Group, error) {

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", flags.Port))
	if err != nil {
		return nil, fmt.Errorf("Failed to listen: %w", err)
	}
	var opts []grpc.ServerOption
	if flags.TLS {
		creds, err := credentials.NewServerTLSFromFile(
			flags.CertFile, flags.KeyFile)
		if err != nil {
			return nil, fmt.Errorf(
				"Failed to generate credentials: %w", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	g, ctx := errgroup.WithContext(ctx)
	srv := crawlServer{Crawler: NewCrawler(flags, ctx)}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterCrawlServer(grpcServer, &srv)

	g.Go(func() error {
		return grpcServer.Serve(lis)
	})
	g.Go(func() error {
		<-ctx.Done()
		fmt.Println("Shutting down...")
		grpcServer.GracefulStop()
		srv.Crawler.Queue.Close()
		return nil
	})

	return g, nil
}

type crawlServer struct {
	pb.UnimplementedCrawlServer
	Crawler *Crawler
}

func (s *crawlServer) List(ctx context.Context, _ *pb.Empty) (*pb.AllHosts, error) {
	fmt.Println("list")
	var res pb.AllHosts
	for u, h := range s.Crawler.Hosts {
		h.mu.RLock()
		defer h.mu.RUnlock()
		res.Hosts = append(res.Hosts, &pb.HostStatus{
			Url:      u,
			Status:   h.Status(),
			SiteTree: &h.SiteTree,
		})
	}
	return &res, nil
}

func (s *crawlServer) Start(ctx context.Context, u *pb.Host) (*pb.Status, error) {
	fmt.Println("start", u.Url)

	if u.Url == "" {
		return nil, fmt.Errorf("empty URL")
	}

	url, err := url.Parse(prependScheme(u.Url))
	if err != nil {
		return nil, err
	}

	status, err := s.Crawler.Start(url)

	return &pb.Status{Status: status}, err
}

// prependScheme returns url with scheme http:// regardless of any scheme
// already specified.
func prependScheme(url string) string {
	parts := strings.SplitN(url, "://", 2)
	if len(parts) == 2 {
		// A scheme was specified, so omit it.
		url = parts[1]
	}
	return fmt.Sprintf("http://%s", url)
}

func (s *crawlServer) Stop(ctx context.Context, u *pb.Host) (*pb.Empty, error) {
	fmt.Println("stop", u.Url)
	if u.Url == "" {
		return nil, fmt.Errorf("empty URL")
	}

	url, err := url.Parse(prependScheme(u.Url))
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, s.Crawler.Stop(url)
}
