package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	loggerKey  = "logger"
	backendKey = "backend"
	reqIDKey   = "requestid"
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
	c.Set(reqIDKey, reqID)
	logger = logger.With(zap.String("req_id", reqID.(string)))
	return logger
}

func mustGetBackend(c *gin.Context) BackendStore {
	client := c.MustGet(backendKey).(BackendStore)

	return client
}

func GetItem(c *gin.Context) {
	logger := mustGetLogger(c).With(zap.String("method", "GetItem"))
	backend := mustGetBackend(c)

	key := c.Param("key")
	if len(key) < 1 {
		logger.Error("invalid key")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var val itemValue

	value, found, err := backend.Get(key)
	if err != nil {
		logger.Error("error on get", zap.String("key", key), zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else if !found {
		c.Status(http.StatusNotFound)
		return
	} else {
		val.Value = value
		c.JSON(http.StatusOK, val)
		return
	}
}

func SetItem(c *gin.Context) {
	logger := mustGetLogger(c).With(zap.String("method", "SetItem"))
	backend := mustGetBackend(c)

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

	if err := backend.Set(key, val.Value); err != nil {
		logger.Error("error on set", zap.String("key", key), zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func DeleteItem(c *gin.Context) {
	logger := mustGetLogger(c).With(zap.String("method", "DeleteItem"))
	backend := mustGetBackend(c)

	key := c.Param("key")
	if len(key) < 1 {
		logger.Error("invalid key")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := backend.Delete(key); err != nil {
		logger.Error("error on delete", zap.String("key", key), zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func SetupRoutes(rg *gin.RouterGroup) {
	rg.GET("/items/*key", GetItem)
	rg.POST("/items/*key", SetItem)
	rg.DELETE("/items/*key", DeleteItem)
}

func NewServer(logger *zap.Logger, backend BackendStore) *gin.Engine {
	s := gin.New()
	s.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/metrics", "/health"),
		gin.Recovery(),
		setInContext(loggerKey, logger),
		setInContext(backendKey, backend),
	)

	api := s.Group("/api")
	SetupRoutes(api)

	return s
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	client, err := NewEtcdBackend([]string{"http://0.0.0.0:2379"})
	if err != nil {
		panic(err)
	}

	NewServer(logger, &client).Run(":8000")
}
