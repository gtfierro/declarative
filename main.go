package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
	"github.com/skratchdot/open-golang/open"
)

// logger
var log *logging.Logger

// set up logging facilities
func init() {
	log = logging.MustGetLogger("http")
	var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
	var logBackend = logging.NewLogBackend(os.Stderr, "", 0)
	logBackendLeveled := logging.AddModuleLevel(logBackend)
	logging.SetBackend(logBackendLeveled)
	logging.SetFormatter(logging.MustStringFormatter(format))
}

type Config struct {
	ServerPort    string
	StaticPath    string
	ListenAddress string
	UseIPv6       bool
}

type server struct {
	router     *httprouter.Router
	port       string
	staticpath string
}

func startServer(cfg *Config) {
	s := &server{
		port:       cfg.ServerPort,
		staticpath: cfg.StaticPath,
	}
	r := httprouter.New()
	r.ServeFiles("/static/*filepath", http.Dir(cfg.StaticPath+"/static"))
	r.GET("/", s.serveMain)
	s.router = r
	// configure server
	var (
		addrString string
		nettype    string
	)

	// check if ipv6
	if cfg.UseIPv6 {
		nettype = "tcp6"
		addrString = "[" + cfg.ListenAddress + "]:" + s.port
	} else {
		nettype = "tcp4"
		addrString = cfg.ListenAddress + ":" + s.port
	}

	address, err := net.ResolveTCPAddr(nettype, addrString)
	if err != nil {
		log.Fatalf("Error resolving address %s (%s)", addrString, err.Error())
	}

	http.Handle("/", s.router)
	log.Notice("Starting HTTP Server on ", addrString)
	srv := &http.Server{
		Addr: address.String(),
	}
	log.Fatal(srv.ListenAndServe())

}

func (srv *server) serveMain(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve help from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/index.html")
}

func main() {
	cfg := &Config{
		ServerPort:    "8888",
		StaticPath:    ".",
		ListenAddress: "127.0.0.1",
		UseIPv6:       false,
	}

	go func() {
		<-time.After(2 * time.Second)
		open.Run("http://localhost:8888")
	}()
	startServer(cfg)
}
