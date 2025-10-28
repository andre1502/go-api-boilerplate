package server

import (
	"fmt"
	boRoute "go-api-boilerplate/api/backoffice/routes"
	platformRoute "go-api-boilerplate/api/platform/routes"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/response"
	"go-api-boilerplate/internal/validation"
	"go-api-boilerplate/module"
	"go-api-boilerplate/module/config"
	"go-api-boilerplate/module/elastic"
	"go-api-boilerplate/module/logger"
	"net/http"
	"slices"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	config        *config.Config
	Echo          *echo.Echo
	api           *echo.Group
	middleware    *middleware.Middleware
	response      *response.Response
	platformRoute *platformRoute.Route
	boRoute       *boRoute.Route
}

func NewServer(
	cfg *config.Config,
	vld *validation.Validation,
	mdl *middleware.Middleware,
	resp *response.Response,
	platformRoute *platformRoute.Route,
	boRoute *boRoute.Route,
) *Server {
	srv := &Server{
		config:        cfg,
		Echo:          echo.New(),
		middleware:    mdl,
		response:      resp,
		platformRoute: platformRoute,
		boRoute:       boRoute,
	}

	srv.init(vld)
	srv.serverRoute()

	// platform
	platformRoute.SetupRoutes(srv.api)

	// backend
	boRoute.SetupRoutes(srv.api)

	return srv
}

func (s *Server) init(vld *validation.Validation) {
	s.api = s.Echo.Group("")
	s.Echo.HTTPErrorHandler = s.CustomErrorHandler
	s.Echo.Validator = vld

	s.Echo.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(false),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	)

	s.Echo.Pre(echoMiddleware.CORS())
	s.Echo.Pre(echoMiddleware.RemoveTrailingSlash())
	s.Echo.Pre(s.middleware.Paginate)
	s.Echo.Use(echoMiddleware.RequestID())
	s.Echo.Use(s.middleware.RequestIPAddress)
	s.Echo.Use(s.middleware.CustomRecover(logger.Log))

	s.monitorBody()
}

func (s *Server) monitorBody() {
	monitorUrls := []string{
		// fmt.Sprintf("/%s", "url_path"),
	}

	if len(monitorUrls) > 0 {
		s.Echo.Use(s.pathBodyDumpMiddleware(monitorUrls...))
	}
}

// PathBodyDumpMiddleware logs the path and body for specific routes.
func (s *Server) pathBodyDumpMiddleware(paths ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if the current request path is in the list of paths to log.
			logPath := false
			for _, p := range paths {
				if c.Path() == p {
					logPath = true
					break
				}
			}

			// If the path is not one we want to log, just proceed.
			if !logPath {
				return next(c)
			}

			// If the path should be logged, we use the body dump middleware.
			return echoMiddleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
				transactionUrls := []string{
					// fmt.Sprintf("/%s", "url_path"),
				}

				logFields := map[string]interface{}{
					"request":  reqBody,
					"response": resBody,
				}

				if slices.Contains(transactionUrls, c.Path()) {
					logFields["elastic_index"] = elastic.ELASTIC_TRANSACTION_ACTIVITY_INDEX
				}

				logger.Log.WithFields(logFields).Info(fmt.Sprintf("Body Dump for URL: %s, Request Body: %s, Response Body: %s.", c.Request().URL, string(reqBody), string(resBody)))
			})(next)(c)
		}
	}
}

func (s *Server) serverStatus() string {
	platformHealth := s.platformRoute.GetHealth()
	boHealth := s.boRoute.GetHealth()

	if platformHealth && boHealth && s.config.REPO_NAME == "go-api" {
		return "ok"
	}

	return "error"
}

func (s *Server) serverRoute() {
	s.api.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, s.config.APP_NAME)
	})

	s.api.GET("/ok", func(c echo.Context) error {
		return c.String(http.StatusOK, s.config.APP_NAME)
	})

	s.api.GET("/health", func(c echo.Context) error {
		c.Response().Header().Set(module.HEADER_CONTENT_TYPE, module.APPLICATION_JSON)

		return c.JSONPretty(200, map[string]interface{}{
			"status": s.serverStatus(),
		}, "  ")
	})

	s.api.GET("/info", func(c echo.Context) error {
		c.Response().Header().Set(module.HEADER_CONTENT_TYPE, module.APPLICATION_JSON)

		platformHealth := s.platformRoute.GetHealth()
		boHealth := s.boRoute.GetHealth()

		data := map[string]interface{}{
			"status":      s.serverStatus(),
			"hostname":    s.config.HOST_NAME,
			"repo_name":   s.config.REPO_NAME,
			"branch_name": s.config.BRANCH_NAME,
			"commit_hash": s.config.COMMIT_HASH,
			"build_date":  s.config.BUILD_DATE,
			"start_time":  s.config.START_TIME,
			"version":     s.config.VERSION,
			"app_name":    s.config.APP_NAME,
			"fe_service":  platformHealth,
			"be_service":  boHealth,
			"date":        time.Now().Format(time.DateTime),
			"client_ip":   module.GetClientIP(c.Request()),
		}

		if !module.IsEmptyString(s.config.POD_ID) {
			data["pod_id"] = s.config.POD_ID
		}

		if !module.IsEmptyString(s.config.POD_NAME) {
			data["pod_name"] = s.config.POD_NAME
		}

		return c.JSONPretty(200, data, "  ")
	})
}
