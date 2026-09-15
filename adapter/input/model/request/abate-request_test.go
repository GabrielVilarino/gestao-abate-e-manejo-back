package request

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestAbateRequestCatalogValidation(t *testing.T) {
	validate := validator.New()
	validate.SetTagName("binding")
	zero, one, eight := 0, 1, 8

	tests := []struct {
		name  string
		value any
		valid bool
	}{
		{name: "dentição zero válida", value: QtdDenticaoRequest{QtdDenticao: &zero}, valid: true},
		{name: "dentição oito válida", value: QtdDenticaoRequest{QtdDenticao: &eight}, valid: true},
		{name: "dentição ausente", value: QtdDenticaoRequest{}, valid: false},
		{name: "dentição inválida", value: QtdDenticaoRequest{QtdDenticao: &one}, valid: false},
		{name: "acabamento válido", value: AcabamentoCarcacaRequest{Acabamento: "MEDIANO UP"}, valid: true},
		{name: "acabamento inválido", value: AcabamentoCarcacaRequest{Acabamento: "MEDIANO up"}, valid: false},
		{name: "classificação válida", value: ClassificacaoFrigorificoRequest{Classificacao: "BOI MÉDIO / NORMAL"}, valid: true},
		{name: "classificação inválida", value: ClassificacaoFrigorificoRequest{Classificacao: "NORMAL"}, valid: false},
		{name: "distribuição válida", value: DistribuicaoPesoRequest{Classificacao: "acima de 24"}, valid: true},
		{name: "distribuição inválida", value: DistribuicaoPesoRequest{Classificacao: "24 ou mais"}, valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validate.Struct(test.value)
			if test.valid && err != nil {
				t.Fatalf("valor válido rejeitado: %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("valor inválido aceito")
			}
		})
	}
}

func TestAbateStageRequestsValidateNestedCatalogItems(t *testing.T) {
	validate := validator.New()
	validate.SetTagName("binding")
	invalidDenticao := 3

	fazenda := EtapaFazendaRequest{QuantidadeAnimal: []QtdDenticaoRequest{{QtdDenticao: &invalidDenticao}}}
	if err := validate.Struct(fazenda); err == nil {
		t.Fatal("etapa fazenda aceitou dentição inválida")
	}

	frigorifico := EtapaFrigorificoRequest{
		AcabamentoCarcaca:        []AcabamentoCarcacaRequest{{Acabamento: "INVÁLIDO"}},
		ClassificacaoFrigorifico: []ClassificacaoFrigorificoRequest{{Classificacao: "BOI FRACO"}},
		DistribuicaoPeso:         []DistribuicaoPesoRequest{{Classificacao: "18 a 19.9"}},
	}
	if err := validate.Struct(frigorifico); err == nil {
		t.Fatal("etapa frigorífico aceitou acabamento inválido")
	}
}
