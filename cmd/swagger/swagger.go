package swagger

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/urfave/cli/v2"

	"scheduler/cmd/swagger/auto_generate"
)

var Server = &cli.Command{
	Name:   "doc",
	Usage:  "start swagger server",
	Action: runDoc,
}

// @contact.name   i19.voyager
// @contact.url    http://www.scheduler.io/support
// @contact.email  support@scheduler.io
func runDoc(cCtx *cli.Context) error {
	// programmatically set swagger info
	auto_generate.SwaggerInfo.Title = "scheduler backend API"
	auto_generate.SwaggerInfo.Description = " Handle requests about custom linting rules from scheduler"
	auto_generate.SwaggerInfo.Version = "1.0"
	//swagger.SwaggerInfo.Host = "petstore.swagger.io"
	//swagger.SwaggerInfo.BasePath = "/v2"
	auto_generate.SwaggerInfo.Schemes = []string{"http"}

	r := gin.New()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	println("check swagger http://localhost:8080/swagger/index.html")
	r.Run()
	return nil
}
