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
	"path"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

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
	index        = flag.Int("index", 0, "RPC port is 36061+index; debug port is 36661+index")
	redis_port   = flag.String("redis_port", "127.0.0.1:6379", "database port for connection")
	swaggerDir   = flag.String("swagger_dir", "../protos/video", "path to the directory which contains swagger definitions")
	swaggerUIDir = flag.String("swaggerui_dir", "../swagger-ui/dist/", "path to the directory which contains swagger definitions")
)

// define server type generic
type server struct{}

// Get Cover
func (s *server) GetCover(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoCover, error) { // HL
	cover, err := getCoverImage(req.VideoID)
	if err != nil {
		return &pb.ResponseVideoCover{
			Cover:       cover,
			Error:       pb.ErrorCode_DB_ERROR,
			Description: pb.ErrorCode_DB_ERROR.String(),
		}, err
	}
	return &pb.ResponseVideoCover{
		Cover:       cover,
		Error:       pb.ErrorCode_OK,
		Description: pb.ErrorCode_OK.String(),
	}, err

}

func (s *server) GetDetail(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoDetail, error) {
	return nil, nil
}

func (s *server) GetTS(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoTS, error) {
	return nil, nil
}

func (s *server) GetTotalView(ctx context.Context, req *pb.RequestVideoID) (*pb.ResponseVideoTotalView, error) {
	return nil, nil
}

func (s *server) DeleteVideo(ctx context.Context, req *pb.RequestDeleteVideo) (*pb.ResponseDeleteVideo, error) {
	return nil, nil
}

func (s *server) SetVideo(ctx context.Context, req *pb.RequestUpdateVideo) (*pb.ResponseUpdateVideo, error) {
	return nil, nil
}

//gateway
// newGateway returns a new gateway server which translates HTTP into gRPC.
func newGateway(ctx context.Context, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	mux := runtime.NewServeMux(opts...)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterVideoServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", 36061+*index), dialOpts)
	if err != nil {
		return nil, err
	}
	return mux, nil
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/swagger-ui/")
	p = path.Join(*swaggerUIDir, p)
	http.ServeFile(w, r, p)
}

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
		fmt.Errorf("Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	log.Info("Serving %s", r.URL.Path)
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join(*swaggerDir, p)
	http.ServeFile(w, r, p)
}

// allowCORS allows Cross Origin Resoruce Sharing from any origin.
// Don't do this without consideration in production systems.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	log.Info("preflight request for %s", r.URL.Path)
	return
}

// Run starts a HTTP server and blocks forever if successful.
func Run(address string, opts ...runtime.ServeMuxOption) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger-ui/", serveSwaggerUI)
	mux.HandleFunc("/swagger/", serveSwagger)

	gw, err := newGateway(ctx, opts...)
	if err != nil {
		return err
	}
	mux.Handle("/", gw)
	http.ListenAndServe(address, allowCORS(mux))

	return nil
}

func main() {
	//parse config
	flag.Parse()
	//init redisa
	video_pool_info.Dial = zconn.RedisConnect("tcp", *redis_port)

	//	go http.ListenAndServe(fmt.Sprintf(":%d", 36661+*index), nil)   // HTTP debugging
	//run gateway for debug
	go Run(":8080")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 36061+*index)) // RPC port // HL
	if err != nil {
		log.Error("failed to listen: %v", err)
	}
	g := grpc.NewServer()                         // HL
	pb.RegisterVideoServiceServer(g, new(server)) // HL
	g.Serve(lis)                                  // HL

	//run gateway
	//	if err := Run(":10001"); err != nil {
	//		log.Error("%v", err)
	//	}
}
