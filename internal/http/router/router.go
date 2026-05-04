package router

import (
	"net/http"
	"time"

	"github.com/felip/api-fidelidade/internal/auth"
	"github.com/felip/api-fidelidade/internal/config"
	"github.com/felip/api-fidelidade/internal/http/handler"
	"github.com/felip/api-fidelidade/internal/http/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Config         config.Config
	TokenManager   *auth.TokenManager
	UserHandler    *handler.UserHandler
	ProgramHandler *handler.ProgramHandler
	StampHandler   *handler.StampHandler
	QRHandler      *handler.QRHandler
}

const redocHTML = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>API de Programa de Fidelidade - Documentacao</title>
	<style>
		body {
			margin: 0;
			padding: 0;
		}
	</style>
</head>
<body>
	<redoc spec-url="/openapi.yaml"></redoc>
	<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>
`

func New(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Content-Length", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:   []string{"Content-Length", "Content-Type"},
		MaxAge:          12 * time.Hour,
	}))
	router.MaxMultipartMemory = deps.Config.MaxUploadSizeBytes
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/openapi.yaml", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "application/yaml; charset=utf-8")
		ctx.File("openapi.yaml")
	})
	router.GET("/docs", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(redocHTML))
	})
	router.GET("/docs/", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(redocHTML))
	})
	router.StaticFS("/uploads", http.Dir(deps.Config.UploadDir))

	v1 := router.Group("/v1")
	v1.POST("/usuarios", deps.UserHandler.Criar)
	v1.POST("/usuarios/autenticacao", deps.UserHandler.Autenticar)

	protected := v1.Group("")
	// protected.Use(middleware.RequireAuth(deps.TokenManager))

	protected.GET("/programas", deps.ProgramHandler.ListarProximos)
	protected.GET("/programas/:programaId", deps.ProgramHandler.ObterDetalhe)
	protected.POST("/qr-codes/validar", deps.QRHandler.Validar)

	userScoped := protected.Group("/usuarios/:usuarioId")
	userScoped.Use(middleware.RequireUserMatch("usuarioId"))
	userScoped.GET("", deps.UserHandler.Obter)
	userScoped.PUT("", deps.UserHandler.Atualizar)
	userScoped.DELETE("/excluir", deps.UserHandler.Excluir)
	userScoped.POST("/imagem", func(ctx *gin.Context) {
		deps.UserHandler.UploadImagem(ctx, deps.Config.MaxUploadSizeBytes)
	})
	userScoped.GET("/programas/ativos", deps.ProgramHandler.ListarAtivos)
	userScoped.GET("/programas/finalizados", deps.ProgramHandler.ListarFinalizados)
	userScoped.GET("/programas/ultimo-selo", deps.ProgramHandler.ObterUltimoComSelo)
	userScoped.GET("/programas/:programaId/selos", deps.StampHandler.ListarPorPrograma)
	userScoped.POST("/programas/:programaId/selos", deps.StampHandler.Ganhar)
	userScoped.GET("/recompensas", deps.ProgramHandler.ListarRecompensas)
	userScoped.POST("/programas/:programaId/resgatar", deps.ProgramHandler.ResgatarRecompensa)

	return router
}
