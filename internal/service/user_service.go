package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/felip/api-fidelidade/internal/apperror"
	"github.com/felip/api-fidelidade/internal/auth"
	"github.com/felip/api-fidelidade/internal/id"
	"github.com/felip/api-fidelidade/internal/storage"
	"github.com/felip/api-fidelidade/internal/store/sqlc"
	"github.com/felip/api-fidelidade/internal/support/pgutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var passwordPolicy = regexp.MustCompile(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$`)

type UserService struct {
	queries      *sqlc.Queries
	tokenManager *auth.TokenManager
	storage      *storage.Local
}

type CriarUsuarioInput struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type AtualizarUsuarioInput struct {
	Nome         *string `json:"nome"`
	ImagemPerfil *string `json:"imagemPerfil"`
}

type AutenticacaoInput struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type Usuario struct {
	ID              string  `json:"id"`
	Nome            string  `json:"nome"`
	Email           string  `json:"email"`
	ImagemPerfil    string  `json:"imagemPerfil,omitempty"`
	Ativo           bool    `json:"ativo"`
	DataCriacao     string  `json:"dataCriacao"`
	DataAtualizacao *string `json:"dataAtualizacao,omitempty"`
}

type AutenticacaoResponse struct {
	Token   string  `json:"token"`
	Usuario Usuario `json:"usuario"`
}

func NewUserService(queries *sqlc.Queries, tokenManager *auth.TokenManager, fileStorage *storage.Local) *UserService {
	return &UserService{queries: queries, tokenManager: tokenManager, storage: fileStorage}
}

func (s *UserService) Criar(ctx context.Context, input CriarUsuarioInput) (Usuario, error) {
	if err := validateCreateUser(input); err != nil {
		return Usuario{}, err
	}

	passwordHash, err := auth.HashPassword(input.Senha)
	if err != nil {
		return Usuario{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível criar o usuário", err)
	}

	userID, err := id.New("usr_")
	if err != nil {
		return Usuario{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível criar o usuário", err)
	}

	row, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:              userID,
		Name:            strings.TrimSpace(input.Nome),
		Email:           strings.TrimSpace(strings.ToLower(input.Email)),
		PasswordHash:    passwordHash,
		ProfileImageUrl: pgutil.Text(""),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Usuario{}, apperror.New(http.StatusConflict, "EMAIL_DUPLICADO", "Este email já está em uso")
		}

		return Usuario{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível criar o usuário", err)
	}

	return mapCreatedUser(row), nil
}

func (s *UserService) Autenticar(ctx context.Context, input AutenticacaoInput) (AutenticacaoResponse, error) {
	user, err := s.queries.GetUserAuthByEmail(ctx, strings.TrimSpace(strings.ToLower(input.Email)))
	if err != nil {
		if err == pgx.ErrNoRows {
			return AutenticacaoResponse{}, apperror.New(http.StatusUnauthorized, "NAO_AUTORIZADO", "Credenciais inválidas")
		}

		return AutenticacaoResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível autenticar o usuário", err)
	}

	if err := auth.CheckPassword(user.PasswordHash, input.Senha); err != nil {
		return AutenticacaoResponse{}, apperror.New(http.StatusUnauthorized, "NAO_AUTORIZADO", "Credenciais inválidas")
	}

	token, err := s.tokenManager.Issue(user.ID)
	if err != nil {
		return AutenticacaoResponse{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível autenticar o usuário", err)
	}

	return AutenticacaoResponse{Token: token, Usuario: mapUser(user)}, nil
}

func (s *UserService) Obter(ctx context.Context, userID string) (Usuario, error) {
	row, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Usuario{}, apperror.New(http.StatusNotFound, "NAO_ENCONTRADO", "Recurso não encontrado")
		}

		return Usuario{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível obter o usuário", err)
	}

	return mapGetUser(row), nil
}

func (s *UserService) Atualizar(ctx context.Context, userID string, input AtualizarUsuarioInput) (Usuario, error) {
	if input.Nome == nil && input.ImagemPerfil == nil {
		return Usuario{}, apperror.WithDetails(
			apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Dados inválidos"),
			apperror.Detail{Campo: "body", Mensagem: "Informe ao menos um campo para atualização"},
		)
	}

	name := ""
	if input.Nome != nil {
		trimmed := strings.TrimSpace(*input.Nome)
		if len(trimmed) < 3 || len(trimmed) > 100 {
			return Usuario{}, apperror.WithDetails(
				apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Dados inválidos"),
				apperror.Detail{Campo: "nome", Mensagem: "Nome deve conter entre 3 e 100 caracteres"},
			)
		}
		name = trimmed
	}

	imageURL := ""
	if input.ImagemPerfil != nil {
		imageURL = strings.TrimSpace(*input.ImagemPerfil)
	}

	row, err := s.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		UpdateName:      input.Nome != nil,
		Name:            name,
		UpdateImage:     input.ImagemPerfil != nil,
		ProfileImageUrl: imageURL,
		ID:              userID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return Usuario{}, apperror.New(http.StatusNotFound, "NAO_ENCONTRADO", "Recurso não encontrado")
		}

		return Usuario{}, apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível atualizar o usuário", err)
	}

	return mapUpdatedUser(row), nil
}

