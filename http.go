package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	ginlimiter "github.com/julianshen/gin-limiter"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
	ginprometheus "github.com/zsais/go-gin-prometheus"
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
	logger    *logrus.Logger
}

func createHttpServer(endpoints map[string]Service, logger *logrus.Logger) (*httpServer, error) {
	gin.SetMode(gin.ReleaseMode)

	server := &httpServer{
		endpoints: endpoints,
		engine:    gin.New(),
		logger:    logger,
	}

	server.engine.Use(ginlogrus.Logger(logger))
	server.engine.Use(gin.Recovery())
	server.registerPrometheus()

	for _, endpoint := range endpoints {
		err := server.registerHandlers(endpoint)
		if err != nil {
			return nil, err
		}
	}

	server.registerLogLevelHandler()
	server.registerIndexHandler()
	server.engine.NoRoute(notFound)
	server.engine.NoMethod(notFound)

	return server, nil
}

func (server *httpServer) registerHandlers(endpoint Service) error {
	name := endpoint.Cfg.Name

	type limitFunc func() (Limit, error)
	type handler struct {
		httpMethod   string
		url          string
		limitFunc    limitFunc
		callbackFunc gin.HandlerFunc
	}

	handlers := []handler{
		{
			httpMethod:   http.MethodGet,
			url:          "",
			limitFunc:    endpoint.Cfg.ApiConfig.Limits.ParseList,
			callbackFunc: server.createListEndpoint(endpoint),
		},
		{
			httpMethod:   http.MethodGet,
			url:          "/:id",
			limitFunc:    endpoint.Cfg.ApiConfig.Limits.ParseGet,
			callbackFunc: server.createGetEndpoint(endpoint),
		},
		{
			httpMethod:   http.MethodOptions,
			url:          "",
			limitFunc:    endpoint.Cfg.ApiConfig.Limits.ParsePut,
			callbackFunc: func(c *gin.Context) {},
		},
		{
			httpMethod:   http.MethodPut,
			url:          "",
			limitFunc:    endpoint.Cfg.ApiConfig.Limits.ParsePut,
			callbackFunc: server.createPutEndpoint(endpoint),
		},
		{
			httpMethod:   http.MethodDelete,
			url:          "/:id",
			limitFunc:    endpoint.Cfg.ApiConfig.Limits.ParseDelete,
			callbackFunc: server.createDeleteEndpoint(endpoint),
		},
	}

	group := server.engine.Group(fmt.Sprintf("/%s", name))
	group.Use(createCORSMiddleware(endpoint))
	group.Use(server.createAuthMiddleware(endpoint))

	for _, h := range handlers {
		hLimit, err := h.limitFunc()
		if err != nil {
			return err
		}
		if !hLimit.Disabled {
			group.Handle(
				h.httpMethod,
				h.url,
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
				server.logger.Tracef("invalid auth input: %q", bearerTokenHeader)
				_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid Authorization header input"))
				return
			}

			if *endpoint.Cfg.ApiConfig.Bearer != parts[1] {
				server.logger.Tracef("Wrong bearer token: %q", parts[1])
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
			server.logger.WithError(err).Debugf("could not parse input data: %q", rawData)
			return
		}

		if err := endpoint.Validate(rawJson); err != nil {
			c.String(http.StatusBadRequest, "invalid input: %s", err.Error())
			server.logger.WithError(err).Debugf("invalid input data: %q", rawData)
			return
		}

		if err := endpoint.Put(rawJson); err != nil {
			c.String(http.StatusInternalServerError, "could not store requested data")
			server.logger.WithError(err).Errorf("could not store data")
			return
		}

		c.String(http.StatusNoContent, "")
	}
}
func (server *httpServer) createDeleteEndpoint(endpoint Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := endpoint.Delete(id); err != nil {
			server.logger.WithError(err).Info("could not delete entity")
			c.String(http.StatusNotFound, "could not find an entity with id %q", id)
			return
		}

		c.String(http.StatusNoContent, "")
	}
}

func (server *httpServer) registerPrometheus() {
	p := ginprometheus.NewPrometheus("gin")
	p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.FullPath()
		return url
	}
	p.Use(server.engine)
	server.logger.Trace("Prometheus handler registered")
}

func (server *httpServer) registerLogLevelHandler() {
	apiToken := os.Getenv("LOG_LEVEL_API_KEY")

	if apiToken != "" {
		server.engine.POST("/log_level", func(c *gin.Context) {
			bearerTokenHeader := c.GetHeader("Authorization")
			parts := strings.Split(bearerTokenHeader, " ")
			if len(parts) != 2 {
				_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid Authorization header input"))
				c.Abort()
				return
			}

			if apiToken != parts[1] {
				_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid Authorization bearer token"))
				c.Abort()
				return
			}

			var data struct {
				Level string `json:"level"`
			}

			if err := c.BindJSON(&data); err != nil {
				c.String(http.StatusBadRequest, "could not decode input json (level field is required)")
				c.Abort()
				return
			}

			level, err := logrus.ParseLevel(data.Level)
			if err != nil {
				c.String(http.StatusBadRequest, "%q is not supported log level", data.Level)
				c.Abort()
				return
			}

			server.logger.SetLevel(level)
			c.String(http.StatusNoContent, "")
			c.Abort()
			server.logger.Infof("Log level changed to %q", level.String())
		})

		server.logger.Trace("Log level change endpoint created")
	} else {
		server.logger.Trace("Log level change endpoint not going to work because there is no LOG_LEVEL_API_KEY environment variable")
	}
}

func (server *httpServer) registerIndexHandler() {
	server.engine.GET("/", func(c *gin.Context) {
		var services []string
		for _, endpoint := range server.endpoints {
			services = append(services, endpoint.Cfg.Name)
		}

		response := struct {
			Services []string `json:"services"`
			LogLevel string   `json:"log_level"`
		}{
			Services: services,
			LogLevel: server.logger.Level.String(),
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
			server.logger.Infof("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	server.logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		server.logger.Fatal("Server forced to shutdown:", err)
	}

	server.logger.Info("Server exiting...")
}

func createRateLimiterMiddleware(limit Limit) gin.HandlerFunc {
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

func createCORSMiddleware(endpoint Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := "*"
		if endpoint.Cfg.ApiConfig.Client != nil && len(*endpoint.Cfg.ApiConfig.Client) > 0 {
			origin = *endpoint.Cfg.ApiConfig.Client
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func notFound(c *gin.Context) {
	c.String(404, "not found")
}
