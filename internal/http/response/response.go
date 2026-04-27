package response

import (
	"net/http"

	"github.com/felip/api-fidelidade/internal/apperror"
	"github.com/gin-gonic/gin"
)

func Error(ctx *gin.Context, err error) {
	if appErr, ok := apperror.As(err); ok {
		payload := gin.H{
			"codigo":   appErr.Codigo,
			"mensagem": appErr.Mensagem,
		}
		if len(appErr.Detalhes) > 0 {
			payload["detalhes"] = appErr.Detalhes
		}

		ctx.AbortWithStatusJSON(appErr.Status, payload)
		return
	}

	ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"codigo":   "ERRO_INTERNO",
		"mensagem": "Erro interno do servidor",
	})
}
