package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"green-irrig/internal/service"
)

type IrrigationHandler struct {
	svc *service.IrrigationService
}

func NewIrrigationHandler(svc *service.IrrigationService) *IrrigationHandler {
	return &IrrigationHandler{svc: svc}
}

func (h *IrrigationHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.POST("/zones/:id/irrigate", h.ScheduleIrrigation)
	api.GET("/zones/:id/plans", h.ListPlans)
	api.POST("/zones/:id/readings", h.RecordReading)
	api.POST("/plans/:id/execute", h.ExecuteIrrigation)
}

func (h *IrrigationHandler) ScheduleIrrigation(c *gin.Context) {
	zoneID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}
	var req struct {
		Volume float64 `json:"volume_liters"`
		Hours  int     `json:"delay_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	when := time.Now().Add(time.Duration(req.Hours) * time.Hour)
	plan, err := h.svc.ScheduleIrrigation(c.Request.Context(), uint(zoneID), req.Volume, when)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *IrrigationHandler) ListPlans(c *gin.Context) {
	zoneID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}
	plans, err := h.svc.ListPlans(c.Request.Context(), uint(zoneID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *IrrigationHandler) RecordReading(c *gin.Context) {
	zoneID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}
	var req struct {
		SoilMoist float64 `json:"soil_moist"`
		TempC     float64 `json:"temp_c"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reading, err := h.svc.RecordReading(c.Request.Context(), uint(zoneID), req.SoilMoist, req.TempC)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reading)
}

func (h *IrrigationHandler) ExecuteIrrigation(c *gin.Context) {
	planID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}
	if err := h.svc.ExecuteIrrigation(c.Request.Context(), uint(planID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
