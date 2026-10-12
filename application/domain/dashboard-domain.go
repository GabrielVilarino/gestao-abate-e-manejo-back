package domain

import "time"

type FiltroDashboardAbates struct {
	ProprietarioID *int
	FazendaID      *int
	DataInicio     *time.Time
	DataFim        *time.Time
}

type DashboardResumoAbates struct {
	TotalAbates               int
	TotalCabecas              int
	MediaCabecasPorAbate      float64
	PesoMedioFazenda          float64
	PesoMedioFrigorifico      float64
	RendimentoMedioCarcaca    float64
	DistanciaMediaFrigorifico float64
}

type DashboardPorFazenda struct {
	FazendaID            int
	NomeFazenda          string
	TotalAbates          int
	TotalCabecas         int
	MediaCabecasPorAbate float64
}

type DashboardEvolucao struct {
	Data         time.Time
	TotalAbates  int
	TotalCabecas int
}

type DashboardPorCategoria struct {
	CategoriaAnimal string
	TotalAbates     int
	TotalCabecas    int
}

type DashboardAbates struct {
	Resumo             DashboardResumoAbates
	PorFazenda         []DashboardPorFazenda
	Evolucao           []DashboardEvolucao
	PorCategoriaAnimal []DashboardPorCategoria
}
