package webapi

import (
	"errors"
	"fmt"
	"github.com/gissleh/sarfya"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(addr string) (*echo.Echo, <-chan error) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())
	e.HTTPErrorHandler = wrapError

	errCh := make(chan error)
	go func() {
		defer close(errCh)

		err := e.Start(addr)
		if err != nil {
			errCh <- err
		}
	}()

	return e, errCh
}

func wrapError(err error, c echo.Context) {
	var httpErr *echo.HTTPError
	var bindingErr *echo.BindingError

	switch {
	case errors.As(err, &httpErr):
		_ = c.JSON(httpErr.Code, map[string]string{"error": fmt.Sprint(httpErr.Message)})
	case errors.As(err, &bindingErr):
		_ = c.JSON(bindingErr.Code, map[string]string{"error": fmt.Sprint(bindingErr.Message)})
	case errors.Is(err, sarfya.ErrReadOnly):
		_ = c.JSON(502, map[string]string{"error": err.Error()})
	case errors.Is(err, sarfya.ErrExampleNotFound), errors.Is(err, sarfya.ErrDictionaryEntryNotFound):
		_ = c.JSON(404, map[string]string{"error": err.Error()})
	default:
		_ = c.JSON(500, map[string]string{"error": err.Error()})
	}
}
