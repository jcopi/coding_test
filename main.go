package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

const (
	loggerKey = "logger"
	etcdKey   = "etcd"
	itemKey   = "key"
	reqIDKey  = "requestid"
)

type value struct {
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
		reqID = uuid.NewV4().String()
	}
	setInContext(reqIDKey, reqID)
	logger = logger.With(zap.String("req_id", reqID))
	return logger
}

func mustGetEtcd(c *gin.Context) *clientv3.Client {
	client := c.MustGet(etcdKey).(*clientv3.Client)

	return client
}

func mustGetKey(c *gin.Context) string {
	return c.MustGet(itemKey).(string)
}

func GetItem(c *gin.Context) {
	logger := mustGetLogger(c).With(zap.String("method", "GetItem"))
	client := mustGetEtcd(c)

	key := mustGetKey(c)
	if len(key) < 1 {
		logger.Error("invalid key")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var val value
	etcdKey := "/item/" + key

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

	key := mustGetKey(c)
	if len(key) < 1 {
		logger.Error("invalid key")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var val value
	if err := c.BindJSON(&val); err != nil {
		logger.Error("error unmarshalling request", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	etcdKey := "/item/" + key

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := client.Put(ctx, etcdKey, val.Value)
	if err != nil {
		logger.Error("error on get", zap.String("key", etcdKey), zap.Error(err))
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
	api.GET("/:key", GetItem)
	api.POST("/:key", SetItem)

	return s
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	NewServer(logger, client).Run(":8000")
}
