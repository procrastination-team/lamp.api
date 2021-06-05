package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	ginzap "github.com/akath19/gin-zap"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/procrastination-team/lamp.api/internal/db"
	"github.com/procrastination-team/lamp.api/pkg/config"
	"github.com/procrastination-team/lamp.api/pkg/format"
	"go.uber.org/zap"
)

type LampAPI struct {
	http        *http.Server
	mqttClient  mqtt.Client
	mongoClient *db.Storage
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
		zap.L().Error("failed to connect to mqtt", zap.Error(err))
	}

	mongo, err := db.New(&conf.Database, ctx)
	if err != nil {
		return nil, err
	}

	l := &LampAPI{
		http: &http.Server{
			Addr: net.JoinHostPort(conf.Host, conf.Port),
		},
		mongoClient: mongo,
		mqttClient:  client,
	}
	l.http.Handler = l.setupRouter()

	return l, nil
}

func (l *LampAPI) Run() {
	errs := make(chan error, 1)

	defer func() {
		if err := l.http.Close(); err != nil {
			zap.L().Error("server stopped with error", zap.Error(err))
		}
	}()

	go func() {
		zap.L().Info("server started")
		errs <- l.http.ListenAndServe()
	}()

	err := <-errs
	if err != nil {
		zap.L().Error("server exited with error", zap.Error(err))
	}
}

func (l *LampAPI) setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(ginzap.Logger(3*time.Second, zap.L()))

	r.GET("/helloworld", l.helloworld)

	r.GET("/api/lamps", l.getLamps)
	r.GET("/api/lamp/:id", l.getLampByID)
	r.POST("/api/lamps", l.connectLamp)
	r.PUT("/api/lamp/:id", l.updateLamp)
	r.DELETE("/api/lamp/:id", l.deleteLamp)

	return r
}

func (l *LampAPI) getLampByID(c *gin.Context) {
	id := c.Param("id")

	lamp, err := l.mongoClient.GetLampByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 404, "error": "failed to get lamp"})
		zap.L().Error("failed to get lamp", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, lamp)
}

func (l *LampAPI) getLamps(c *gin.Context) {
	lamps, err := l.mongoClient.GetLamps()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 404, "error": "failed to get lamps"})
		zap.L().Error("failed to get lamps", zap.Error(err))
		return
	}
	c.JSON(http.StatusOK, lamps)
}

func (l *LampAPI) connectLamp(c *gin.Context) {
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to read request body"})
		zap.L().Error("failed to read request body", zap.Error(err))
		return
	}

	lamp := format.Lamp{}
	err = json.Unmarshal(bodyBytes, &lamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 404, "error": "failed to parse request"})
		zap.L().Error("failed to parse request", zap.Error(err))
		return
	}

	err = l.mongoClient.CreateLamp(lamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to connect new lamp"})
		zap.L().Error("failed to connect new lamp", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (l *LampAPI) updateLamp(c *gin.Context) {
	id := c.Param("id")

	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to read request body"})
		zap.L().Error("failed to read request body", zap.Error(err))
		return
	}

	lamp := format.Lamp{}
	err = json.Unmarshal(bodyBytes, &lamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 404, "error": "failed to parse request"})
		zap.L().Error("failed to parse request", zap.Error(err))
		return
	}
	lamp.ID = id

	current, err := l.mongoClient.GetLampByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 404, "error": "failed to get lamp"})
		zap.L().Error("failed to get lamp", zap.Error(err))
		return
	}

	err = l.mongoClient.UpdateLamp(lamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "error": "failed to update lamp"})
		zap.L().Error("failed to update lamp", zap.Error(err))
		return
	}

	var msg int
	var topic strings.Builder
	topic.WriteString("/room/lamp")
	topic.WriteString(id)

	if current.Power != lamp.Power {
		topic.WriteString("/power")
		if lamp.Power {
			msg = 1
		} else {
			msg = 0
		}
	} else if current.Brightness != lamp.Brightness {
		topic.WriteString("/brightness")
		msg = lamp.Brightness
	}

	t := l.mqttClient.Publish(topic.String(), 0, false, msg)
	go func() {
		<-t.Done()
		if t.Error() != nil {
			zap.L().Error("failed to publish changes", zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (l *LampAPI) deleteLamp(c *gin.Context) {
	id := c.Param("id")
	err := l.mongoClient.DeleteLamp(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "error": "failed to delete lamp"})
		zap.L().Error("failed to delete lamp", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (l *LampAPI) helloworld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello procrastinating world!"})
}
