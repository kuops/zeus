package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	"net/http"
	"time"
	"zeus/internal/handler"
)

type Server struct {
	config *Config
	routes *gin.Engine
}

func New(cfg *Config) *Server {
	return &Server{
		config: cfg,
	}
}

func (s *Server) HTTPServer(stopCh <-chan struct{}) error {
	addr := fmt.Sprintf(":%v", s.config.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.routes,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			klog.Errorf("failed start : %w", err)
		}
	}()
	klog.Infof("start http server at %v", httpServer.Addr)

	<-stopCh
	shutdownCtx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	klog.Info("shutting down http server")
	return httpServer.Shutdown(shutdownCtx)
}

func (s *Server) Run(handlers *handler.Handlers, stopCh <-chan struct{}) error {
	s.routes = s.Routes(handlers)
	return s.HTTPServer(stopCh)
}
