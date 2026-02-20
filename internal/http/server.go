package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/INOVA/DML/internal/config"
	"github.com/INOVA/DML/internal/db"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	// Swagger imports
	_ "github.com/INOVA/DML/docs" // Import generated docs

	auditHTTP "github.com/INOVA/DML/internal/http/audit"
	authHTTP "github.com/INOVA/DML/internal/http/auth"
	hrHTTP "github.com/INOVA/DML/internal/http/hr"
	iamHTTP "github.com/INOVA/DML/internal/http/iam"
	orgHTTP "github.com/INOVA/DML/internal/http/org"
	tenancyHTTP "github.com/INOVA/DML/internal/http/tenancy"

	auditLogic "github.com/INOVA/DML/internal/logic/audit"
	authLogic "github.com/INOVA/DML/internal/logic/auth"
	hrLogic "github.com/INOVA/DML/internal/logic/hr"
	iamLogic "github.com/INOVA/DML/internal/logic/iam"
	orgLogic "github.com/INOVA/DML/internal/logic/org"
	tenancyLogic "github.com/INOVA/DML/internal/logic/tenancy"

	"github.com/INOVA/DML/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

// Server represents the HTTP server
type Server struct {
	router *chi.Mux
	db     *db.DB
	config *config.Config
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, database *db.DB) *Server {
	s := &Server{
		router: chi.NewRouter(),
		db:     database,
		config: cfg,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Standard chi middlewares
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RedirectSlashes)

	// Set a timeout value on the request context
	s.router.Use(middleware.Timeout(60 * time.Second))

	// CORS Setup
	c := cors.New(cors.Options{
		AllowedOrigins:   s.config.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	s.router.Use(c.Handler)

	// Health check endpoint
	s.router.Get("/health", s.handleHealthCheck())

	// Swagger UI
	s.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The url pointing to API definition
	))

	// Initialize Services
	auditSvc := auditLogic.NewAuditService(s.db)
	authSvc := authLogic.NewAuthService(s.db, s.config.JWTSecret)
	tenantSvc := tenancyLogic.NewService(s.db)
	buSvc := orgLogic.NewBusinessUnitService(s.db)
	deptSvc := orgLogic.NewDepartmentService(s.db)
	jobSvc := orgLogic.NewJobTitleService(s.db)
	empSvc := hrLogic.NewEmployeeService(s.db, auditSvc)
	onboardSvc := hrLogic.NewOnboardingService(s.db, auditSvc)
	userSvc := iamLogic.NewUserService(s.db, auditSvc)
	userRoleSvc := iamLogic.NewUserRoleService(s.db)
	roleSvc := iamLogic.NewRoleService(s.db, auditSvc)

	// Initialize Handlers
	auditHandler := auditHTTP.NewAuditHandler(auditSvc)
	authHandler := authHTTP.NewAuthHandler(authSvc)
	tenantHandler := tenancyHTTP.NewHandler(tenantSvc)
	buHandler := orgHTTP.NewBusinessUnitHandler(buSvc)
	deptHandler := orgHTTP.NewDepartmentHandler(deptSvc)
	jobHandler := orgHTTP.NewJobTitleHandler(jobSvc)
	empHandler := hrHTTP.NewEmployeeHandler(empSvc)
	onboardHandler := hrHTTP.NewOnboardingHandler(onboardSvc)
	userHandler := iamHTTP.NewUserHandler(userSvc, userRoleSvc)
	roleHandler := iamHTTP.NewRoleHandler(roleSvc)

	// JWT Config
	jwtMiddleware := authHTTP.AuthMiddleware(authHTTP.MiddlewareConfig{
		JWTSecret: s.config.JWTSecret,
	})

	// API version grouping
	s.router.Route("/api/v1", func(r chi.Router) {

		// Public Routes
		r.Route("/auth", func(public chi.Router) {
			authHandler.RegisterRoutes(public)
		})
		r.Route("/tenants", tenantHandler.RegisterRoutes) // Tenants might be public to register

		// Protected Routes
		r.Group(func(protected chi.Router) {
			protected.Use(jwtMiddleware)

			protected.Route("/audit-logs", auditHandler.RegisterRoutes)
			protected.Route("/business-units", buHandler.RegisterRoutes)
			protected.Route("/departments", deptHandler.RegisterRoutes)
			protected.Route("/job-titles", jobHandler.RegisterRoutes)
			protected.Route("/employees", empHandler.RegisterRoutes)
			protected.Route("/onboard", onboardHandler.RegisterRoutes)
			protected.Route("/users", userHandler.RegisterRoutes)
			protected.Route("/roles", roleHandler.RegisterRoutes)
		})
	})
}

func (s *Server) handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ping DB to ensure not just API is up, but DB connection is healthy
		err := s.db.Pool.Ping(r.Context())
		if err != nil {
			log.Printf("Health check DB ping failed: %v", err)
			response.Error(w, http.StatusServiceUnavailable, "Database unavailable")
			return
		}

		response.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"db":     "connected",
		})
	}
}

// Run starts the HTTP server with graceful shutdown
func (s *Server) Run() error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.APIPort),
		Handler: s.router,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		log.Println("Shutting down server...")
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	log.Printf("Starting server on port %s", s.config.APIPort)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	log.Println("Server gracefully stopped.")
	return nil
}
