package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	pb "github.com/AdamSLevy/spider-oak-crawler/internal/crawl"
	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	// Threadsafe
	*fetchbot.Mux
	*fetchbot.Queue

	// Set once by NewCrawler
	ctx context.Context

	mu    sync.RWMutex // Protects the following
	Hosts map[string]*Host
}

func NewCrawler(flags *Flags, ctx context.Context) *Crawler {
	m := fetchbot.NewMux()
	f := flags.Fetcher
	f.Handler = fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {
			if err == nil {
				fmt.Printf("%v %v\n", ctx.Cmd.Method(), ctx.Cmd.URL())
			}
			m.Handle(ctx, res, err)
		})
	q := f.Start()

	c := Crawler{Mux: m, Queue: q, Hosts: make(map[string]*Host), ctx: ctx}

	// Handle all errors.
	m.HandleErrors(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {
			fmt.Printf("[ERR] %s %s - %s\n",
				ctx.Cmd.Method(), ctx.Cmd.URL(), err)
		}))

	// Handle all GET requests. Parse the body and enqueue all links as
	// HEAD requests. The Hosts ensure that only HEAD responses with
	// content type text/html are enqueued as GET requests.
	m.Response().Method("GET").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, _ error) {
			// Process the body to find the links
			doc, err := goquery.NewDocumentFromResponse(res)
			if err != nil {
				fmt.Printf("[ERR] %s %s - %s\n",
					ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				return
			}

			c.mu.RLock()
			h := c.Hosts[ctx.Cmd.URL().Host]
			c.mu.RUnlock()

			// Enqueue all links as HEAD requests
			h.enqueueLinks(ctx, doc)
		}))

	return &c
}

func (c *Crawler) Start(url *url.URL) (pb.CrawlStatus, error) {

	// TODO: Normalize URL

	c.mu.Lock()
	defer c.mu.Unlock()

	if h, ok := c.Hosts[url.Host]; ok {
		status := h.Status()
		// TODO: should these be errors to the client?
		switch status {
		case pb.CrawlStatus_CRAWLING:
			return status, fmt.Errorf(
				"already started crawling: %v", url.Host)
		case pb.CrawlStatus_FINISHED:
			return status, fmt.Errorf(
				"already finished crawling: %v", url.Host)
		}

		h.Resume(c.ctx)
		return pb.CrawlStatus_CRAWLING, nil

	}
	ctx, cancel := context.WithCancel(c.ctx)

	h := Host{
		URL:      url,
		Queue:    c.Queue,
		SiteTree: pb.SiteTree{Children: make(map[string]*pb.SiteTree)},
		dup:      make(map[string]struct{}),
		ctx:      ctx,
		stop:     cancel,
	}

	// Match HEAD responses from this Host that are text/html. The handler
	// then enqueues them as GET requests. This ensures we only ever GET
	// actual HTML pages, and not any other content type.
	//
	// Although we attempt to never queue up external links, a redirect
	// could bring us to an external site, so the Host filter is still
	// useful here.
	h.ResponseMatcher = c.Mux.Response().
		Method("HEAD").
		Host(url.Host).
		ContentType("text/html").
		Handler(fetchbot.HandlerFunc(h.HEADToGETHandler))

	// If the HEAD response isn't text/html, just decrement the enqueued
	// count.
	c.Mux.Response().
		Method("HEAD").
		Host(url.Host).
		Handler(fetchbot.HandlerFunc(
			func(*fetchbot.Context, *http.Response, error) {
				h.mu.Lock()
				defer h.mu.Unlock()
				h.enqueued--
			}))

	c.Hosts[url.Host] = &h

	// TODO: Ensure this first query is successful before returning.
	if _, err := c.SendStringHead(url.String()); err != nil {
		return pb.CrawlStatus_UNKNOWN, err
	}

	return pb.CrawlStatus_CRAWLING, nil
}

// Stop crawling the given URL.
func (c *Crawler) Stop(url *url.URL) error {

	// TODO: Normalize URL

	c.mu.Lock()
	defer c.mu.Unlock()

	h, ok := c.Hosts[url.Host]
	if !ok {
		return fmt.Errorf("not currently crawling: %v", url.Host)
	}

	switch h.Status() {
	case pb.CrawlStatus_STOPPED:
		return fmt.Errorf("already stopped crawling: %v", url.Host)
	case pb.CrawlStatus_FINISHED:
		return fmt.Errorf("already finished crawling: %v", url.Host)
	}

	h.stop()

	return nil
}
