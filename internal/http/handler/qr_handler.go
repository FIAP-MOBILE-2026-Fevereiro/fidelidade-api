package handler

import (
	"net/http"

	"github.com/felip/api-fidelidade/internal/apperror"
	"github.com/felip/api-fidelidade/internal/http/middleware"
	"github.com/felip/api-fidelidade/internal/http/response"
	"github.com/felip/api-fidelidade/internal/service"
	"github.com/gin-gonic/gin"
)

type QRHandler struct {
	stamps *service.StampService
}

func NewQRHandler(stamps *service.StampService) *QRHandler {
	return &QRHandler{stamps: stamps}
}

func (h *QRHandler) Validar(ctx *gin.Context) {
	var body struct {
		QRCodeData string `json:"qrCodeData"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil || body.QRCodeData == "" {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Dados inválidos"))
		return
	}

	result, err := h.stamps.ValidarQRCode(ctx.Request.Context(), middleware.CurrentUserID(ctx), body.QRCodeData)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}
