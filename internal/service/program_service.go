package service

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/felip/api-fidelidade/internal/apperror"
	"github.com/felip/api-fidelidade/internal/store/sqlc"
	"github.com/felip/api-fidelidade/internal/support/pgutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProgramService struct {
	queries *sqlc.Queries
}

type ListaProgramasResponse struct {
	Programas    []ProgramaDetalhado `json:"programas"`
	Total        int                 `json:"total"`
	Pagina       int                 `json:"pagina"`
	TotalPaginas int                 `json:"totalPaginas"`
}

type Recompensa struct {
	Nome        string `json:"nome"`
	Imagem      string `json:"imagem,omitempty"`
	Descricao   string `json:"descricao,omitempty"`
	Resgatada   bool   `json:"resgatada,omitempty"`
	DataResgate string `json:"dataResgate,omitempty"`
}

type ProgramaDetalhado struct {
	ID                string     `json:"id,omitempty"`
	ProgramaID        string     `json:"programaId,omitempty"`
	Estabelecimento   string     `json:"estabelecimento"`
	Lat               float64    `json:"lat,omitempty"`
	Lng               float64    `json:"lng,omitempty"`
	Distancia         float64    `json:"distancia,omitempty"`
	SelosNecessarios  int        `json:"selosNecessarios"`
	SelosAtuais       int        `json:"selosAtuais,omitempty"`
	SelosObtidos      int        `json:"selosObtidos,omitempty"`
	SelosRestantes    int        `json:"selosRestantes,omitempty"`
	Progresso         int        `json:"progresso,omitempty"`
	UltimoSelo        string     `json:"ultimoSelo,omitempty"`
	ProximoAoConcluir bool       `json:"proximoAoConcluir,omitempty"`
	Recompensa        Recompensa `json:"recompensa"`
	Descricao         string     `json:"descricao,omitempty"`
	Regras            string     `json:"regras,omitempty"`
	Ativo             bool       `json:"ativo,omitempty"`
	DataInicio        string     `json:"dataInicio,omitempty"`
	DataFim           string     `json:"dataFim,omitempty"`
	Status            string     `json:"status,omitempty"`
	DataFinalizacao   string     `json:"dataFinalizacao,omitempty"`
	DataConclusao     string     `json:"dataConclusao,omitempty"`
	ExpiraEm          string     `json:"expiraEm,omitempty"`
}

type ProgramaComUltimoSelo struct {
	ProgramaID       string `json:"programaId"`
	Estabelecimento  string `json:"estabelecimento"`
	SelosNecessarios int    `json:"selosNecessarios"`
	SelosObtidos     int    `json:"selosObtidos"`
	UltimoSelo       Selo   `json:"ultimoSelo"`
}

type ListaRecompensasResponse struct {
	Recompensas []ProgramaDetalhado `json:"recompensas"`
}

func NewProgramService(queries *sqlc.Queries) *ProgramService {
	return &ProgramService{queries: queries}
}

func (s *ProgramService) ListarProximos(ctx context.Context, userID string, lat float64, lng float64, raio int, pagina int, limite int) (ListaProgramasResponse, error) {
	if pagina < 1 {
		pagina = 1
	}
	if limite < 1 {
		limite = 20
	}
	offset := (pagina - 1) * limite

	total, err := s.queries.CountNearbyPrograms(ctx, sqlc.CountNearbyProgramsParams{Lng: lng, Lat: lat, RadiusMeters: raio})
	if err != nil {
		return ListaProgramasResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível listar os programas", err)
	}

	rows, err := s.queries.ListNearbyPrograms(ctx, sqlc.ListNearbyProgramsParams{
		Lng:          lng,
		Lat:          lat,
		UserID:       userID,
		RadiusMeters: raio,
		OffsetRows:   int32(offset),
		LimitRows:    int32(limite),
	})
	if err != nil {
		return ListaProgramasResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível listar os programas", err)
	}

	programas := make([]ProgramaDetalhado, 0, len(rows))
	for _, row := range rows {
		programas = append(programas, ProgramaDetalhado{
			ID:               row.ID,
			Estabelecimento:  row.MerchantName,
			Lat:              row.Lat,
			Lng:              row.Lng,
			Distancia:        row.Distance,
			SelosNecessarios: int(row.StampGoal),
			SelosAtuais:      int(row.CurrentStamps),
			Recompensa:       recompensa(row.RewardName, row.RewardImageUrl, row.RewardDescription),
			Descricao:        row.Description,
			Regras:           row.Rules,
			Ativo:            row.Active,
			DataInicio:       pgutil.Time(row.StartsAt).Format(time.RFC3339),
			DataFim:          optionalTime(row.EndsAt),
		})
	}

	totalPaginas := 0
	if total > 0 {
		totalPaginas = int(math.Ceil(float64(total) / float64(limite)))
	}

	return ListaProgramasResponse{Programas: programas, Total: int(total), Pagina: pagina, TotalPaginas: totalPaginas}, nil
}

