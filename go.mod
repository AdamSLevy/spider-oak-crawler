module github.com/AdamSLevy/spider-oak-crawler

go 1.14

require (
	github.com/AdamSLevy/flagbind v0.0.0-20200317230050-c9dd74cc7efd
	github.com/PuerkitoBio/fetchbot v1.2.0
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/golang/protobuf v1.3.5
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/grpc v1.28.0
)

// Improved fetchbot.Mux handler matching.
replace github.com/PuerkitoBio/fetchbot v1.2.0 => github.com/AdamSLevy/fetchbot v1.2.1-0.20200320041741-10eaeff1774b
