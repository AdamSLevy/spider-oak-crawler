# Spider Oak Programming Take Home

Hello Spider Oak Go devs!

I had fun writing this program. I had not used gRPC or protobuf before, so that
was a small learning curve to get through. There are a number of rough edges
and I have hardly had a chance to thoroughly road test it, so I'm sure there
are bugs in the crawler. However the core requirements are complete, and the
groundwork is laid to make this a fairly robust crawler service.

### TODO && Potential Improvements

Below are a list of things I would improve if I had more time. I may still get
to some this week.

- Normalize URLs using github.com/PuerkitoBio/purell. This is very important
  but trivial to add. I have marked all the places it needs to occur.
- Expose URL Normalization flags. There are a number of ways to normalize URLs.
- Improve the printing of SiteTrees for `crawl -list`. The output is really
  ugly, hard to validate at all right now. I hope this can be forgiven with the
knowledge I'd never leave it like that.
- Add a debug logger so that the `-debug` flag is meaningful.
- Ensure first query is successful before returning the Start RPC.
- Handle permanent redirects on the first URL so that sites that redirect to a
  subdomain are respected. example.com -> www.example.com
- Modify fetchbot to allow multiple concurrent queries on a single host. This
  violates robots.txt politeness policies.
- Improve SiteTree insertion efficiency. It's a bit of a kludge right now but
  it works.
- Add a SQLite database so that SiteTree and crawl progress can be persistent.
- Allow host specific crawling settings to be passed through the Start RPC
  call.
- Add a Status RPC call that doesn't print the SiteTree but shows what the
  crawler is working on.
- Allow the List RPC to be specific to a host.

### Build
```
$ make
go build ./cmd/crawl
go build ./cmd/crawld
```

#### Generate gRPC/protobuf
```
$ make generate
go generate ./internal/crawl
```

### External Dependencies

#### github.com/AdamSLevy/flagbind

This is a package of my own that parses flags from struct tags. It makes
setting up and maintaining flags very easy. This is still in development and
I'm finding places it chafes and needs some improvements, but I enjoy using it
in its current state.

#### github.com/PuerkitoBio/fetchbot

This maintains a queue of requests and allows response handlers to be
registered. It respects robots.txt policies, which also makes it a bit slow. It
launches one goroutine per host, so scanning a single site is not the most
performant. I forked this during this project to make some minor improvements,
see go.mod for details.

One possible way to speed up site crawling is to add more query runners per
host inside the fetchbot package.

Use `./crawld -disable-politeness -crawl-delay 0s` for fastest possible
crawling.

#### github.com/PuerkitoBio/goquery

I pulled this straight from the fetchbot example. It makes finding all links
very straightforward.

#### github.com/PuerkitoBio/purell

I plan to use this for URL normalization.
