// The backend command runs a Google server that returns fake results.
package main

import (
	"flag"
	"fmt"
	pbcomment "muvik/muvik_admin/protos/comment"
	pbvideo "muvik/muvik_admin/protos/video"
	log "muvik/muvik_admin/utilities/logging"
	"net/http"
	_ "net/http/pprof"
	"path"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"golang.org/x/net/context"
	//	"golang.org/x/net/trace"
	"google.golang.org/grpc"
)

// param in command line
var (
	gatewayPort  = flag.String("port", "8080", "Gateway port")
	videoDB      = flag.String("video_db", "localhost:36010", "video db")
	commentDB    = flag.String("comment_db", "localhost:36020", "comment db")
	audioDB      = flag.String("audio_db", "localhost:36030", "audio db")
	userDB       = flag.String("user_db", "localhost:36040", "user db")
	swaggerDir   = flag.String("swagger_dir", "../protos", "path to the directory which contains swagger definitions")
	swaggerUIDir = flag.String("swaggerui_dir", "../swagger-ui/dist/", "path to the directory which contains swagger definitions")
)

//gateway
// newGateway returns a new gateway server which translates HTTP into gRPC.
func newGateway(ctx context.Context, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	mux := runtime.NewServeMux(opts...)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	//regiser video Db service
	err := pbvideo.RegisterVideoServiceHandlerFromEndpoint(ctx, mux, *videoDB, dialOpts)
	if err != nil {
		return nil, err
	}
	//regiser commend DB service
	err = pbcomment.RegisterCommentServiceHandlerFromEndpoint(ctx, mux, *commentDB, dialOpts)
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

	if err := Run(":" + *gatewayPort); err != nil {
		log.Info("%v", err)
	}
}
