package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
	"github.com/reynerpantou/bookcabin/internal/usecase"
	flightsearchusecase "github.com/reynerpantou/bookcabin/internal/usecase/flight-search"
)

type HTTP struct {
	router              *gin.Engine
	flightSearchUseCase flightsearchusecase.UseCase
}

func NewHTTP(useCases usecase.UseCases) *HTTP {
	h := &HTTP{
		router:              gin.New(),
		flightSearchUseCase: useCases.FlightSearch,
	}
	h.router.Use(gin.Logger(), gin.Recovery())
	h.router.GET("/search/flight/v1", h.FlightSearch)
	h.router.POST("/search/flight/v1", h.FlightSearch)
	h.router.GET("/health", h.Health)
	return h
}

func (h *HTTP) Router() *gin.Engine {
	return h.router
}

func (h *HTTP) FlightSearch(c *gin.Context) {
	var params *requestparamsmodel.RequestParams
	var err error

	// prepare parameters
	switch c.Request.Method {
	case http.MethodGet:
		err = c.ShouldBindQuery(&params)
	case http.MethodPost:
		err = c.ShouldBindJSON(&params)
	default:
		c.Status(http.StatusMethodNotAllowed)
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	params.Normalize()
	result, err := h.flightSearchUseCase.FlightSearch(
		c.Request.Context(),
		params,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to search flights",
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTP) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
