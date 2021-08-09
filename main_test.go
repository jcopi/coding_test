package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSingle(t *testing.T) {
	cases := []struct {
		description string
		key         string
		value       string
	}{
		{
			description: "set-get-delete-get test",
			key:         "test_key_0",
			value:       "test value #1",
		},
	}

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	backend := NewMockBackend()

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {

			testSetup := func(c *gin.Context) {
				c.Set(loggerKey, logger)
				c.Set(backendKey, &backend)
			}

			gin.SetMode(gin.TestMode)
			r := gin.Default()
			r.Use(testSetup)
			testApi := r.Group("/api")
			SetupRoutes(testApi)

			valueJson, err := json.Marshal(itemValue{Value: tc.value})
			assert.NoError(t, err)

			endpoint := "/api/items/" + tc.key

			req, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(valueJson))
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)

			req, _ = http.NewRequest(http.MethodGet, endpoint, nil)
			resp = httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			assert.Equal(t, string(valueJson), resp.Body.String())

			req, _ = http.NewRequest(http.MethodDelete, endpoint, nil)
			resp = httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)

			req, _ = http.NewRequest(http.MethodGet, endpoint, nil)
			resp = httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusNotFound, resp.Code)
		})
	}
}

func TestMany(t *testing.T) {
	cases := []struct {
		description string
		keys        []string
		values      []string
	}{
		{
			description: "set-get-delete-get test",
			keys:        []string{"test/keys/0", "test/keys/1", "test_keys_0", "test0", "key0", "key1", "fgndsfkjgnuhuertwgruigshnfjkndsfklj"},
			values:      []string{"test/keys/0", "test/keys/1", "test_keys_0", "test0", "key0", "key1", "fgndsfkjgnuhuertwgruigshnfjkndsfklj"},
		},
	}

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	backend := NewMockBackend()

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {

			testSetup := func(c *gin.Context) {
				c.Set(loggerKey, logger)
				c.Set(backendKey, &backend)
			}

			gin.SetMode(gin.TestMode)
			r := gin.Default()
			r.Use(testSetup)
			testApi := r.Group("/api")
			SetupRoutes(testApi)

			for i, key := range tc.keys {
				valueJson, err := json.Marshal(itemValue{Value: tc.values[i]})
				assert.NoError(t, err)

				endpoint := "/api/items/" + key

				req, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(valueJson))
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				assert.Equal(t, http.StatusOK, resp.Code)
			}

			for i, key := range tc.keys {
				valueJson, err := json.Marshal(itemValue{Value: tc.values[i]})
				assert.NoError(t, err)

				endpoint := "/api/items/" + key

				req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				assert.Equal(t, http.StatusOK, resp.Code)
				assert.Equal(t, string(valueJson), resp.Body.String())

			}

			for _, key := range tc.keys {
				endpoint := "/api/items/" + key

				req, _ := http.NewRequest(http.MethodDelete, endpoint, nil)
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				assert.Equal(t, http.StatusOK, resp.Code)
			}

			for _, key := range tc.keys {
				endpoint := "/api/items/" + key

				req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
				resp := httptest.NewRecorder()
				r.ServeHTTP(resp, req)

				assert.Equal(t, http.StatusNotFound, resp.Code)
			}
		})
	}
}
