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

	demo, err := sarfya.NewExample(context.Background(), sarfya.Input{
		ID:   "demo",
		Text: "1Kaltxì, 2ulte 3zola'u 4nìprrte' 5fìweptseng6ne! 7(Pamrel si) 8lì'uor 9nefä 10fu 11takuk 12pumit 17a 13mì 14fìpamrel 15fte 16fwivew.",
		LookupFilter: map[int]string{
			11: "vtr.",
		},
		Translations: map[string]string{
			"en": "1Hello, 2and 3+4(welcome) 6to 5(this website)! 7Write 8(a word) 9above 10or 11click 12one 17+13in 14(this writing) 15to 16search.",
		},
		Source: sarfya.Source{},
	}, svc.Dictionary)
	if err != nil {
		panic(err)
	}

	assets, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}

	group.StaticFS("/static/", assets)

	group.GET("/", func(c echo.Context) error {
		return outputHtml(c, http.StatusOK, layoutWrapper(fmt.Sprintf("Sarfya"), indexPage(sarfya.FilterMatch{
			Example:             *demo,
			Selections:          []int{},
			Spans:               [][]int{},
			TranslationAdjacent: map[string][][]int{"en": {}},
			TranslationSpans: map[string][][]int{
				"en": {},
			},
		})))
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
}
