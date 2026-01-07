package order

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type Handler struct {
    service *Service
}

func NewHandler(s *Service) *Handler {
    return &Handler{service: s}
}

func (h *Handler) PayOrder(c *gin.Context) {
    orderID := c.Param("id")
    oid, _ := uuid.Parse(orderID)

    if err := h.service.PayOrder(c.Request.Context(), oid); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "PAID"})
}
