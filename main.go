package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPort    = 8080
	defaultHost    = "127.0.0.1" // bind to localhost only by default; use -host 0.0.0.0 to expose to LAN
	appVersion     = "1.0.0"
	appName        = "sub2api"
)

// Config holds the application configuration
type Config struct {
	Host    string
	Port    int
	Debug   bool
	Token   string
}

func main() {
	cfg := parseConfig()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := setupRouter(cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("[%s] v%s starting on %s", appName, appVersion, addr)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// parseConfig reads configuration from flags and environment variables.
// Environment variables take precedence over default values;
// command-line flags take highest precedence.
func parseConfig() *Config {
	cfg := &Config{}

	// Resolve defaults from environment, falling back to constants
	defHost := getEnv("SUB2API_HOST", defaultHost)
	defPort := defaultPort
	if p, err := strconv.Atoi(getEnv("SUB2API_PORT", "")); err == nil && p > 0 {
		defPort = p
	}
	defDebug := getEnv("SUB2API_DEBUG", "false") == "true"
	defToken := getEnv("SUB2API_TOKEN", "")

	flag.StringVar(&cfg.Host, "host", defHost, "Listening host")
	flag.IntVar(&cfg.Port, "port", defPort, "Listening port")
	flag.BoolVar(&cfg.Debug, "debug", defDebug, "Enable debug mode")
	flag.StringVar(&cfg.Token, "token", defToken, "Optional bearer token for authentication")
	flag.Parse()

	return cfg
}

// setupRouter initialises the Gin engine and registers all routes.
func setupRouter(cfg *Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health / readiness probe
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": appVersion,
		})
	})

	// API v1 group — protected by optional token middleware
	v1 := router.Group("/api/v1")
	if cfg.Token != "" {
		v1.Use(tokenAuthMiddleware(cfg.Token))
	}

	// Subscription conversion endpoint
	v1.GET("/sub", handleSub)

	return router
}

// tokenAuthMiddleware validates a Bearer token when one is configured.
func tokenAuthMiddleware(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != "Bearer "+expectedToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

// handleSub is a placeholder handler for the subscription conversion endpoint.
// Full implementation lives in handler/sub.go.
func handleSub(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
		return
	}
	// TODO: delegate to subscri
