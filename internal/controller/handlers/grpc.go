package handlers

import (
	"context"

	"github.com/bbt-t/lets-go-shortener/internal/adapter/storage"
	"github.com/bbt-t/lets-go-shortener/internal/config"
	"github.com/bbt-t/lets-go-shortener/internal/entity"
	"github.com/bbt-t/lets-go-shortener/internal/usecase"
	pb "github.com/bbt-t/lets-go-shortener/pkg/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ShortenerServer is struct for grpc.
type ShortenerServer struct {
	pb.UnimplementedShortenerServer
	cfg     config.Config
	service *usecase.ShortenerService
}

// NewShortenerServer gets new ShortenerServer.
func NewShortenerServer(cfg config.Config, service *usecase.ShortenerService) *ShortenerServer {
	return &ShortenerServer{
		cfg:     cfg,
		service: service,
	}
}

// Ping check connection to storage.
func (server *ShortenerServer) Ping(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	empty := &emptypb.Empty{}
	err := server.service.PingDB()
	if err != nil {
		return empty, status.Error(codes.Unavailable, "Storage doesn't response.")
	}
	return empty, nil
}

// CreateShort creates short from long url.
func (server *ShortenerServer) CreateShort(ctx context.Context, in *pb.Link) (*pb.Link, error) {
	result := &pb.Link{}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	if len(md.Get("userID")) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	id, err := ShortSingleURL(server.service, md.Get("userID")[0], in.LongUrl)

	if err != storage.ErrExists && err != nil {
		return result, err
	}

	result.ShortUrl = server.cfg.BaseURL + "/" + id
	result.Id = id

	if err == storage.ErrExists {
		err = nil
	}

	return result, err
}

// GetStatistics gets count of urls and users.
func (server *ShortenerServer) GetStatistics(ctx context.Context, _ *emptypb.Empty) (*pb.Statistic, error) {
	result := &pb.Statistic{}
	stat, err := server.service.GetStatistic()
	result.Users = uint32(stat.Users)
	result.Urls = uint32(stat.Urls)

	return result, err
}

// GetLong gets long url from short one.
func (server *ShortenerServer) GetLong(ctx context.Context, in *pb.Link) (*pb.Link, error) {
	result := &pb.Link{}
	long, err := server.service.GetOriginal(in.Id)
	if err == storage.ErrNotFound {
		return nil, status.Error(codes.NotFound, "Link not in storage")
	}
	result.LongUrl = long
	return result, err
}

// Delete deletes url from storage.
func (server *ShortenerServer) Delete(ctx context.Context, in *pb.Link) (*emptypb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	if len(md.Get("userID")) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get("userID")[0]

	err := server.service.MarkAsDeleted(userID, []string{in.Id}...)
	return nil, err
}

// GetHistory gets history.
func (server *ShortenerServer) GetHistory(ctx context.Context, in *emptypb.Empty) (*pb.Batch, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("userID")) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get("userID")[0]

	history, err := server.service.GetURLArrayByUser(userID)

	if err != nil {
		return nil, err
	}

	result := &pb.Batch{}
	for _, elem := range history {
		result.Result = append(result.Result, &pb.Link{
			LongUrl:  elem.OriginalURL,
			ShortUrl: elem.ShortURL,
		})
	}

	return result, nil
}

// BatchShort shorts many urls, not single one.
func (server *ShortenerServer) BatchShort(ctx context.Context, in *pb.Batch) (*pb.Batch, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("userID")) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get("userID")[0]

	result := &pb.Batch{}
	query := make([]entity.URLBatch, 0, len(in.Result))

	for _, url := range in.Result {
		query = append(query, entity.URLBatch{
			CorrelationID: url.CorrelationId,
			OriginalURL:   url.LongUrl,
		})
	}

	urls, err := ShortURLs(server.service, userID, query)

	if err == storage.ErrExists {
		err = nil
	}

	if err != nil {
		return nil, err
	}

	result.Result = make([]*pb.Link, 0, len(urls))

	for _, url := range urls {
		result.Result = append(result.Result, &pb.Link{
			CorrelationId: url.CorrelationID,
			ShortUrl:      url.ShortURL,
		})
	}

	return result, nil
}
