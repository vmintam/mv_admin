// The backend command runs a Google server that returns fake results.
package main

import (
	"flag"
	"fmt"
	pb "muvik/muvik_admin/protos/user"
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
	user_pool *redis.Pool = &redis.Pool{
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
	port    = flag.Int("port", 36000, "RPC port for user db services")
	debug   = flag.Int("debug", 35000, "HTTP debug port for user db services")
	user_db = flag.String("user_db", "127.0.0.1:6379", "user database for connection")
)

// define server type generic
type server struct{}

//get detail any hash in DB 9500
func (s *server) GetOne(ctx context.Context, req *pb.RequestGetOne) (*pb.ResponseGetOne, error) {
	result, err := gDetailOne(req.RequestID, req.Kind, req.Field)
	if err != nil {
		return &pb.ResponseGetOne{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGetOne{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Result:      result,
	}, err
}

func (s *server) GetDetail(ctx context.Context, req *pb.RequestGetDetail) (*pb.ResponseGetDetail, error) {
	detail, err := gDetail(req.RequestID, req.Kind)
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
	err := dDetail(req.RequestID, req.Kind, req.Field)
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
	err := sDetail(req.RequestID, req.Kind, req.Field)
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
func (s *server) GetListOne(ctx context.Context, req *pb.RequestGetListOne) (*pb.ResponseGetListOne, error) {
	result, err := gListOne(req.RequestID, req.Listtype, req.Member)
	if err != nil {
		return &pb.ResponseGetListOne{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGetListOne{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Result:      result,
	}, err
}

func (s *server) GetList(ctx context.Context, req *pb.RequestGetList) (*pb.ResponseGetList, error) {
	list, err := gList(req.RequestID, req.Listtype)
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

func (s *server) RemoveToList(ctx context.Context, req *pb.RequestModifyList) (*pb.ResponseGeneral, error) {
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
	//init redis
	user_pool.Dial = zconn.RedisConnect("tcp", *user_db)

	go http.ListenAndServe(fmt.Sprintf(":%d", *debug), nil) // HTTP debugging

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)) // RPC port
	if err != nil {
		log.Error("failed to listen: %v", err)
	}
	g := grpc.NewServer()
	pb.RegisterUserServiceServer(g, new(server))
	g.Serve(lis)
}
