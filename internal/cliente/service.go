package cliente

import (
	"fmt"

	"github.com/mundoinvest/client-management/internal/pipefy"
)

type clientPersister interface {
	Create(c *Cliente) error
	UpdateCardID(email, cardID string) error
}

type pipefyClient interface {
	SimulateSend(payload map[string]interface{}) string
	BuildCreateCardPayload(input pipefy.CreateCardInput) map[string]interface{}
}

type Service struct {
	repo         clientPersister
	pipefyClient pipefyClient
}

func NewService(repo clientPersister, pc pipefyClient) *Service {
	return &Service{repo: repo, pipefyClient: pc}
}

func (s *Service) Criar(input CriarClienteInput) (*Cliente, error) {
	c := &Cliente{
		Nome:            input.Nome,
		Email:           input.Email,
		TipoSolicitacao: input.TipoSolicitacao,
		ValorPatrimonio: input.ValorPatrimonio,
		Status:          "Aguardando Análise",
	}

	if err := s.repo.Create(c); err != nil {
		return nil, fmt.Errorf("failed to persist client: %w", err)
	}

	pipefyPayload := s.buildCreateCardPayload(c)
	cardID := s.pipefyClient.SimulateSend(pipefyPayload)

	if err := s.repo.UpdateCardID(c.Email, cardID); err != nil {
		return nil, fmt.Errorf("failed to update card_id: %w", err)
	}
	c.CardID = cardID

	return c, nil
}

func (s *Service) buildCreateCardPayload(c *Cliente) map[string]interface{} {
	return s.pipefyClient.BuildCreateCardPayload(pipefy.CreateCardInput{
		PipeID: 123,
		Title:  c.Nome,
		FieldsAttributes: []pipefy.FieldAttribute{
			{FieldID: "nome", FieldValue: c.Nome},
			{FieldID: "email", FieldValue: c.Email},
			{FieldID: "tipo_solicitacao", FieldValue: c.TipoSolicitacao},
			{FieldID: "valor_patrimonio", FieldValue: fmt.Sprintf("%.2f", c.ValorPatrimonio)},
		},
	})
}
