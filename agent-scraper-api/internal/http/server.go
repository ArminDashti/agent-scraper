package httpserver

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArminDashti/agent-scraper-api/internal/auth"
	"github.com/ArminDashti/agent-scraper-api/internal/config"
	"github.com/ArminDashti/agent-scraper-api/internal/pipeline"
	"github.com/ArminDashti/agent-scraper-api/internal/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg          config.Config
	auth         *auth.Service
	store        *store.Store
	runner       *pipeline.Runner
	cronSchedule string
}

func New(cfg config.Config, authSvc *auth.Service, st *store.Store, runner *pipeline.Runner, cronSchedule string) *Server {
	return &Server{cfg: cfg, auth: authSvc, store: st, runner: runner, cronSchedule: cronSchedule}
}

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     s.cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		v1.POST("/auth/login", s.postLogin)
	}

	protected := v1.Group("")
	protected.Use(jwtMiddleware(s.auth))
	{
		protected.GET("/jobs", s.getJobs)
		protected.POST("/jobs/run", s.postJobRun)
		protected.GET("/expenses", s.getExpenses)
		protected.GET("/sources", s.getSources)
		protected.POST("/sources", s.postSource)
		protected.PATCH("/sources/:id", s.patchSource)
		protected.DELETE("/sources/:id", s.deleteSource)
		protected.GET("/forwarding", s.getForwarding)
		protected.PUT("/forwarding", s.putForwarding)
	}

	return r
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) postLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := s.auth.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) getJobs(c *gin.Context) {
	runs, err := s.store.ListJobRuns(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cronSchedule": s.cronSchedule, "runs": runs})
}

func (s *Server) postJobRun(c *gin.Context) {
	if err := s.runner.Run(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	runs, err := s.store.ListJobRuns(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "runs": runs})
}

func (s *Server) getExpenses(c *gin.Context) {
	rows, err := s.store.ListExpenses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"expenses": rows})
}

func (s *Server) getSources(c *gin.Context) {
	rows, err := s.store.ListSources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sources": rows})
}

type sourceRequest struct {
	Name       string `json:"name" binding:"required"`
	WebsiteURL string `json:"websiteUrl"`
	IsEnabled  bool   `json:"isEnabled"`
}

func (s *Server) postSource(c *gin.Context) {
	var req sourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	src, err := s.store.CreateSource(c.Request.Context(), strings.TrimSpace(req.Name), strings.TrimSpace(req.WebsiteURL), req.IsEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, src)
}

func (s *Server) patchSource(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req sourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	src, err := s.store.UpdateSource(c.Request.Context(), id, strings.TrimSpace(req.Name), strings.TrimSpace(req.WebsiteURL), req.IsEnabled)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, src)
}

func (s *Server) deleteSource(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := s.store.DeleteSource(c.Request.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) getForwarding(c *gin.Context) {
	state, err := s.store.ForwardingState(c.Request.Context(), s.cfg.TargetAPIURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, state)
}

type forwardingRequest struct {
	TargetAPIURL string `json:"targetApiUrl"`
}

func (s *Server) putForwarding(c *gin.Context) {
	var req forwardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := s.store.SetTargetAPIURL(c.Request.Context(), strings.TrimSpace(req.TargetAPIURL)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	state, err := s.store.ForwardingState(c.Request.Context(), s.cfg.TargetAPIURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusOK, state)
}
