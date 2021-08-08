package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	loggerKey = "logger"
	etcdKey   = "etcd"
	reqIDKey  = "requestid"
)

type itemValue struct {
	Value string `json:"value"`
}

func setInContext(key string, value interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(key, value)
	}
}

func mustGetLogger(c *gin.Context) *zap.Logger {
	logger := c.MustGet(loggerKey).(*zap.Logger)
	reqID, ok := c.Get(reqIDKey)
	if !ok {
		reqID = uuid.New().String()
	}
	setInContext(reqIDKey, reqID)
	logger = logger.With(zap.String("req_id", reqID.(string)))
	return logger
}

func mustGetEtcd(c *gin.Context) *clientv3.Client {
	client := c.MustGet(etcdKey).(*clientv3.Client)

	return client
}

func GetItem(c *gin.Context) {
	logger := mustGetLogger(c).With(zap.String("method", "GetItem"))
	client := mustGetEtcd(c)

	key := c.Param("key")
	if len(key) < 1 {
		logger.Error("invalid key")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var val itemValue
	etcdKey := "/items/" + key

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.Get(ctx, etcdKey)
	if err != nil {
		logger.Error("error on get", zap.String("key", etcdKey), zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(resp.Kvs) == 0 {
		// key is not set in etcd return not found status code
		c.Status(http.StatusNotFound)
		return
	}
	// for now using first value from etcd
	// this will need to be adjusted when doing more complex gets
	val.Value = string(resp.Kvs[0].Value)

	c.JSON(http.StatusOK, val)
}

func SetItem(c *gin.Context) {
	logger := mustGetLogger(c).With(zap.String("method", "SetItem"))
	client := mustGetEtcd(c)

	key := c.Param("key")
	if len(key) < 1 {
		logger.Error("invalid key")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var val itemValue
	if err := c.BindJSON(&val); err != nil {
		logger.Error("error unmarshalling request", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	etcdKey := "/items/" + key

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := client.Put(ctx, etcdKey, val.Value)
	if err != nil {
		logger.Error("error on set", zap.String("key", etcdKey), zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func DeleteItem(c *gin.Context) {
	logger := mustGetLogger(c).With(zap.String("method", "DeleteItem"))
	client := mustGetEtcd(c)

	key := c.Param("key")
	if len(key) < 1 {
		logger.Error("invalid key")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	etcdKey := "/items/" + key

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := client.Delete(ctx, etcdKey)
	if err != nil {
		logger.Error("error on delete", zap.String("key", etcdKey), zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func NewServer(logger *zap.Logger, etcdClient *clientv3.Client) *gin.Engine {
	s := gin.New()
	s.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/metrics", "/health"),
		gin.Recovery(),
		setInContext(loggerKey, logger),
		setInContext(etcdKey, etcdClient),
	)

	api := s.Group("/api")
	api.GET("/items/*key", GetItem)
	api.POST("/items/*key", SetItem)
	api.DELETE("/items/*key", DeleteItem)

	return s
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://0.0.0.0:2379"},
	})
	if err != nil {
		panic(err)
	}

	NewServer(logger, client).Run(":8000")
}