func (s *UserService) Excluir(ctx context.Context, userID string) error {
	if _, err := s.Obter(ctx, userID); err != nil {
		return err
	}

	if err := s.queries.DeleteUser(ctx, userID); err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível excluir o usuário", err)
	}

	return nil
}

func (s *UserService) UploadImagem(ctx context.Context, userID string, contentType string, data []byte) (string, error) {
	if _, err := s.Obter(ctx, userID); err != nil {
		return "", err
	}

	imageURL, err := s.storage.SaveProfileImage(userID, contentType, data)
	if err != nil {
		return "", apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", fmt.Sprintf("Formato de imagem inválido: %s", contentType))
	}

	_, err = s.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		UpdateName:      false,
		Name:            "",
		UpdateImage:     true,
		ProfileImageUrl: imageURL,
		ID:              userID,
	})
	if err != nil {
		return "", apperror.Wrap(http.StatusInternalServerError, "ERRO_INTERNO", "Não foi possível atualizar a imagem do usuário", err)
	}

	return imageURL, nil
}

func validateCreateUser(input CriarUsuarioInput) error {
	var detalhes []apperror.Detail

	nome := strings.TrimSpace(input.Nome)
	if len(nome) < 3 || len(nome) > 100 {
		detalhes = append(detalhes, apperror.Detail{Campo: "nome", Mensagem: "Nome deve conter entre 3 e 100 caracteres"})
	}

	email := strings.TrimSpace(input.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		detalhes = append(detalhes, apperror.Detail{Campo: "email", Mensagem: "Formato de email inválido"})
	}

	if !passwordPolicy.MatchString(input.Senha) {
		detalhes = append(detalhes, apperror.Detail{Campo: "senha", Mensagem: "Senha fora da política mínima"})
	}

	if len(detalhes) > 0 {
		return apperror.WithDetails(apperror.New(http.StatusBadRequest, "VALIDACAO_FALHA", "Dados inválidos"), detalhes...)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func mapCreatedUser(row sqlc.CreateUserRow) Usuario {
	user := Usuario{
		ID:           row.ID,
		Nome:         row.Name,
		Email:        row.Email,
		ImagemPerfil: pgutil.TextOrEmpty(row.ProfileImageUrl),
		Ativo:        row.Active,
		DataCriacao:  pgutil.Time(row.CreatedAt).Format(time.RFC3339),
	}
	if row.UpdatedAt.Valid {
		updated := pgutil.Time(row.UpdatedAt).Format(time.RFC3339)
		user.DataAtualizacao = &updated
	}
	return user
}

func mapUpdatedUser(row sqlc.UpdateUserRow) Usuario {
	user := Usuario{
		ID:           row.ID,
		Nome:         row.Name,
		Email:        row.Email,
		ImagemPerfil: pgutil.TextOrEmpty(row.ProfileImageUrl),
		Ativo:        row.Active,
		DataCriacao:  pgutil.Time(row.CreatedAt).Format(time.RFC3339),
	}
	if row.UpdatedAt.Valid {
		updated := pgutil.Time(row.UpdatedAt).Format(time.RFC3339)
		user.DataAtualizacao = &updated
	}
	return user
}

func mapGetUser(row sqlc.GetUserByIDRow) Usuario {
	user := Usuario{
		ID:           row.ID,
		Nome:         row.Name,
		Email:        row.Email,
		ImagemPerfil: pgutil.TextOrEmpty(row.ProfileImageUrl),
		Ativo:        row.Active,
		DataCriacao:  pgutil.Time(row.CreatedAt).Format(time.RFC3339),
	}
	if row.UpdatedAt.Valid {
		updated := pgutil.Time(row.UpdatedAt).Format(time.RFC3339)
		user.DataAtualizacao = &updated
	}
	return user
}

func mapUser(row sqlc.User) Usuario {
	user := Usuario{
		ID:           row.ID,
		Nome:         row.Name,
		Email:        row.Email,
		ImagemPerfil: pgutil.TextOrEmpty(row.ProfileImageUrl),
		Ativo:        row.Active,
		DataCriacao:  pgutil.Time(row.CreatedAt).Format(time.RFC3339),
	}
	if row.UpdatedAt.Valid {
		updated := pgutil.Time(row.UpdatedAt).Format(time.RFC3339)
		user.DataAtualizacao = &updated
	}
	return user
}
