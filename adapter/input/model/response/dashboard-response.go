package response

type DashboardResumoAbatesResponse struct {
	TotalAbates               int     `json:"total_abates"`
	TotalCabecas              int     `json:"total_cabecas"`
	MediaCabecasPorAbate      float64 `json:"media_cabecas_por_abate"`
	PesoMedioFazenda          float64 `json:"peso_medio_fazenda"`
	PesoMedioFrigorifico      float64 `json:"peso_medio_frigorifico"`
	RendimentoMedioCarcaca    float64 `json:"rendimento_medio_carcaca"`
	DistanciaMediaFrigorifico float64 `json:"distancia_media_frigorifico"`
}

type DashboardPorFazendaResponse struct {
	FazendaID            int     `json:"fazenda_id"`
	NomeFazenda          string  `json:"nome_fazenda"`
	TotalAbates          int     `json:"total_abates"`
	TotalCabecas         int     `json:"total_cabecas"`
	MediaCabecasPorAbate float64 `json:"media_cabecas_por_abate"`
}

type DashboardEvolucaoResponse struct {
	Data         string `json:"data"`
	TotalAbates  int    `json:"total_abates"`
	TotalCabecas int    `json:"total_cabecas"`
}

type DashboardPorCategoriaResponse struct {
	CategoriaAnimal string `json:"categoria_animal"`
	TotalAbates     int    `json:"total_abates"`
	TotalCabecas    int    `json:"total_cabecas"`
}

type DashboardAbatesResponse struct {
	Resumo             DashboardResumoAbatesResponse   `json:"resumo"`
	PorFazenda         []DashboardPorFazendaResponse   `json:"por_fazenda"`
	Evolucao           []DashboardEvolucaoResponse     `json:"evolucao"`
	PorCategoriaAnimal []DashboardPorCategoriaResponse `json:"por_categoria_animal"`
}