func (s *ProgramService) ObterDetalhe(ctx context.Context, userID string, programID string) (ProgramaDetalhado, error) {
	row, err := s.queries.GetProgramDetail(ctx, sqlc.GetProgramDetailParams{ID: programID, UserID: userID})
	if err != nil {
		if err == pgx.ErrNoRows {
			return ProgramaDetalhado{}, apperror.New(http.StatusNotFound, "NAO_ENCONTRADO", "Recurso não encontrado")
		}

		return ProgramaDetalhado{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível obter o programa", err)
	}

	return ProgramaDetalhado{
		ID:               row.ID,
		Estabelecimento:  row.MerchantName,
		Lat:              row.Lat,
		Lng:              row.Lng,
		SelosNecessarios: int(row.StampGoal),
		SelosAtuais:      int(row.CurrentStamps),
		Recompensa:       recompensa(row.RewardName, row.RewardImageUrl, row.RewardDescription),
		Descricao:        row.Description,
		Regras:           row.Rules,
		Ativo:            row.Active,
		DataInicio:       pgutil.Time(row.StartsAt).Format(time.RFC3339),
		DataFim:          optionalTime(row.EndsAt),
	}, nil
}

func (s *ProgramService) ListarAtivosDoUsuario(ctx context.Context, userID string) ([]ProgramaDetalhado, error) {
	rows, err := s.queries.ListActiveUserPrograms(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível listar os programas ativos", err)
	}

	programas := make([]ProgramaDetalhado, 0, len(rows))
	for _, row := range rows {
		programas = append(programas, ProgramaDetalhado{
			ProgramaID:        row.ProgramID,
			Estabelecimento:   row.MerchantName,
			SelosNecessarios:  int(row.StampGoal),
			SelosObtidos:      int(row.ObtainedStamps),
			SelosRestantes:    int(row.RemainingStamps),
			Progresso:         int(row.Progress),
			UltimoSelo:        pgutil.Time(row.LastStampAt).Format(time.RFC3339),
			ProximoAoConcluir: row.AlmostComplete,
			Recompensa:        recompensa(row.RewardName, row.RewardImageUrl, row.RewardDescription),
		})
	}

	return programas, nil
}

func (s *ProgramService) ListarFinalizadosDoUsuario(ctx context.Context, userID string, status string) ([]ProgramaDetalhado, error) {
	rows, err := s.queries.ListFinishedUserPrograms(ctx, sqlc.ListFinishedUserProgramsParams{UserID: userID, FilterStatus: status})
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível listar os programas finalizados", err)
	}

	programas := make([]ProgramaDetalhado, 0, len(rows))
	for _, row := range rows {
		programas = append(programas, ProgramaDetalhado{
			ProgramaID:      row.ProgramID,
			Estabelecimento: row.MerchantName,
			DataFinalizacao: pgutil.Time(row.FinishedAt).Format(time.RFC3339),
			Status:          row.Status,
			Recompensa: Recompensa{
				Nome:        row.RewardName,
				Imagem:      pgutil.TextOrEmpty(row.RewardImageUrl),
				Descricao:   row.RewardDescription,
				Resgatada:   pgutil.Bool(row.Redeemed),
				DataResgate: optionalTime(row.RedeemedAt),
			},
		})
	}

	return programas, nil
}

func (s *ProgramService) ObterUltimoComSelo(ctx context.Context, userID string) (*ProgramaComUltimoSelo, error) {
	row, err := s.queries.GetLastProgramStamp(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível obter o último selo", err)
	}

	return &ProgramaComUltimoSelo{
		ProgramaID:       row.ProgramID,
		Estabelecimento:  row.MerchantName,
		SelosNecessarios: int(row.StampGoal),
		SelosObtidos:     int(row.ObtainedStamps),
		UltimoSelo: Selo{
			ID:             row.StampID,
			ProgramaID:     row.StampProgramID,
			UsuarioID:      row.UserID,
			DataAquisicao:  pgutil.Time(row.AcquiredAt).Format(time.RFC3339),
			ChaveValidacao: row.ValidationKey,
			Sequencia:      int(row.Sequence),
			Validado:       row.Validated,
		},
	}, nil
}

func (s *ProgramService) ListarRecompensasDisponiveis(ctx context.Context, userID string) (ListaRecompensasResponse, error) {
	rows, err := s.queries.ListAvailableRewards(ctx, userID)
	if err != nil {
		return ListaRecompensasResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível listar as recompensas", err)
	}

	recompensas := make([]ProgramaDetalhado, 0, len(rows))
	for _, row := range rows {
		recompensas = append(recompensas, ProgramaDetalhado{
			ProgramaID:      row.ProgramID,
			Estabelecimento: row.MerchantName,
			Recompensa:      recompensa(row.RewardName, row.RewardImageUrl, row.RewardDescription),
			DataConclusao:   pgutil.Time(row.CompletedAt).Format(time.RFC3339),
			ExpiraEm:        optionalTime(row.ExpiresAt),
		})
	}

	return ListaRecompensasResponse{Recompensas: recompensas}, nil
}

func (s *ProgramService) ResgatarRecompensa(ctx context.Context, userID string, programID string, codigoResgate string) (string, error) {
	_, err := s.queries.GetRewardRedemptionByProgram(ctx, sqlc.GetRewardRedemptionByProgramParams{UserID: userID, ProgramID: programID})
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", apperror.New(http.StatusBadRequest, "RECOMPENSA_INDISPONIVEL", "Recompensa não disponível para resgate")
		}

		return "", apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível resgatar a recompensa", err)
	}

	redeemedAt, err := s.queries.RedeemReward(ctx, sqlc.RedeemRewardParams{UserID: userID, ProgramID: programID, RewardCode: codigoResgate})
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", apperror.New(http.StatusBadRequest, "RECOMPENSA_INDISPONIVEL", "Código de resgate inválido ou recompensa já resgatada")
		}

		return "", apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível resgatar a recompensa", err)
	}

	return pgutil.Time(redeemedAt).Format(time.RFC3339), nil
}

func recompensa(nome string, imagem pgtype.Text, descricao string) Recompensa {
	return Recompensa{Nome: nome, Imagem: pgutil.TextOrEmpty(imagem), Descricao: descricao}
}

func optionalTime(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case time.Time:
		if typed.IsZero() {
			return ""
		}
		return typed.UTC().Format(time.RFC3339)
	case pgtype.Timestamptz:
		if !typed.Valid {
			return ""
		}
		return typed.Time.UTC().Format(time.RFC3339)
	default:
		return ""
	}
}
