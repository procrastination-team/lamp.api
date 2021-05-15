package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/procrastination-team/lamp.api/internal/db"
	"github.com/procrastination-team/lamp.api/pkg/config"
	"github.com/procrastination-team/lamp.api/pkg/format"
)

type LampAPI struct {
	http        *http.Server
	mqttClient  mqtt.Client
	mongoClient *db.Storage
}

func New(conf *config.Settings, ctx context.Context) (*LampAPI, error) {
	/*	opts := mqtt.NewClientOptions()
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
		}*/

	mongo, err := db.New(&conf.Database, ctx)
	if err != nil {
		return nil, err
	}

	l := &LampAPI{
		http: &http.Server{
			Addr: conf.ListenAddress,
		},
		mongoClient: mongo,
		//	mqttClient: client,
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
	r.POST("/api/lamps", l.connectLamp)
	r.PUT("/api/lamp/{id}", l.updateLamp)
	r.DELETE("/api/lamp/{id}", l.deleteLamp)

	return r
}

func (l *LampAPI) getLampByID(c *gin.Context) {
	id := c.Param("id")

	lamp, err := l.mongoClient.GetLampByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failed to get lamp": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lamp)
}

func (l *LampAPI) getLamps(c *gin.Context) {
	lamps, err := l.mongoClient.GetLamps()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failed to get lamps": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lamps)
}

func (l *LampAPI) connectLamp(c *gin.Context) {
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failed to read request body": err.Error()})
		return
	}

	lamp := format.Lamp{}
	err = json.Unmarshal(bodyBytes, &lamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failed to parse JSON": err.Error()})
		return
	}

	err = l.mongoClient.CreateLamp(lamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failed to connect new lamp": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (l *LampAPI) updateLamp(c *gin.Context) {
	id := c.Param("id")

	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failed to read request body": err.Error()})
		return
	}

	lamp := format.Lamp{}
	err = json.Unmarshal(bodyBytes, &lamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failed to parse JSON": err.Error()})
		return
	}

	lamp.ID = id

	err = l.mongoClient.UpdateLamp(lamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failed to update lamp": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (l *LampAPI) deleteLamp(c *gin.Context) {
	id := c.Param("id")
	err := l.mongoClient.DeleteLamp(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failed to delete lamp": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (l *LampAPI) helloworld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello procrastinating world!"})
}
