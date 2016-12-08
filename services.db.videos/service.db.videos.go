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
	index    = flag.Int("index", 0, "RPC port is 36010+index; debug port is 35010+index")
	video_db = flag.String("video_db", "127.0.0.1:6379", "video database for connection")
)

// define server type generic
type server struct{}

func (s *server) GeOnetVideoDetail(ctx context.Context, req *pb.RequestVideoOne) (*pb.ResponseOneVideoDetail, error) {
	return nil, nil
}

func (s *server) GetVideoDetail(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoDetail, error) {
	return nil, nil
}

func (s *server) DeleteVideoDetail(ctx context.Context, req *pb.RequestModifyVideo) (*pb.ResponseGeneral, error) {
	return nil, nil
}

func (s *server) SetVideoDetail(ctx context.Context, req *pb.RequestModifyVideo) (*pb.ResponseGeneral, error) {
	return nil, nil
}

//promote video
func (s *server) GetPromoteVideo(ctx context.Context, req *pb.RequestGetPromoteVideo) (*pb.ResponseGetPromoteVideo, error) {
	return nil, nil
}

func (s *server) SetPromoteVideo(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseGeneral, error) {
	return nil, nil
}

//list user Like, list comment of video

func (s *server) GetListVideo(ctx context.Context, req *pb.RequestGetListVideo) (*pb.ResponseGetListVideo, error) {
	return nil, nil
}

func (s *server) AddListVideo(ctx context.Context, req *pb.RequestModifyListVideo) (*pb.ResponseGeneral, error) {
	return nil, nil
}

func (s *server) RemoveListVideo(ctx context.Context, req *pb.RequestModifyListVideo) (*pb.ResponseGeneral, error) {
	return nil, nil
}
func main() {
	//parse config
	flag.Parse()
	//init redisa
	video_pool_info.Dial = zconn.RedisConnect("tcp", *video_db)

	go http.ListenAndServe(fmt.Sprintf(":%d", 35010+*index), nil) // HTTP debugging

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 36010+*index)) // RPC port
	if err != nil {
		log.Error("failed to listen: %v", err)
	}
	g := grpc.NewServer()
	pb.RegisterVideoServiceServer(g, new(server))
	g.Serve(lis)
}
