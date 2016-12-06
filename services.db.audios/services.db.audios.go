// The backend command runs a Google server that returns fake results.
package main

import (
	"flag"
	"fmt"
	pb "muvik/muvik_admin/protos/audio"
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
	audio_pool *redis.Pool = &redis.Pool{
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
	index    = flag.Int("index", 0, "RPC port is 36030+index; debug port is 35030+index")
	audio_db = flag.String("audio_db", "127.0.0.1:6379", "audio database for connection")
)

// define server type generic
type server struct{}

//get detail any hash in DB 9701
func (s *server) GetDetail(ctx context.Context, req *pb.RequestGetDetail) (*pb.ResponseGetDetail, error) {
	detail, err := gDetail(req.RequestID, pb.AudioKind(req.Kind))
	if err != nil {
		return &pb.ResponseGetDetail{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGetDetail{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Detail:      detail,
	}, err
}

func (s *server) DeleteDetail(ctx context.Context, req *pb.RequestModifyDetail) (*pb.ResponseGeneral, error) {
	err := dDetail(req.RequestID, pb.AudioKind(req.Kind), req.Field)
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

func (s *server) SetDetail(ctx context.Context, req *pb.RequestModifyDetail) (*pb.ResponseGeneral, error) {
	err := sDetail(req.RequestID, pb.AudioKind(req.Kind), req.Field)
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

//get list any zset in DB 9701
func (s *server) GetList(ctx context.Context, req *pb.RequestGetList) (*pb.ResponseGetList, error) {
	list, err := gList(req.RequestID, pb.AudioListType(req.Listtype))
	if err != nil {
		return &pb.ResponseGetList{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGetList{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		List:        list,
	}, err
}

func (s *server) AddToList(ctx context.Context, req *pb.RequestModifyList) (*pb.ResponseGeneral, error) {
	err := aList(req.RequestID, pb.AudioListType(req.Listtype), req.MemberScore)
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

func (s *server) RemoveToList(ctx context.Context, req *pb.RequestModifyList) (*pb.ResponseGeneral, error) {
	err := rList(req.RequestID, pb.AudioListType(req.Listtype), req.MemberScore)
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
	audio_pool.Dial = zconn.RedisConnect("tcp", *audio_db)

	go http.ListenAndServe(fmt.Sprintf(":%d", 35030+*index), nil) // HTTP debugging

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 36030+*index)) // RPC port
	if err != nil {
		log.Error("failed to listen: %v", err)
	}
	g := grpc.NewServer()
	pb.RegisterAudioServiceServer(g, new(server))
	g.Serve(lis)
}
