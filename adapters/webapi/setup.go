package webapi

import (
	"errors"
	"fmt"
	"github.com/gissleh/sarfya"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

func SetupWithoutListener() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())
	e.HTTPErrorHandler = wrapError

	return e
}

func Setup(addr string) (*echo.Echo, <-chan error) {
	e := SetupWithoutListener()

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
	var exampleErr sarfya.ExampleError
	var filterParseError sarfya.FilterParseError

	switch {
	case errors.As(err, &httpErr):
		_ = c.JSON(httpErr.Code, map[string]string{"error": fmt.Sprint(httpErr.Message)})
	case errors.As(err, &bindingErr):
		_ = c.JSON(bindingErr.Code, map[string]string{"error": fmt.Sprint(bindingErr.Message)})
	case errors.As(err, &exampleErr):
		_ = c.JSON(http.StatusBadRequest, map[string]any{
			"error":     exampleErr.Error(),
			"errorData": exampleErr,
		})
	case errors.As(err, &filterParseError):
		_ = c.JSON(http.StatusBadRequest, map[string]any{
			"error":     filterParseError.Error(),
			"errorData": filterParseError,
		})
	case errors.Is(err, sarfya.ErrReadOnly):
		_ = c.JSON(http.StatusPreconditionFailed, map[string]string{"error": err.Error()})
	case errors.Is(err, sarfya.ErrExampleNotFound), errors.Is(err, sarfya.ErrDictionaryEntryNotFound):
		_ = c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		_ = c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}
