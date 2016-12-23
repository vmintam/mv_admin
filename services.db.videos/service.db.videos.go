// The backend command runs a Google server that returns fake results.
package main

import (
	"flag"
	"fmt"
	pb "muvik/muvik_admin/protos/video"
	log "muvik/muvik_admin/utilities/logging"
	"muvik/muvik_admin/utilities/zconn"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/context"
	//	"golang.org/x/net/trace"
	"google.golang.org/grpc"
)

//init redis connection pool

var (
	video_pool_info *redis.Pool = &redis.Pool{
		MaxIdle:     500,
		MaxActive:   500,
		IdleTimeout: 5 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
)

// param in command line
var (
	port     = flag.Int("port", 36020, "RPC port for video db services")
	debug    = flag.Int("debug", 35020, "HTTP debug port for video db services")
	video_db = flag.String("video_db", "127.0.0.1:6379", "video database for connection")
)

// define server type generic
type server struct{}

func (s *server) GetVideoDetailOne(ctx context.Context, req *pb.RequestVideoDetailOne) (*pb.ResponseVideoDetailOne, error) {
	result, err := gVideoOne(req.VideoID, req.Field)
	if err != nil {
		return &pb.ResponseVideoDetailOne{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseVideoDetailOne{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Result:      result,
	}, err
}

func (s *server) GetVideoDetail(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoDetail, error) {
	result, err := gVideoDetail(req.VideoID)
	if err != nil {
		return &pb.ResponseVideoDetail{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseVideoDetail{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		VideoDetail: result,
	}, err
}

func (s *server) DeleteVideoDetail(ctx context.Context, req *pb.RequestModifyVideo) (*pb.ResponseGeneral, error) {
	err := dVideoDetail(req.VideoID, req.Field)
	if err != nil {
		return &pb.ResponseGeneral{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGeneral{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
	}, err
}

func (s *server) SetVideoDetail(ctx context.Context, req *pb.RequestModifyVideo) (*pb.ResponseGeneral, error) {
	err := sVideoDetail(req.VideoID, req.Field)
	if err != nil {
		return &pb.ResponseGeneral{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGeneral{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
	}, err
}

//promote video
func (s *server) GetPromoteVideo(ctx context.Context, req *pb.RequestGetPromoteVideo) (*pb.ResponseGetPromoteVideo, error) {
	videoID, err := gPromoteVideo()
	if err != nil {
		return &pb.ResponseGetPromoteVideo{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGetPromoteVideo{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		VideoID:     videoID,
	}, err
}

func (s *server) SetPromoteVideo(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseGeneral, error) {
	err := sPromoteVideo(req.VideoID)
	if err != nil {
		return &pb.ResponseGeneral{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGeneral{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
	}, err
}

//list user Like, list comment of video

func (s *server) GetListVideoOne(ctx context.Context, req *pb.RequestGetListVideoOne) (*pb.ResponseGetListVideoOne, error) {
	result, err := gListOne(req.RequestID, req.Listtype, req.Member)
	if err != nil {
		return &pb.ResponseGetListVideoOne{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGetListVideoOne{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Result:      result,
	}, err
}

func (s *server) GetListVideo(ctx context.Context, req *pb.RequestGetListVideo) (*pb.ResponseGetListVideo, error) {
	result, err := gList(req.RequestID, req.Listtype)
	if err != nil {
		return &pb.ResponseGetListVideo{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGetListVideo{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		List:        result,
	}, err
}

func (s *server) AddListVideo(ctx context.Context, req *pb.RequestModifyListVideo) (*pb.ResponseGeneral, error) {
	err := aList(req.RequestID, req.Listtype, req.MemberScore)
	if err != nil {
		return &pb.ResponseGeneral{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGeneral{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
	}, err
}

func (s *server) RemoveListVideo(ctx context.Context, req *pb.RequestModifyListVideo) (*pb.ResponseGeneral, error) {
	err := rList(req.RequestID, req.Listtype, req.MemberScore)
	if err != nil {
		return &pb.ResponseGeneral{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGeneral{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
	}, err
}
func main() {
	//parse config
	flag.Parse()
	//init redisa
	video_pool_info.Dial = zconn.RedisConnect("tcp", *video_db)

	go http.ListenAndServe(fmt.Sprintf(":%d", *debug), nil) // HTTP debugging

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)) // RPC port
	if err != nil {
		log.Error("failed to listen: %v", err)
	}
	g := grpc.NewServer()
	pb.RegisterVideoServiceServer(g, new(server))
	g.Serve(lis)
}
