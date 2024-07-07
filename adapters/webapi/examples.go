package webapi

import (
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya/service"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"net/url"
	"time"
)

func Examples(group *echo.Group, svc *service.Service) {
	group.GET("/:search", func(c echo.Context) error {
		search, err := url.QueryUnescape(c.Param("search"))
		if err != nil {
			return err
		}

		start := time.Now()
		res, err := svc.QueryExample(c.Request().Context(), search)
		if err != nil {
			return err
		}
		log.Printf("Query %#+v exectured in %s", search, time.Since(start))

		return c.JSON(http.StatusOK, map[string]any{
			"examples":    res,
			"executionMs": float64(time.Since(start)) / float64(time.Millisecond),
		})
	})

	group.GET("/:id/input", func(c echo.Context) error {
		example, err := svc.FindExample(c.Request().Context(), c.Param("id"))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{
			"input": example.Input(),
		})
	})

	group.POST("/", func(c echo.Context) error {
		input := sarfya.Input{}
		if err := c.Bind(&input); err != nil {
			return err
		}
		if input.Text == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Text cannot be left blank",
			})
		}

		if input.Source.ID == "" || input.Source.Date == "" || input.Source.URL == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Missing fields in source",
			})
		}

		res, err := svc.SaveExample(c.Request().Context(), input, c.QueryParam("dry") == "true")
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{
			"example": *res,
		})
	})

	group.DELETE("/:id", func(c echo.Context) error {
		example, err := svc.DeleteExample(c.Request().Context(), c.Param("id"))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{
			"example": example,
		})
	})

}
