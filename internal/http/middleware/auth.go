package middleware

import (
	"net/http"
	"strings"

	"github.com/felip/api-fidelidade/internal/apperror"
	"github.com/felip/api-fidelidade/internal/auth"
	"github.com/felip/api-fidelidade/internal/http/response"
	"github.com/gin-gonic/gin"
)

const authUserIDKey = "authUserID"

func RequireAuth(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := strings.TrimSpace(ctx.GetHeader("Authorization"))
		if header == "" {
			response.Error(ctx, apperror.New(http.StatusUnauthorized, "NAO_AUTORIZADO", "Token inválido ou expirado"))
			return
		}

		userID, err := tokenManager.Parse(header)
		if err != nil {
			response.Error(ctx, apperror.Wrap(http.StatusUnauthorized, "NAO_AUTORIZADO", "Token inválido ou expirado", err))
			return
		}

		ctx.Set(authUserIDKey, userID)
		ctx.Next()
	}
}

func RequireUserMatch(paramName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authUserID := CurrentUserID(ctx)
		if authUserID == "" {
			response.Error(ctx, apperror.New(http.StatusUnauthorized, "NAO_AUTORIZADO", "Token inválido ou expirado"))
			return
		}

		if ctx.Param(paramName) != authUserID {
			response.Error(ctx, apperror.New(http.StatusForbidden, "ACESSO_PROIBIDO", "Você não tem permissão para acessar este recurso"))
			return
		}

		ctx.Next()
	}
}

func CurrentUserID(ctx *gin.Context) string {
	value, _ := ctx.Get(authUserIDKey)
	userID, _ := value.(string)
	return userID
}
