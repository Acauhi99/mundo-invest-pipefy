package cliente

import (
	"fmt"

	"github.com/mundoinvest/client-management/internal/pipefy"
)

type Service struct {
	repo         *Repository
	pipefyClient *pipefy.Client
}

func NewService(repo *Repository, pipefyClient *pipefy.Client) *Service {
	return &Service{repo: repo, pipefyClient: pipefyClient}
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
		return nil, fmt.Errorf("erro ao persistir cliente: %w", err)
	}

	pipefyPayload := s.buildCreateCardPayload(c)
	s.pipefyClient.SimulateSend(pipefyPayload)

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
