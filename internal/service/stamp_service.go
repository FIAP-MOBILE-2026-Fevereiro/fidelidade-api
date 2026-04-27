package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/felip/api-fidelidade/internal/apperror"
	"github.com/felip/api-fidelidade/internal/id"
	"github.com/felip/api-fidelidade/internal/store"
	"github.com/felip/api-fidelidade/internal/store/sqlc"
	"github.com/felip/api-fidelidade/internal/support/hashutil"
	"github.com/felip/api-fidelidade/internal/support/pgutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type StampService struct {
	store *store.Store
}

type QRCodeValidationResponse struct {
	Valido   bool            `json:"valido"`
	Programa *QRCodePrograma `json:"programa,omitempty"`
	Mensagem string          `json:"mensagem,omitempty"`
}

type QRCodePrograma struct {
	ID              string     `json:"id"`
	Estabelecimento string     `json:"estabelecimento"`
	Recompensa      Recompensa `json:"recompensa"`
	SelosRestantes  int        `json:"selosRestantes"`
}

type Selo struct {
	ID             string `json:"id"`
	ProgramaID     string `json:"programaId,omitempty"`
	UsuarioID      string `json:"usuarioId,omitempty"`
	DataAquisicao  string `json:"dataAquisicao"`
	ChaveValidacao string `json:"chaveValidacao"`
	Sequencia      int    `json:"sequencia"`
	Validado       bool   `json:"validado"`
}

type ListaSelosResponse struct {
	Selos            []Selo `json:"selos"`
	TotalSelos       int    `json:"totalSelos"`
	SelosNecessarios int    `json:"selosNecessarios"`
	SelosRestantes   int    `json:"selosRestantes"`
	Progresso        int    `json:"progresso"`
}

type GanharSeloResponse struct {
	Selo              Selo `json:"selo"`
	CompletouPrograma bool `json:"completouPrograma"`
	SelosRestantes    int  `json:"selosRestantes"`
	Progresso         int  `json:"progresso"`
}

func NewStampService(appStore *store.Store) *StampService {
	return &StampService{store: appStore}
}

func (s *StampService) ListarPorPrograma(ctx context.Context, userID string, programID string) (ListaSelosResponse, error) {
	program, err := s.store.Queries.GetProgramDetail(ctx, sqlc.GetProgramDetailParams{ID: programID, UserID: userID})
	if err != nil {
		if err == pgx.ErrNoRows {
			return ListaSelosResponse{}, apperror.New(http.StatusNotFound, "NAO_ENCONTRADO", "Recurso não encontrado")
		}
		return ListaSelosResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível listar os selos", err)
	}

	rows, err := s.store.Queries.ListStampsByProgram(ctx, sqlc.ListStampsByProgramParams{UserID: userID, ProgramID: programID})
	if err != nil {
		return ListaSelosResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível listar os selos", err)
	}

	totalSelos := len(rows)
	stamps := make([]Selo, 0, totalSelos)
	for _, row := range rows {
		stamps = append(stamps, Selo{
			ID:             row.ID,
			ProgramaID:     row.ProgramID,
			UsuarioID:      row.UserID,
			DataAquisicao:  pgutil.Time(row.AcquiredAt).Format(time.RFC3339),
			ChaveValidacao: row.ValidationKey,
			Sequencia:      int(row.Sequence),
			Validado:       row.Validated,
		})
	}

	remaining := int(program.StampGoal) - totalSelos
	if remaining < 0 {
		remaining = 0
	}

	progress := 0
	if program.StampGoal > 0 {
		progress = int((float64(totalSelos) / float64(program.StampGoal)) * 100)
		if progress > 100 {
			progress = 100
		}
	}

	return ListaSelosResponse{Selos: stamps, TotalSelos: totalSelos, SelosNecessarios: int(program.StampGoal), SelosRestantes: remaining, Progresso: progress}, nil
}

func (s *StampService) ValidarQRCode(ctx context.Context, userID string, qrCodeData string) (QRCodeValidationResponse, error) {
	validation, err := s.lookupQRCode(ctx, qrCodeData)
	if err != nil {
		return QRCodeValidationResponse{}, err
	}

	currentStamps, err := s.store.Queries.CountStampsByProgram(ctx, sqlc.CountStampsByProgramParams{UserID: userID, ProgramID: validation.ProgramID})
	if err != nil {
		return QRCodeValidationResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível validar o QR code", err)
	}

	remaining := int(validation.StampGoal - currentStamps)
	if remaining < 0 {
		remaining = 0
	}

	return QRCodeValidationResponse{
		Valido: true,
		Programa: &QRCodePrograma{
			ID:              validation.ProgramID,
			Estabelecimento: validation.MerchantName,
			Recompensa:      recompensa(validation.RewardName, validation.RewardImageUrl, validation.RewardDescription),
			SelosRestantes:  remaining,
		},
	}, nil
}

