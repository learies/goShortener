package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/learies/goShortener/internal/models"
	"github.com/learies/goShortener/internal/services"
	pb "github.com/learies/goShortener/proto"
	"google.golang.org/grpc"
)

// Server implements the gRPC URLShortener service
type Server struct {
	pb.UnimplementedURLShortenerServer
	*grpc.Server
	service *services.URLShortenerService
}

// NewServer creates a new gRPC server instance
func NewServer(service *services.URLShortenerService) *Server {
	s := &Server{
		Server:  grpc.NewServer(),
		service: service,
	}
	pb.RegisterURLShortenerServer(s.Server, s)
	return s
}

// CreateShortURL implements the CreateShortURL RPC method
func (s *Server) CreateShortURL(ctx context.Context, req *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	// TODO: Get user ID from context (implement authentication)
	userID := uuid.New()

	result, err := s.service.CreateShortURL(ctx, req.Url, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create short URL: %w", err)
	}

	return &pb.CreateShortURLResponse{
		Result: result,
	}, nil
}

// GetOriginalURL implements the GetOriginalURL RPC method
func (s *Server) GetOriginalURL(ctx context.Context, req *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	result, err := s.service.GetOriginalURL(ctx, req.ShortUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get original URL: %w", err)
	}

	return &pb.GetOriginalURLResponse{
		OriginalUrl: result,
	}, nil
}

// CreateBatchShortURL implements the CreateBatchShortURL RPC method
func (s *Server) CreateBatchShortURL(ctx context.Context, req *pb.CreateBatchShortURLRequest) (*pb.CreateBatchShortURLResponse, error) {
	// TODO: Get user ID from context (implement authentication)
	userID := uuid.New()

	batchRequest := make([]models.ShortenBatchRequest, len(req.Urls))
	for i, url := range req.Urls {
		batchRequest[i] = models.ShortenBatchRequest{
			CorrelationID: url.CorrelationId,
			OriginalURL:   url.OriginalUrl,
		}
	}

	result, err := s.service.CreateBatchShortURL(ctx, batchRequest, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch short URLs: %w", err)
	}

	response := &pb.CreateBatchShortURLResponse{
		Urls: make([]*pb.BatchURLResponse, len(result)),
	}

	for i, url := range result {
		response.Urls[i] = &pb.BatchURLResponse{
			CorrelationId: url.CorrelationID,
			ShortUrl:      url.ShortURL,
		}
	}

	return response, nil
}

// GetUserURLs implements the GetUserURLs RPC method
func (s *Server) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	// TODO: Get user ID from context (implement authentication)
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	result, err := s.service.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}

	response := &pb.GetUserURLsResponse{
		Urls: make([]*pb.UserURL, len(result)),
	}

	for i, url := range result {
		response.Urls[i] = &pb.UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}
	}

	return response, nil
}

// DeleteUserURLs implements the DeleteUserURLs RPC method
func (s *Server) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	// TODO: Get user ID from context (implement authentication)
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	err = s.service.DeleteUserURLs(ctx, userID, req.ShortUrls)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user URLs: %w", err)
	}

	return &pb.DeleteUserURLsResponse{
		Success: true,
	}, nil
}

// GetStats implements the GetStats RPC method
func (s *Server) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	urlsCount, usersCount, err := s.service.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &pb.GetStatsResponse{
		UrlsCount:  int32(urlsCount),
		UsersCount: int32(usersCount),
	}, nil
}
