package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/felip/api-fidelidade/internal/apperror"
	"github.com/felip/api-fidelidade/internal/http/response"
	"github.com/felip/api-fidelidade/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Criar(ctx *gin.Context) {
	var input service.CriarUsuarioInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Dados inválidos"))
		return
	}

	user, err := h.users.Criar(ctx.Request.Context(), input)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Autenticar(ctx *gin.Context) {
	var input service.AutenticacaoInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Dados inválidos"))
		return
	}

	result, err := h.users.Autenticar(ctx.Request.Context(), input)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *UserHandler) Obter(ctx *gin.Context) {
	user, err := h.users.Obter(ctx.Request.Context(), ctx.Param("usuarioId"))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (h *UserHandler) Atualizar(ctx *gin.Context) {
	var input service.AtualizarUsuarioInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Dados inválidos"))
		return
	}

	user, err := h.users.Atualizar(ctx.Request.Context(), ctx.Param("usuarioId"), input)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (h *UserHandler) Excluir(ctx *gin.Context) {
	if err := h.users.Excluir(ctx.Request.Context(), ctx.Param("usuarioId")); err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *UserHandler) UploadImagem(ctx *gin.Context, maxUploadSize int64) {
	fileHeader, err := ctx.FormFile("imagem")
	if err != nil {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Arquivo de imagem é obrigatório"))
		return
	}
	if fileHeader.Size > maxUploadSize {
		response.Error(ctx, apperror.New(http.StatusRequestEntityTooLarge, "ARQUIVO_MUITO_GRANDE", "Imagem muito grande"))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(ctx, apperror.Wrap(http.StatusBadRequest, "VALIDACAO_FALHA", "Não foi possível ler a imagem", err))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(ctx, apperror.Wrap(http.StatusBadRequest, "VALIDACAO_FALHA", "Não foi possível ler a imagem", err))
		return
	}

	contentType := http.DetectContentType(data)
	if contentType != "image/jpeg" && contentType != "image/png" {
		response.Error(ctx, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Formato de imagem inválido"))
		return
	}

	imageURL, err := h.users.UploadImagem(ctx.Request.Context(), ctx.Param("usuarioId"), contentType, bytes.Clone(data))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"imagemUrl": imageURL})
}
