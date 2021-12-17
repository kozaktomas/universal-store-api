package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	ginlimiter "github.com/julianshen/gin-limiter"
	"github.com/kozaktomas/universal-store-api/config"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type httpServer struct {
	endpoints map[string]Service
	engine    *gin.Engine
}

func createHttpServer(endpoints map[string]Service) (*httpServer, error) {
	server := &httpServer{
		endpoints: endpoints,
		engine:    gin.Default(),
	}

	server.installPrometheus()

	for _, endpoint := range endpoints {
		err := server.registerHandlers(endpoint)
		if err != nil {
			return nil, err
		}
	}

	server.registerIndexHandler()
	server.engine.NoRoute(notFound)
	server.engine.NoMethod(notFound)

	return server, nil
}

func (server *httpServer) registerHandlers(endpoint Service) error {
	name := endpoint.Cfg.Name

	type limitFunc func() (config.Limit, error)
	type handler struct {
		httpMethod   string
		url          string
		limitFunc    limitFunc
		callbackFunc gin.HandlerFunc
	}

	handlers := []handler{
		{
			httpMethod:   http.MethodGet,
			url:          fmt.Sprintf("/%s", name),
			limitFunc:    endpoint.Cfg.ApiConfig.Limits.ParseList,
			callbackFunc: server.createListEndpoint(endpoint),
		},
		{
			httpMethod:   http.MethodGet,
			url:          fmt.Sprintf("/%s/:id", name),
			limitFunc:    endpoint.Cfg.ApiConfig.Limits.ParseGet,
			callbackFunc: server.createGetEndpoint(endpoint),
		},
		{
			httpMethod:   http.MethodPut,
			url:          fmt.Sprintf("/%s", name),
			limitFunc:    endpoint.Cfg.ApiConfig.Limits.ParsePut,
			callbackFunc: server.createPutEndpoint(endpoint),
		},
		{
			httpMethod:   http.MethodDelete,
			url:          fmt.Sprintf("/%s/:id", name),
			limitFunc:    endpoint.Cfg.ApiConfig.Limits.ParseDelete,
			callbackFunc: server.createDeleteEndpoint(endpoint),
		},
	}

	for _, h := range handlers {
		hLimit, err := h.limitFunc()
		if err != nil {
			return err
		}
		if !hLimit.Disabled {
			server.engine.Handle(
				h.httpMethod,
				h.url,
				server.createAuthMiddleware(endpoint),
				createRateLimiterMiddleware(hLimit),
				h.callbackFunc,
			)
		}
	}

	return nil
}

func (server *httpServer) createAuthMiddleware(endpoint Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if endpoint.Cfg.ApiConfig.Bearer != nil && *endpoint.Cfg.ApiConfig.Bearer != "" {
			bearerTokenHeader := c.GetHeader("Authorization")
			parts := strings.Split(bearerTokenHeader, " ")
			if len(parts) != 2 {
				_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid Authorization header input"))
				return
			}

			if *endpoint.Cfg.ApiConfig.Bearer != parts[1] {
				_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid Authorization bearer token"))
				return
			}
		}
	}
}

func (server *httpServer) createListEndpoint(endpoint Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		list, err := endpoint.List()
		if err != nil {
			c.String(http.StatusInternalServerError, "could not read data from storage")
			return
		}

		c.JSON(200, list)
	}
}

func (server *httpServer) createGetEndpoint(endpoint Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		entity, err := endpoint.Get(id)
		if err != nil {
			c.String(http.StatusNotFound, "could not find entity with id %q", id)
			return
		}

		c.JSON(200, entity.Payload)
	}
}
func (server *httpServer) createPutEndpoint(endpoint Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rawJson map[string]interface{}
		rawData, _ := c.GetRawData()
		if err := json.Unmarshal(rawData, &rawJson); err != nil {
			c.String(http.StatusBadRequest, "could not parse json data: %w", err)
			return
		}

		if err := endpoint.Validate(rawJson); err != nil {
			c.String(http.StatusBadRequest, "invalid input: %s", err.Error())
			return
		}

		if err := endpoint.Put(rawJson); err != nil {
			c.String(http.StatusInternalServerError, "could not store requested data")
			return
		}

		c.String(http.StatusNoContent, "")
	}
}
func (server *httpServer) createDeleteEndpoint(endpoint Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := endpoint.Delete(id); err != nil {
			c.String(http.StatusNotFound, "could not find an entity with id %q", id)
			return
		}

		c.String(204, "")
	}
}

func (server *httpServer) installPrometheus() {
	p := ginprometheus.NewPrometheus("gin")
	p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.FullPath()
		return url
	}
	p.Use(server.engine)
}

func (server *httpServer) registerIndexHandler() {
	server.engine.GET("/", func(c *gin.Context) {
		type responseItem struct {
			Method      string `json:"method"`
			Endpoint    string `json:"endpoint"`
			Description string `json:"description"`
		}

		var response []responseItem

		for _, endpoint := range server.endpoints {
			response = append(response, responseItem{
				Method:      "GET",
				Endpoint:    fmt.Sprintf("/%s", endpoint.Cfg.Name),
				Description: fmt.Sprintf("Get list of %s", endpoint.Cfg.Name),
			})

			response = append(response, responseItem{
				Method:      "GET",
				Endpoint:    fmt.Sprintf("/%s/:id", endpoint.Cfg.Name),
				Description: fmt.Sprintf("Get detail information about %s", endpoint.Cfg.Name),
			})

			response = append(response, responseItem{
				Method:      "PUT",
				Endpoint:    fmt.Sprintf("/%s", endpoint.Cfg.Name),
				Description: fmt.Sprintf("Creates a new entity of %s list", endpoint.Cfg.Name),
			})

			response = append(response, responseItem{
				Method:      "DELETE",
				Endpoint:    fmt.Sprintf("/%s/:id", endpoint.Cfg.Name),
				Description: fmt.Sprintf("Deletes entity of %s list", endpoint.Cfg.Name),
			})
		}

		c.JSON(http.StatusOK, response)
	})
}

func (server *httpServer) Run(port int) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func createRateLimiterMiddleware(limit config.Limit) gin.HandlerFunc {
	if limit.Unlimited {
		return func(c *gin.Context) {
			// unlimited
		}
	}

	lm := ginlimiter.NewRateLimiter(limit.Interval, int64(limit.Count), func(ctx *gin.Context) (string, error) {
		return "d", nil // just use one key for everything
	})

	return lm.Middleware()
}

func notFound(c *gin.Context) {
	c.String(404, "not found")
}
