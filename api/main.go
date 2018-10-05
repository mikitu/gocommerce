package main

import (
	"github.com/labstack/echo"
	md "github.com/labstack/echo/middleware"
	"github.com/mikitu/gocommerce/api/src/route"
)

func main() {
	e := echo.New()
	e.Pre(md.AddTrailingSlash())
	e.Use(md.CORS());

	route.ImportBackendRoutes(e)
	route.ImportFrontendRoutes(e)
	e.Logger.Fatal(e.Start(":1323"))
}

