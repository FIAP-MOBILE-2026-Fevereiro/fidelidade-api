package router

import (
	"net/http"

	"github.com/felip/api-fidelidade/internal/auth"
	"github.com/felip/api-fidelidade/internal/config"
	"github.com/felip/api-fidelidade/internal/http/handler"
	"github.com/felip/api-fidelidade/internal/http/middleware"
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

func New(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.MaxMultipartMemory = deps.Config.MaxUploadSizeBytes
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.StaticFS("/uploads", http.Dir(deps.Config.UploadDir))

	v1 := router.Group("/v1")
	v1.POST("/usuarios", deps.UserHandler.Criar)
	v1.POST("/usuarios/autenticacao", deps.UserHandler.Autenticar)

	protected := v1.Group("")
	protected.Use(middleware.RequireAuth(deps.TokenManager))

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
