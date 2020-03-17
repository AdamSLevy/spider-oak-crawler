package main

import (
	"context"
	"fmt"

	pb "github.com/AdamSLevy/spider-oak-crawler/internal/crawl"
)

func list(client pb.CrawlClient) error {
	var req pb.Empty
	_, err := client.List(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("%v.List(): %w", client, err)
	}
	fmt.Println("list")
	return nil
}
func start(client pb.CrawlClient, url string) error {
	req := pb.URL{Url: url}
	_, err := client.Start(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("%v.Start(%v): %w", client, url, err)
	}
	fmt.Println("start", url)
	return nil
}
func stop(client pb.CrawlClient, url string) error {
	req := pb.URL{Url: url}
	_, err := client.Stop(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("%v.Stop(%v): %w", client, url, err)
	}
	fmt.Println("stop", url)
	return nil
}
