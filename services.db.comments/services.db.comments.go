// The backend command runs a Google server that returns fake results.
package main

import (
	"flag"
	"fmt"
	pb "muvik/muvik_admin/protos/comment"
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
	comment_pool *redis.Pool = &redis.Pool{
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
	port       = flag.Int("port", 36030, "RPC port for comments db service")
	debug      = flag.Int("port", 35030, "RPC port for debug comments db service")
	comment_db = flag.String("comment_db", "127.0.0.1:6379", "comment database for connection")
)

// define server type generic
type server struct{}

func (s *server) GetCommentDetail(ctx context.Context, req *pb.RequestCommentID) (*pb.ResponseGetCommentDetail, error) {
	commentDetail, err := gcommentDetail(req.CommentID)
	if err != nil {
		return &pb.ResponseGetCommentDetail{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseGetCommentDetail{
		Error:         pb.ErrorCode_OK,
		Description:   pb.ErrorCode_OK.String(),
		CommentDetail: commentDetail,
	}, err
}
func (s *server) SetCommentDetail(ctx context.Context, req *pb.RequestSetCommentID) (*pb.ResponseGeneral, error) {
	err := scommentDetail(req.CommentID, req.Field)
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
func (s *server) DeleteCommentDetail(ctx context.Context, req *pb.RequestDeleteCommentID) (*pb.ResponseGeneral, error) {
	err := dcommentDetail(req.CommentID, req.Field)
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

// Comment Reply
func (s *server) CountReplyOfComment(ctx context.Context, req *pb.RequestCommentID) (*pb.ResponseCountReply, error) {
	total, err := countReplyComment(req.CommentID)
	if err != nil {
		return &pb.ResponseCountReply{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseCountReply{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Total:       int32(total),
	}, err
}
func (s *server) GetListReplyComment(ctx context.Context, req *pb.RequestCommentID) (*pb.ResponseListReply, error) {
	replies, err := glistReplyComment(req.CommentID)
	if err != nil {
		return &pb.ResponseListReply{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	var res_replies []*pb.ReplyComment
	var res_reply *pb.ReplyComment
	for r_id, r_score := range replies {
		res_reply.ReplyID = r_id
		res_reply.Score = r_score
		res_replies = append(res_replies, res_reply)
	}
	return &pb.ResponseListReply{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Reply:       res_replies,
	}, err
}
func (s *server) AddToListReplyComment(ctx context.Context, req *pb.RequestAddToListReplyComment) (*pb.ResponseGeneral, error) {
	err := addtolistReplyComment(req.CommentID, req.Reply)
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
func (s *server) DeleteFromListReplyComment(ctx context.Context, req *pb.RequestDeleteFromListReplyComment) (*pb.ResponseGeneral, error) {
	err := dlistReplyComment(req.CommentID, req.Reply)
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
	comment_pool.Dial = zconn.RedisConnect("tcp", *comment_db)

	go http.ListenAndServe(fmt.Sprintf(":%d", *debug), nil) // HTTP debugging

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)) // RPC port
	if err != nil {
		log.Error("failed to listen: %v", err)
	}
	g := grpc.NewServer()
	pb.RegisterCommentServiceServer(g, new(server))
	g.Serve(lis)
}
