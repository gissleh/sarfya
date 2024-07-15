package templfrontend

import (
	"context"
	"embed"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya/service"
	"github.com/labstack/echo/v4"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"time"
)

//go:embed assets/*
var assets embed.FS

func Endpoints(group *echo.Group, svc *service.Service) {
	outputHtml := func(c echo.Context, code int, component templ.Component) error {
		c.Response().Header().Add("Content-Type", "text/html; charset=utf-8")
		c.Response().WriteHeader(code)
		return component.Render(c.Request().Context(), c.Response())
	}

	demo := createDemo(svc.Dictionary)

	assets, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}

	group.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			lang := c.QueryParam("lang")
			if lang == "" {
				langCookie, err := c.Cookie("lang")
				if err != nil && langCookie != nil {
					lang = langCookie.Value
				} else {
					lang = "en"
				}
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(
				context.WithValue(
					c.Request().Context(),
					langCtxKey,
					lang,
				),
				demoCtxKey,
				demo,
			)))

			return next(c)
		}
	})

	group.StaticFS("/static/", assets)

	group.GET("/", func(c echo.Context) error {
		return outputHtml(c, http.StatusOK, layoutWrapper(fmt.Sprintf("Sarfya"), indexPage()))
	})

	group.GET("/search/:search", func(c echo.Context) error {
		search, err := url.QueryUnescape(c.Param("search"))
		if err != nil {
			return outputHtml(c, http.StatusUnprocessableEntity, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, err.Error(), nil)))
		}

		startTime := time.Now()
		res, err := svc.QueryExample(c.Request().Context(), search)
		if err != nil {
			return outputHtml(c, http.StatusInternalServerError, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, err.Error(), nil)))
		}

		duration := time.Since(startTime)
		if duration > time.Millisecond*30 {
			log.Printf("Slow! %#+v exectured in %s", search, time.Since(startTime))
		}

		return outputHtml(c, 200, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, "", res)))
	})

	group.GET("/search", func(c echo.Context) error {
		search, err := url.QueryUnescape(c.QueryParam("q"))
		if err != nil {
			return outputHtml(c, http.StatusUnprocessableEntity, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, err.Error(), nil)))
		}

		startTime := time.Now()
		res, err := svc.QueryExample(c.Request().Context(), search)
		if err != nil {
			return outputHtml(c, http.StatusInternalServerError, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, err.Error(), nil)))
		}

		duration := time.Since(startTime)
		if duration > time.Millisecond*30 {
			log.Printf("Slow! %#+v exectured in %s", search, time.Since(startTime))
		}

		return outputHtml(c, 200, layoutWrapper(fmt.Sprintf("Sarfya – %s", search), searchPage(search, "", res)))
	})

	group.GET("/fragments/example/:id", func(c echo.Context) error {
		id, err := url.QueryUnescape(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		res, err := svc.FindExample(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": err.Error(),
			})
		}

		filter, resolved, err := sarfya.ParseFilter(c.Request().Context(), c.QueryParam("filter"), svc.Dictionary)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": err.Error(),
			})
		}

		match := filter.CheckExample(*res, resolved[0])
		if match == nil {
			return c.JSON(http.StatusPreconditionFailed, map[string]string{
				"error": "filter did not match on example",
			})
		}

		return outputHtml(c, 200, example(*match, c.Request().Context().Value(langCtxKey).(string)))
	})
}