func (s *StampService) GanharSelo(ctx context.Context, userID string, programID string, qrCodeData string) (GanharSeloResponse, error) {
	validation, err := s.lookupQRCode(ctx, qrCodeData)
	if err != nil {
		return GanharSeloResponse{}, err
	}
	if validation.ProgramID != programID {
		return GanharSeloResponse{}, apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "QR code não pertence ao programa informado")
	}

	qrHash := hashutil.SHA256(qrCodeData)
	tx, err := s.store.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := s.store.Queries.WithTx(tx)
	sequence, err := queries.GetMaxStampSequence(ctx, sqlc.GetMaxStampSequenceParams{UserID: userID, ProgramID: programID})
	if err != nil {
		return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
	}

	if _, err := queries.MarkQRCodeUsed(ctx, sqlc.MarkQRCodeUsedParams{ID: validation.ID, UsedByUserID: pgtype.Text{String: userID, Valid: true}}); err != nil {
		if err == pgx.ErrNoRows {
			return GanharSeloResponse{}, apperror.New(http.StatusConflict, "SELO_JA_REGISTRADO", "Este QR code já foi utilizado")
		}
		return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
	}

	stampID, err := id.New("selo_")
	if err != nil {
		return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
	}
	validationKey, err := id.New("QR")
	if err != nil {
		return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
	}

	created, err := queries.CreateStamp(ctx, sqlc.CreateStampParams{
		ID:            stampID,
		UserID:        userID,
		ProgramID:     programID,
		QrCodeID:      validation.ID,
		QrCodeHash:    qrHash,
		ValidationKey: validationKey,
		Sequence:      sequence + 1,
	})
	if err != nil {
		if isStampConflict(err) {
			return GanharSeloResponse{}, apperror.New(http.StatusConflict, "SELO_JA_REGISTRADO", "Este QR code já foi utilizado ou o limite diário foi atingido")
		}
		return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
	}

	count, err := queries.CountStampsByProgram(ctx, sqlc.CountStampsByProgramParams{UserID: userID, ProgramID: programID})
	if err != nil {
		return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
	}

	completou := count >= validation.StampGoal
	remaining := int(validation.StampGoal - count)
	if remaining < 0 {
		remaining = 0
	}
	progress := int((float64(count) / float64(validation.StampGoal)) * 100)
	if progress > 100 {
		progress = 100
	}

	if completou {
		rewardID, err := id.New("rew_")
		if err != nil {
			return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
		}
		rewardCode, err := id.New("RESG")
		if err != nil {
			return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
		}

		_, err = queries.UpsertRewardRedemption(ctx, sqlc.UpsertRewardRedemptionParams{
			ID:          rewardID,
			UserID:      userID,
			ProgramID:   programID,
			RewardCode:  rewardCode,
			CompletedAt: created.AcquiredAt,
			ExpiresAt:   pgtype.Timestamptz{Time: pgutil.Time(created.AcquiredAt).Add(30 * 24 * time.Hour), Valid: true},
		})
		if err != nil {
			return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return GanharSeloResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível registrar o selo", err)
	}

	return GanharSeloResponse{
		Selo: Selo{
			ID:             created.ID,
			ProgramaID:     created.ProgramID,
			UsuarioID:      created.UserID,
			DataAquisicao:  pgutil.Time(created.AcquiredAt).Format(time.RFC3339),
			ChaveValidacao: created.ValidationKey,
			Sequencia:      int(created.Sequence),
			Validado:       created.Validated,
		},
		CompletouPrograma: completou,
		SelosRestantes:    remaining,
		Progresso:         progress,
	}, nil
}

func (s *StampService) lookupQRCode(ctx context.Context, qrCodeData string) (sqlc.GetQRCodeValidationRow, error) {
	qrHash := hashutil.SHA256(qrCodeData)
	validation, err := s.store.Queries.GetQRCodeValidation(ctx, qrHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return sqlc.GetQRCodeValidationRow{}, apperror.New(http.StatusBadRequest, "QR_INVALIDO", "QR code inválido ou expirado")
		}

		return sqlc.GetQRCodeValidationRow{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível validar o QR code", err)
	}

	if validation.Used || (!validation.ExpiresAt.Time.IsZero() && validation.ExpiresAt.Time.Before(time.Now().UTC())) {
		return sqlc.GetQRCodeValidationRow{}, apperror.New(http.StatusBadRequest, "QR_INVALIDO", "QR code inválido ou expirado")
	}

	return validation, nil
}

func isStampConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
