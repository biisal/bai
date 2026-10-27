package web

import (
	"io/fs"
	"net/http"

	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
)

const (
	authCookieName = "bai_web_token"
	viteOrigin     = "http://localhost:5173"
)

type Options struct {
	Gateway  *agent.Gateway
	Config   *config.Config
	Registry *Registry
	Token    string
	Addr     string
	Dev      bool
	Static   fs.FS
}

type Server struct {
	gateway  *agent.Gateway
	config   *config.Config
	registry *Registry
	token    []byte
	addr     string
	dev      bool
	static   fs.FS
	mux      *http.ServeMux
	hosts    map[string]bool
}

func NewServer(opts Options) *Server {
	s := &Server{
		gateway:  opts.Gateway,
		config:   opts.Config,
		registry: opts.Registry,
		token:    []byte(opts.Token),
		addr:     opts.Addr,
		dev:      opts.Dev,
		static:   opts.Static,
		mux:      http.NewServeMux(),
		hosts:    allowedHosts(opts.Addr),
	}
	s.routes()
	return s
}
