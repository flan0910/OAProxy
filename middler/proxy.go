package middler

import (
	"net/http"
	"net/url"
	"fmt"

	"github.com/flan0910/OAProxy/modules"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func MiddleProx(e echo.Echo) {
	config := modules.ConfigLoad()

	for _, v := range config.Server {
		serverMap := v.(map[interface{}]interface{})

		g := e.Group(serverMap["location"].(string))
		urls, err := url.Parse(serverMap["address"].(string))
		if err != nil {
			modules.Logger("error", err.Error())
		}

		target := []*middleware.ProxyTarget{
			{
				URL: urls,
			},
		}

		g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error { //session
				seslogin := ReadSession(c, "login")
				sesguild := ReadSession(c, "guild")
				
				if seslogin == "true" {
					if sesguild == "true" {
						return next(c)
					}else {
						return c.String(http.StatusForbidden, "NoGuilds")
					}
				}else {
					WriteSession(c, "urled", c.Request().URL.Path)
					return c.Redirect(http.StatusFound, fmt.Sprintf("/%s/login",config.Prefix))
				}
			}
		}, func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error { //privert
				if serverMap["privert"].(bool) == true {
					if modules.RoleTest(serverMap["access_roles"], ReadSession(c, "role")) {
						return next(c)
					}else {
						return c.String(http.StatusForbidden, "NoRoles")
					}
				}
				return next(c)
			}
		}, middleware.Proxy(middleware.NewRandomBalancer(target)))
	}
}
