package handler

import (
	"net/http"
	"strconv"

	"github.com/felip/api-fidelidade/internal/apperror"
	"github.com/felip/api-fidelidade/internal/http/middleware"
	"github.com/felip/api-fidelidade/internal/http/response"
	"github.com/felip/api-fidelidade/internal/service"
	"github.com/gin-gonic/gin"
)

type ProgramHandler struct {
	programs *service.ProgramService
}

func NewProgramHandler(programs *service.ProgramService) *ProgramHandler {
	return &ProgramHandler{programs: programs}
}

func (h *ProgramHandler) ListarProximos(ctx *gin.Context) {
	lat, err := strconv.ParseFloat(ctx.Query("lat"), 64)
	if err != nil {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Latitude inválida"))
		return
	}
	lng, err := strconv.ParseFloat(ctx.Query("lng"), 64)
	if err != nil {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Longitude inválida"))
		return
	}
	raio := parseInt(ctx.DefaultQuery("raio", "5000"), 5000)
	pagina := parseInt(ctx.DefaultQuery("pagina", "1"), 1)
	limite := parseInt(ctx.DefaultQuery("limite", "20"), 20)

	result, err := h.programs.ListarProximos(ctx.Request.Context(), middleware.CurrentUserID(ctx), lat, lng, raio, pagina, limite)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *ProgramHandler) ObterDetalhe(ctx *gin.Context) {
	program, err := h.programs.ObterDetalhe(ctx.Request.Context(), middleware.CurrentUserID(ctx), ctx.Param("programaId"))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, program)
}

func (h *ProgramHandler) ListarAtivos(ctx *gin.Context) {
	programs, err := h.programs.ListarAtivosDoUsuario(ctx.Request.Context(), ctx.Param("usuarioId"))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"programas": programs})
}

func (h *ProgramHandler) ListarFinalizados(ctx *gin.Context) {
	programs, err := h.programs.ListarFinalizadosDoUsuario(ctx.Request.Context(), ctx.Param("usuarioId"), ctx.Query("status"))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"programas": programs})
}

func (h *ProgramHandler) ObterUltimoComSelo(ctx *gin.Context) {
	program, err := h.programs.ObterUltimoComSelo(ctx.Request.Context(), ctx.Param("usuarioId"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	if program == nil {
		ctx.Status(http.StatusNoContent)
		return
	}

	ctx.JSON(http.StatusOK, program)
}

func (h *ProgramHandler) ListarRecompensas(ctx *gin.Context) {
	rewards, err := h.programs.ListarRecompensasDisponiveis(ctx.Request.Context(), ctx.Param("usuarioId"))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, rewards)
}

func (h *ProgramHandler) ResgatarRecompensa(ctx *gin.Context) {
	var body struct {
		CodigoResgate string `json:"codigoResgate"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil || body.CodigoResgate == "" {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Dados inválidos"))
		return
	}

	dataResgate, err := h.programs.ResgatarRecompensa(ctx.Request.Context(), ctx.Param("usuarioId"), ctx.Param("programaId"), body.CodigoResgate)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"sucesso": true, "dataResgate": dataResgate})
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
