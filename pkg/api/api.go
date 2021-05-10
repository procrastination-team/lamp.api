package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/procrastination-team/lamp.api/pkg/config"
)

type LampAPI struct {
	http       *http.Server
	mqttClient mqtt.Client
}

func New(conf *config.Settings, ctx context.Context) (*LampAPI, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", conf.Mqtt.Address))
	opts.SetUsername(conf.Mqtt.Username)
	opts.SetPassword(conf.Mqtt.Password)
	opts.SetClientID(conf.Mqtt.ClientID)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}

	l := &LampAPI{
		http: &http.Server{
			Addr: conf.ListenAddress,
		},
		mqttClient: client,
	}
	l.http.Handler = l.setupRouter()

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

	r.GET("/api/lamps", l.getLamps)
	r.GET("/api/lamp/{id}", l.getLampByID)
	r.POST("/api/lamp/{id}", l.connectLamp)
	r.PUT("/api/lamp/{id}", l.changeBrightness)
	r.DELETE("/api/lamp/{id}", l.deleteLamp)

	return r
}

func (l *LampAPI) connectLamp(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (l *LampAPI) changeBrightness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (l *LampAPI) deleteLamp(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (l *LampAPI) getLampByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (l *LampAPI) getLamps(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (l *LampAPI) helloworld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello procrastinating world!"})
}
