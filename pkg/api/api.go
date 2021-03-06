package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/procrastination-team/lamp.api/internal/db"
	"github.com/procrastination-team/lamp.api/pkg/config"
)

type LampAPI struct {
	storage *db.Storage
	http    *http.Server
}

func New(conf *config.Settings, ctx context.Context) (*LampAPI, error) {
	l := &LampAPI{
		http: &http.Server{
			Addr: conf.ListenAddress,
		},
	}
	l.http.Handler = l.setupRouter()

	var err error
	l.storage, err = db.New(&conf.Database, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	return l, nil
}

func (l *LampAPI) Run() {
	errs := make(chan error, 1)

	defer func() {
		if err := l.http.Close(); err != nil {
			log.Fatal(fmt.Errorf("server stopped with error: %w", err))
		}
	}()

	go func() {
		log.Printf("server started")
		errs <- l.http.ListenAndServe()
	}()

	err := <-errs
	if err != nil {
		log.Fatal("server exited with error: %w", err)
	}
}

func (l *LampAPI) setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/helloworld", l.helloworld)

	return r
}

func (l *LampAPI) helloworld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello procrastinating world!"})
}
