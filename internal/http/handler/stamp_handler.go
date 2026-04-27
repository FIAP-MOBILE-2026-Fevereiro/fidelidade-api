package handler

import (
	"net/http"

	"github.com/felip/api-fidelidade/internal/apperror"
	"github.com/felip/api-fidelidade/internal/http/response"
	"github.com/felip/api-fidelidade/internal/service"
	"github.com/gin-gonic/gin"
)

type StampHandler struct {
	stamps *service.StampService
}

func NewStampHandler(stamps *service.StampService) *StampHandler {
	return &StampHandler{stamps: stamps}
}

func (h *StampHandler) ListarPorPrograma(ctx *gin.Context) {
	result, err := h.stamps.ListarPorPrograma(ctx.Request.Context(), ctx.Param("usuarioId"), ctx.Param("programaId"))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *StampHandler) Ganhar(ctx *gin.Context) {
	var body struct {
		QRCodeData string `json:"qrCodeData"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil || body.QRCodeData == "" {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Dados inválidos"))
		return
	}

	result, err := h.stamps.GanharSelo(ctx.Request.Context(), ctx.Param("usuarioId"), ctx.Param("programaId"), body.QRCodeData)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, result)
}
