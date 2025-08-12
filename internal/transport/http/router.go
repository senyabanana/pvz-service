package http

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/senyabanana/pvz-service/internal/handler"
	"github.com/senyabanana/pvz-service/internal/infrastructure/monitoring"
	"github.com/senyabanana/pvz-service/internal/middleware"
)

const (
	moderatorRole = "moderator"
	employeeRole  = "employee"
)

func SetupRouter(handlers *handler.Handler, secretKey string, log *logrus.Logger) *gin.Engine {
	monitoring.RegisterMetrics()
	router := gin.Default()
	router.Use(middleware.PrometheusMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/dummyLogin", handlers.Authorization.DummyLogin)
	router.POST("/register", handlers.Authorization.Register)
	router.POST("/login", handlers.Authorization.Login)

	moderator := router.Group("/")
	moderator.Use(middleware.RequireRole(secretKey, log, moderatorRole))
	{
		moderator.POST("/pvz", handlers.PVZOperations.CreatePVZ)
	}

	employee := router.Group("/")
	employee.Use(middleware.RequireRole(secretKey, log, employeeRole))
	{
		employee.POST("/pvz/:pvzId/close_last_reception", handlers.ReceptionOperations.CloseLastReception)
		employee.POST("/pvz/:pvzId/delete_last_product", handlers.ProductOperations.DeleteLastProduct)
		employee.POST("/receptions", handlers.ReceptionOperations.CreateReception)
		employee.POST("/products", handlers.ProductOperations.AddProduct)
	}

	staff := router.Group("/")
	staff.Use(middleware.RequireRole(secretKey, log, moderatorRole, employeeRole))
	{
		staff.GET("/pvz", handlers.PVZOperations.GetFullInfoPVZ)
	}

	return router
}
