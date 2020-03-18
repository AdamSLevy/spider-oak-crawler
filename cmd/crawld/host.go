package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	pb "github.com/AdamSLevy/spider-oak-crawler/internal/crawl"
	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
)

// TODO: handle permanent redirects...

type Host struct {
	// Threadsafe
	*fetchbot.ResponseMatcher
	*fetchbot.Queue

	// Set once by Crawler.Start
	*url.URL

	mu sync.RWMutex // Protect the following
	pb.SiteTree
	dup      map[string]struct{}
	enqueued int32 // # of enqueued links, >0 while crawl in progress

	ctx        context.Context
	stop       func()
	resumeGET  []string
	resumeHEAD []string
}

func (h *Host) Resume(ctx context.Context) {

	h.mu.Lock()
	defer h.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	h.ctx = ctx
	h.stop = cancel

	// TODO: better error handling
	for _, url := range h.resumeGET {
		if _, err := h.Queue.SendStringGet(url); err != nil {
			fmt.Printf("error: enqueue GET %s - %s\n", url, err)
		}
	}
	for _, url := range h.resumeHEAD {
		if _, err := h.Queue.SendStringHead(url); err != nil {
			fmt.Printf("error: enqueue HEAD %s - %s\n", url, err)
		}
	}
}

func (h *Host) Status() pb.CrawlStatus {

	if h.ctx.Err() != nil {
		return pb.CrawlStatus_STOPPED
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.enqueued > 0 {
		return pb.CrawlStatus_CRAWLING
	}

	return pb.CrawlStatus_FINISHED
}

func (h *Host) HEADToGETHandler(ctx *fetchbot.Context, res *http.Response, _ error) {

	if h.ctx.Err() != nil {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.resumeGET = append(h.resumeGET, ctx.Cmd.URL().String())
		return
	}

	if _, err := ctx.Q.SendStringGet(ctx.Cmd.URL().String()); err != nil {
		fmt.Printf("error: enqueue HEAD %s - %s\n", ctx.Cmd.URL(), err)
	}
}

func (h *Host) enqueueLinks(ctx *fetchbot.Context, doc *goquery.Document) {

	h.mu.Lock()
	defer h.mu.Unlock()

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {

		val, _ := s.Attr("href")
		// Resolve address
		u, err := ctx.Cmd.URL().Parse(val)
		if err != nil {
			fmt.Printf("error: resolve URL %s - %s\n", val, err)
			return
		}
		// TODO: Normalize URL

		if u.Host != h.URL.Host {
			// Ignore external links.
			return
		}
		if _, visited := h.dup[u.String()]; visited {
			// Ignore visited links.
			return
		}

		h.enqueued++
		h.dup[u.String()] = struct{}{}
		h.insertSiteTree(u)

		if h.ctx.Err() != nil {
			h.resumeHEAD = append(h.resumeHEAD, ctx.Cmd.URL().String())
			return
		}

		if _, err := ctx.Q.SendStringHead(u.String()); err != nil {
			fmt.Printf("error: enqueue head %s - %s\n", u, err)
			return
		}
	})

	h.enqueued--
}

// TODO: This is not the most efficient way to insert into the site tree.  It
// would be better if the parent site tree was carried with the context of the
// response. This can be achieved by implementing a custom fetchbot.Cmd. It's a
// little tricky though if you want to maintain the threadsafety of the
// handlers.
func (h *Host) insertSiteTree(u *url.URL) {
	parts := strings.Split(u.Path, "/")
	st := &h.SiteTree
	for _, part := range parts {
		if st.Children == nil {
			st.Children = make(map[string]*pb.SiteTree)
		}
		child, ok := st.Children[part]
		if !ok {
			child = &pb.SiteTree{}
			st.Children[part] = child
		}
		st = child
	}
}
