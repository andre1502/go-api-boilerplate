package cmd

import (
	boHandlers "go-api-boilerplate/api/backoffice/handlers"
	boMiddleware "go-api-boilerplate/api/backoffice/middleware"
	boRoutes "go-api-boilerplate/api/backoffice/routes"
	boServices "go-api-boilerplate/api/backoffice/services"
	platformHandlers "go-api-boilerplate/api/platform/handlers"
	platformMiddleware "go-api-boilerplate/api/platform/middleware"
	platformRoutes "go-api-boilerplate/api/platform/routes"
	platformServices "go-api-boilerplate/api/platform/services"
	"go-api-boilerplate/internal/handlers"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/repositories"
	"go-api-boilerplate/internal/services"

	"go.uber.org/dig"
)

// for internal container
func InitInternalContainer(container *dig.Container) {
	// Repository
	container.Provide(repositories.NewRepository)

	// Service
	container.Provide(services.NewService)

	// Handler
	container.Provide(handlers.NewHandler)

	// Middleware
	container.Provide(middleware.NewMiddleware)
}

// for platform container
func InitPlatformContainer(container *dig.Container) {
	// Service
	container.Provide(platformServices.NewService)

	// Handler
	container.Provide(platformHandlers.NewHandler)

	// Middleware
	container.Provide(platformMiddleware.NewMiddleware)

	// Route
	container.Provide(platformRoutes.NewRoute)
}

// for backoffice container
func InitBackOfficeContainer(container *dig.Container) {
	// Service
	container.Provide(boServices.NewService)

	// Handler
	container.Provide(boHandlers.NewHandler)

	// Middleware
	container.Provide(boMiddleware.NewMiddleware)

	// Route
	container.Provide(boRoutes.NewRoute)
}
