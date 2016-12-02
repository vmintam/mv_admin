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

const (
	COVER_URL_PREFIX = `http://cv.muvik.vn/cover/%s`
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
	index   = flag.Int("index", 0, "RPC port is 36010+index; debug port is 35010+index")
	videodb = flag.String("videodb", "127.0.0.1:6379", "video database port for connection")
)

// define server type generic
type server struct{}

// Get Cover
func (s *server) GetVideoCover(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoCover, error) { // HL
	cover, err := getCoverImage(req.VideoID)
	if err != nil {
		return &pb.ResponseVideoCover{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseVideoCover{
		Cover:       fmt.Sprintf(COVER_URL_PREFIX, cover),
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
	}, err

}

func (s *server) GetVideoDetail(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoDetail, error) {
	videod, err := getVideoDetail(req.VideoID)
	if err != nil {
		return &pb.ResponseVideoDetail{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseVideoDetail{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		VideoDetail: videod,
	}, err
}

func (s *server) GetVideoTS(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoTS, error) {
	ts, err := getVideoCreated(req.VideoID)
	if err != nil {
		return &pb.ResponseVideoTS{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseVideoTS{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Timestamp:   ts,
	}, err
}

func (s *server) GetVideoTotalView(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoTotalView, error) {
	totalv, err := getTotalViews(req.VideoID)
	if err != nil {
		return &pb.ResponseVideoTotalView{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseVideoTotalView{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Totalview:   int32(totalv),
	}, err
}

func (s *server) DeleteVideoDetail(ctx context.Context, req *pb.RequestDeleteVideo) (*pb.ResponseGeneral, error) {
	err := deleteVideo(req.VideoID, req.Field)
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

func (s *server) SetVideoDetail(ctx context.Context, req *pb.RequestUpdateVideo) (*pb.ResponseGeneral, error) {
	err := updateVideo(req.VideoID, req.Field)
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

func (s *server) GetListCommentsVideo(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseListCommentsVideo, error) {
	lists, err := GetListComments(req.VideoID)
	if err != nil {
		return &pb.ResponseListCommentsVideo{
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	var css []*pb.CommentsVideo
	var cs pb.CommentsVideo
	for comment, score := range lists {
		cs.CommentID = comment
		cs.Score = score
		css = append(css, &cs)
	}
	return &pb.ResponseListCommentsVideo{
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
		Comment:     css,
	}, err
}

func (s *server) AddCommentsVideo(ctx context.Context, req *pb.RequestAddCommentsVideoID) (*pb.ResponseGeneral, error) {
	err := AddComments(req.VideoID, req.CommentScore)
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

func (s *server) RemoveCommentsVideo(ctx context.Context, req *pb.RequestRemoveCommentsVideoID) (*pb.ResponseGeneral, error) {
	err := DeleteComments(req.VideoID, req.Comment)
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

func (s *server) GetPromoteVideo(ctx context.Context, req *pb.RequestGetPromoteVideo) (*pb.ResponseGetPromoteVideo, error) {
	videoID, err := GPromoteVideo()
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
	err := SPromoteVideo(req.VideoID)
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
	video_pool_info.Dial = zconn.RedisConnect("tcp", *videodb)

	go http.ListenAndServe(fmt.Sprintf(":%d", 35010+*index), nil) // HTTP debugging

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 36010+*index)) // RPC port
	if err != nil {
		log.Error("failed to listen: %v", err)
	}
	g := grpc.NewServer()
	pb.RegisterVideoServiceServer(g, new(server))
	g.Serve(lis)
}
