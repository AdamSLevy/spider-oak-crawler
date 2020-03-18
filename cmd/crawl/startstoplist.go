package main

import (
	"context"
	"fmt"

	pb "github.com/AdamSLevy/spider-oak-crawler/internal/crawl"
)

func list(client pb.CrawlClient) error {
	var req pb.Empty
	siteTree, err := client.List(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("%v.List(): %w", client, err)
	}
	// TODO: Improve SiteTree output
	fmt.Println("list", siteTree)
	return nil
}
func start(client pb.CrawlClient, url string) error {
	req := pb.Host{Url: url}
	status, err := client.Start(context.Background(), &req)
	if err != nil {
		return err
	}
	fmt.Println("start", status)
	return nil
}
func stop(client pb.CrawlClient, url string) error {
	req := pb.Host{Url: url}
	_, err := client.Stop(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("%v.Stop(%v): %w", client, url, err)
	}
	fmt.Println("stop", url)
	return nil
}
