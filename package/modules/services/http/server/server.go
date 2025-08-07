package server

import (
	"fmt"
	"net/http"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type httpServer struct {
	appConfig    config.AppConfig
	serverConfig config.HTTPConfig

	log *monitoring.Logger

	app *fiber.App

	mux *http.ServeMux
}

type HTTPServer interface {
	Start() error
	Stop() error

	GetApp() *fiber.App
	Router() *http.ServeMux
}

type HttpServerStartHook fx.Hook

type HTTPServerParams struct {
	fx.In

	AppConfig    config.AppConfig
	ServerConfig config.HTTPConfig

	Logger *monitoring.Logger
}

func New(params HTTPServerParams) HTTPServer {
	if !params.ServerConfig.Enabled {
		return nil
	}

	// For better compatibility
	mux := http.NewServeMux()
	app := createFiberApp(params.Logger)

	return &httpServer{
		appConfig:    params.AppConfig,
		serverConfig: params.ServerConfig,

		log: params.Logger,

		app: app,
		mux: mux,
	}
}

func createFiberApp(logger *monitoring.Logger) *fiber.App {
	app := fiber.New()

	// Set fiber logger, then we can use fiber log everywhere``
	flog := fiberzap.NewLogger(fiberzap.LoggerConfig{
		SetLogger: logger.GetLogger(),
	})

	log.SetLogger(flog)

	return app
}

func (h *httpServer) GetApp() *fiber.App {
	return h.app
}

func (h *httpServer) Start() error {
	// Start listen http request
	h.log.Info("HTTP server is listening on port", zap.Int("port", h.serverConfig.Port))

	h.mux.HandleFunc("/", adaptor.FiberApp(h.app))
	var handler http.Handler = h.mux
	return http.ListenAndServe(fmt.Sprintf(":%d", h.serverConfig.Port), handler)
}

func (h *httpServer) Stop() error {
	h.log.Info("HTTP server shutting down...")
	return h.app.Shutdown()
}

func (h *httpServer) Router() *http.ServeMux {
	return h.mux
}
