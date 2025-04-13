package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	pb "github.com/learies/goShortener/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

// extractShortURL extracts the short URL identifier from the full URL
func extractShortURL(fullURL string) string {
	parts := strings.Split(fullURL, "/")
	return parts[len(parts)-1]
}

func main() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewURLShortenerClient(conn)

	// Set timeout for RPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Test CreateShortURL
	createResp, err := client.CreateShortURL(ctx, &pb.CreateShortURLRequest{
		Url: "https://example.com",
	})
	if err != nil {
		log.Fatalf("could not create short URL: %v", err)
	}
	fmt.Printf("Created short URL: %s\n", createResp.Result)

	// Test GetOriginalURL
	shortURL := extractShortURL(createResp.Result)
	getResp, err := client.GetOriginalURL(ctx, &pb.GetOriginalURLRequest{
		ShortUrl: shortURL,
	})
	if err != nil {
		log.Fatalf("could not get original URL: %v", err)
	}
	fmt.Printf("Original URL: %s\n", getResp.OriginalUrl)

	// Test CreateBatchShortURL
	batchResp, err := client.CreateBatchShortURL(ctx, &pb.CreateBatchShortURLRequest{
		Urls: []*pb.BatchURLRequest{
			{
				CorrelationId: "1",
				OriginalUrl:   "https://example1.com",
			},
			{
				CorrelationId: "2",
				OriginalUrl:   "https://example2.com",
			},
		},
	})
	if err != nil {
		log.Fatalf("could not create batch short URLs: %v", err)
	}
	fmt.Println("Created batch short URLs:")
	for _, url := range batchResp.Urls {
		fmt.Printf("Correlation ID: %s, Short URL: %s\n", url.CorrelationId, url.ShortUrl)
	}

	// Test GetStats
	statsResp, err := client.GetStats(ctx, &pb.GetStatsRequest{})
	if err != nil {
		log.Fatalf("could not get stats: %v", err)
	}
	fmt.Printf("Stats - URLs: %d, Users: %d\n", statsResp.UrlsCount, statsResp.UsersCount)
}
