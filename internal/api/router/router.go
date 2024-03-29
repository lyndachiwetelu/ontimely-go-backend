package router

import (
	"fmt"
	"io"
	"os"

	"github.com/antonioalfa22/go-rest-template/internal/api/controllers"
	"github.com/antonioalfa22/go-rest-template/internal/api/middlewares"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup() *gin.Engine {
	app := gin.New()

	// Logging to a file.
	f, _ := os.Create("log/api.log")
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(f)

	// Middlewares
	app.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - - [%s] \"%s %s %s %d %s \" \" %s\" \" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	app.Use(gin.Recovery())
	app.Use(middlewares.CORS())
	app.NoRoute(middlewares.NoRouteHandler())

	// Health check
	app.GET("/api/health", controllers.Health)

	// Routes
	// ================== Auth Routes
	app.POST("/auth/login/google", controllers.GoogleLogin)
	app.POST("/auth/login/validate", controllers.ValidateLoggedIn)

	// ================== Api Calendar Routes
	app.GET("/calendar/authorize/google", controllers.HandleGoogleAuthorizeCalendar)

	authorized := app.Group("/")
	authorized.Use(middlewares.LoginRequired())
	{
		authorized.POST("/calendar/google/add", controllers.ConnectGoogleCalendar)
		authorized.GET("/user/connected-calendars", controllers.GetUserCalendars)
		authorized.GET("/user/connected-calendars/id", controllers.GetUserCalendarByID)
		authorized.PUT("/user/calendars/update", controllers.UpdateUserCalendarDetail)

		authorized.POST("/calendar/outlook/add", controllers.GoogleLogin)
	}

	// ================== Docs Routes
	app.GET("/documentation-route/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return app
}
