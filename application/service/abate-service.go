package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
)

const (
	maxFotoSize      = 10 << 20
	storageTimeout   = 30 * time.Second
	cleanupBatchSize = 20
)

type AbateService struct {
	AbatePort       output.AbatePort
	StoragePort     output.StoragePort
	ReportGenerator output.AbateReportGenerator
	cleanupWake     chan struct{}
}

func NewAbateService(abatePort output.AbatePort, storagePort output.StoragePort, reportGenerator ...output.AbateReportGenerator) *AbateService {
	var generator output.AbateReportGenerator
	if len(reportGenerator) > 0 {
		generator = reportGenerator[0]
	}
	return &AbateService{
		AbatePort: abatePort, StoragePort: storagePort, ReportGenerator: generator,
		cleanupWake: make(chan struct{}, 1),
	}
}

func (a *AbateService) StartStorageCleanupWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		a.retryPendingCleanup(ctx)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.retryPendingCleanup(ctx)
			case <-a.cleanupWake:
				a.retryPendingCleanup(ctx)
			}
		}
	}()
}

func (a *AbateService) CreateAbate(abate *domain.Abate) error {
	return a.AbatePort.CreateAbate(abate)
}

func (a *AbateService) FindAbateByID(id int) (*domain.Abate, error) {
	return a.AbatePort.FindAbateByID(id)
}

func (a *AbateService) FindAbates(filtro domain.FiltroAbate) ([]domain.Abate, error) {
	if filtro.DataInicio != nil && filtro.DataFim != nil && filtro.DataInicio.After(*filtro.DataFim) {
		return nil, domain.ErrPeriodoInvalido
	}
	return a.AbatePort.FindAbates(filtro)
}

func (a *AbateService) GenerateAbateReport(ctx context.Context, abateIDs []int) ([]byte, error) {
	if len(abateIDs) == 0 {
		return nil, domain.ErrListaAbatesInvalida
	}
	for _, id := range abateIDs {
		if id <= 0 {
			return nil, domain.ErrListaAbatesInvalida
		}
	}
	uniqueIDs := uniquePositiveIDs(abateIDs)
	abates, err := a.AbatePort.FindAbatesByIDs(ctx, uniqueIDs)
	if err != nil {
		return nil, err
	}
	abatesByID := make(map[int]domain.Abate, len(abates))
	for _, abate := range abates {
		abatesByID[abate.ID] = abate
	}

	reports := make([]output.AbateReport, 0, len(uniqueIDs))
	for _, id := range uniqueIDs {
		abate, ok := abatesByID[id]
		if !ok {
			return nil, domain.ErrAbateNaoEncontrado
		}
		report, err := a.buildAbateReport(ctx, abate)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	if a.ReportGenerator == nil {
		return nil, errors.New("gerador de relatório de abate não configurado")
	}
	return a.ReportGenerator.Generate(ctx, reports)
}

func (a *AbateService) buildAbateReport(ctx context.Context, abate domain.Abate) (output.AbateReport, error) {
	report := output.AbateReport{Abate: abate}
	denticoes := make(map[int]int, len(domain.DenticoesAbate()))
	for _, denticao := range domain.DenticoesAbate() {
		denticoes[denticao] = 0
	}
	for _, item := range abate.EtapaFazenda.QuantidadeAnimal {
		if _, ok := denticoes[item.QtdDenticao]; ok {
			denticoes[item.QtdDenticao] += item.QtdAnimais
			report.QuantidadeAnimais += item.QtdAnimais
		}
	}
	if report.QuantidadeAnimais > 0 {
		quantidade := float64(report.QuantidadeAnimais)
		report.MediaKGFrigorifico = abate.EtapaFrigorifico.PesoTotal / quantidade
		report.MediaArrobaFrigorifico = report.MediaKGFrigorifico / 15
		report.MediaKGFazenda = abate.EtapaFazenda.PesoTotal / quantidade
		report.PesoMedioBalancao = abate.EtapaFrigorifico.Balancao / quantidade
		report.Esvaziamento = report.MediaKGFazenda - report.PesoMedioBalancao
		if report.MediaKGFazenda != 0 {
			report.RendimentoCarcaca = report.MediaKGFrigorifico / report.MediaKGFazenda
		}
		if report.PesoMedioBalancao != 0 {
			report.RendimentoBalancao = report.MediaKGFrigorifico / report.PesoMedioBalancao
		}
	}

	report.Denticoes = make([]output.AbateReportDenticao, 0, len(denticoes))
	for _, denticao := range domain.DenticoesAbate() {
		quantidade := denticoes[denticao]
		report.Denticoes = append(report.Denticoes, output.AbateReportDenticao{
			Denticao: denticao, Quantidade: quantidade,
			Percentual: percentage(quantidade, report.QuantidadeAnimais),
		})
	}

	acabamentos := make(map[string]int, len(domain.AcabamentosCarcacaAbate()))
	for _, acabamento := range domain.AcabamentosCarcacaAbate() {
		acabamentos[acabamento] = 0
	}
	for _, item := range abate.EtapaFrigorifico.AcabamentoCarcaca {
		if _, ok := acabamentos[item.Acabamento]; ok {
			acabamentos[item.Acabamento] += item.QtdAnimais
		}
	}
	report.Acabamentos = make([]output.AbateReportAcabamento, 0, len(acabamentos))
	for _, acabamento := range domain.AcabamentosCarcacaAbate() {
		quantidade := acabamentos[acabamento]
		report.Acabamentos = append(report.Acabamentos, output.AbateReportAcabamento{
			Classificacao: acabamento, Quantidade: quantidade,
			Percentual: percentage(quantidade, report.QuantidadeAnimais),
		})
	}

	classificacoes := make(map[string]int, len(domain.ClassificacoesFrigorificoAbate()))
	for _, classificacao := range domain.ClassificacoesFrigorificoAbate() {
		classificacoes[classificacao] = 0
	}
	for _, item := range abate.EtapaFrigorifico.ClassificacaoFrigorifico {
		if _, ok := classificacoes[item.Classificacao]; ok {
			classificacoes[item.Classificacao] += item.QtdAnimais
		}
	}
	report.Classificacoes = make([]output.AbateReportClassificacao, 0, len(classificacoes))
	for _, classificacao := range domain.ClassificacoesFrigorificoAbate() {
		quantidade := classificacoes[classificacao]
		report.Classificacoes = append(report.Classificacoes, output.AbateReportClassificacao{
			Classificacao: classificacao, Quantidade: quantidade,
			Percentual: percentage(quantidade, report.QuantidadeAnimais),
		})
	}

	type distribuicaoPesoTotal struct {
		quantidade int
		pesoTotal  float64
	}
	distribuicoes := make(map[string]distribuicaoPesoTotal, len(domain.FaixasDistribuicaoPesoAbate()))
	for _, faixa := range domain.FaixasDistribuicaoPesoAbate() {
		distribuicoes[faixa] = distribuicaoPesoTotal{}
	}
	for _, item := range abate.EtapaFrigorifico.DistribuicaoPeso {
		total, ok := distribuicoes[item.Classificacao]
		if !ok {
			continue
		}
		total.quantidade += item.QtdAnimais
		total.pesoTotal += item.PesoTotal
		distribuicoes[item.Classificacao] = total
	}
	report.DistribuicoesPeso = make([]output.AbateReportDistribuicaoPeso, 0, len(distribuicoes))
	for _, faixa := range domain.FaixasDistribuicaoPesoAbate() {
		total := distribuicoes[faixa]
		distribuicao := output.AbateReportDistribuicaoPeso{
			Classificacao: faixa, Quantidade: total.quantidade, PesoTotal: total.pesoTotal,
		}
		if total.quantidade > 0 {
			distribuicao.MediaKG = total.pesoTotal / float64(total.quantidade)
			distribuicao.MediaArroba = distribuicao.MediaKG / 15
		}
		if abate.EtapaFrigorifico.PesoTotal != 0 {
			distribuicao.Percentual = total.pesoTotal / abate.EtapaFrigorifico.PesoTotal * 100
		}
		report.DistribuicoesPeso = append(report.DistribuicoesPeso, distribuicao)
	}
	report.Observacao = "Não há observação sobre o abate."
	if abate.DadosGeraisAbate.Observacao != nil && strings.TrimSpace(*abate.DadosGeraisAbate.Observacao) != "" {
		report.Observacao = strings.TrimSpace(*abate.DadosGeraisAbate.Observacao)
	}

	var err error
	report.FotosFazenda, err = a.reportPhotos(ctx, abate.EtapaFazenda.Fotos)
	if err != nil {
		return output.AbateReport{}, err
	}
	report.FotosFrigorifico, err = a.reportPhotos(ctx, abate.EtapaFrigorifico.Fotos)
	if err != nil {
		return output.AbateReport{}, err
	}
	return report, nil
}

func (a *AbateService) reportPhotos(ctx context.Context, photos []domain.FotoAbate) ([]output.AbateReportPhoto, error) {
	result := make([]output.AbateReportPhoto, 0, len(photos))
	for _, photo := range photos {
		metadata, body, err := a.DownloadFotoAbate(ctx, photo.ID)
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(body)
		closeErr := body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		result = append(result, output.AbateReportPhoto{
			NomeOriginal: metadata.NomeOriginal,
			DataURL:      "data:" + metadata.ContentType + ";base64," + base64.StdEncoding.EncodeToString(data),
		})
	}
	return result, nil
}

func percentage(value, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(value)/float64(total)*1000) / 10
}

func uniquePositiveIDs(ids []int) []int {
	result := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func (a *AbateService) UpdateDadosGeraisAbate(abateID int, dados domain.DadosGeraisAbate) error {
	return a.AbatePort.UpdateDadosGeraisAbate(abateID, dados)
}

func (a *AbateService) UpdateEtapaFazenda(abateID int, etapa domain.EtapaFazenda) error {
	return a.AbatePort.UpdateEtapaFazenda(abateID, etapa)
}

func (a *AbateService) UpdateEtapaFrigorifico(abateID int, etapa domain.EtapaFrigorifico) error {
	return a.AbatePort.UpdateEtapaFrigorifico(abateID, etapa)
}

func (a *AbateService) Delete(ctx context.Context, abateID int) error {
	queued, err := a.AbatePort.Delete(abateID)
	if err != nil {
		return err
	}
	if len(queued) > 0 {
		a.scheduleCleanup()
	}
	return nil
}

func (a *AbateService) UploadFotoAbate(
	ctx context.Context,
	abateID int,
	etapa string,
	nomeOriginal string,
	data []byte,
) (*domain.FotoAbate, error) {
	etapa = strings.ToUpper(strings.TrimSpace(etapa))
	if etapa != domain.EtapaFazendaFoto && etapa != domain.EtapaFrigorificoFoto {
		return nil, domain.ErrEtapaFotoInvalida
	}
	if len(data) > maxFotoSize {
		return nil, domain.ErrTamanhoFotoInvalido
	}
	contentType := http.DetectContentType(data)
	if contentType != "image/png" && contentType != "image/jpeg" {
		return nil, domain.ErrFormatoFotoInvalido
	}

	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	abateRelease, contexto, err := a.acquireFotoContext(ctx, abateID)
	if err != nil {
		return nil, err
	}
	defer abateRelease()
	objectKey := path.Join(
		fmt.Sprintf("%d", contexto.ProprietarioID),
		fmt.Sprintf("%d", contexto.NumeroLote),
		etapa,
		hash,
	)
	release, err := a.acquireObjectLocks(ctx, []string{objectKey})
	if err != nil {
		return nil, err
	}
	defer release()

	existente, err := a.AbatePort.FindFotoByHash(abateID, etapa, hash)
	if err != nil {
		return nil, err
	}
	if existente != nil {
		return existente, nil
	}
	if err := a.AbatePort.EnqueueObjectCleanup(objectKey); err != nil {
		return nil, err
	}

	storageCtx, cancel := storageContext(ctx)
	err = a.StoragePort.Upload(storageCtx, objectKey, data, contentType)
	cancel()
	if err != nil {
		return nil, err
	}
	foto := &domain.FotoAbate{
		AbateID: abateID, Etapa: etapa, ObjectKey: objectKey,
		NomeOriginal: nomeOriginal, ContentType: contentType,
		Tamanho: int64(len(data)), SHA256: hash,
	}
	if err := a.AbatePort.CreateFoto(foto); err != nil {
		cleanupErr := a.processObjectCleanupLocked(ctx, objectKey)
		if cleanupErr != nil {
			_ = a.AbatePort.MarkObjectCleanupFailed(objectKey, cleanupErr.Error())
			a.scheduleCleanup()
		}
		return nil, errors.Join(err, cleanupErr)
	}
	if err := a.AbatePort.DeletePendingObjectCleanup(objectKey); err != nil {
		a.scheduleCleanup()
	}
	return foto, nil
}

func (a *AbateService) DownloadFotoAbate(ctx context.Context, fotoID int) (*domain.FotoAbate, io.ReadCloser, error) {
	foto, err := a.AbatePort.FindFotoByID(fotoID)
	if err != nil {
		return nil, nil, err
	}
	if foto == nil {
		return nil, nil, domain.ErrFotoNaoEncontrada
	}
	release, err := a.acquireObjectLocks(ctx, []string{foto.ObjectKey})
	if err != nil {
		return nil, nil, err
	}
	foto, err = a.AbatePort.FindFotoByID(fotoID)
	if err != nil {
		release()
		return nil, nil, err
	}
	if foto == nil {
		release()
		return nil, nil, domain.ErrFotoNaoEncontrada
	}
	storageCtx, cancel := storageContext(ctx)
	body, err := a.StoragePort.Download(storageCtx, foto.ObjectKey)
	if err != nil {
		cancel()
		release()
		return nil, nil, err
	}
	return foto, &lockedReadCloser{ReadCloser: body, release: func() { cancel(); release() }}, nil
}

type lockedReadCloser struct {
	io.ReadCloser
	release func()
	once    sync.Once
}

func (r *lockedReadCloser) Close() error {
	err := r.ReadCloser.Close()
	r.once.Do(r.release)
	return err
}

func (a *AbateService) DeleteFotoAbate(ctx context.Context, fotoID int) error {
	foto, err := a.AbatePort.FindFotoByID(fotoID)
	if err != nil {
		return err
	}
	if foto == nil {
		return domain.ErrFotoNaoEncontrada
	}
	foto, err = a.AbatePort.DeleteFoto(fotoID)
	if err != nil {
		return err
	}
	if foto == nil {
		return domain.ErrFotoNaoEncontrada
	}
	a.scheduleCleanup()
	return nil
}

func (a *AbateService) retryPendingCleanup(ctx context.Context) {
	keys, err := a.AbatePort.FindPendingObjectCleanup(cleanupBatchSize)
	if err != nil {
		return
	}
	for _, key := range keys {
		release, err := a.acquireObjectLocks(ctx, []string{key})
		if err != nil {
			_ = a.AbatePort.MarkObjectCleanupFailed(key, err.Error())
			continue
		}
		err = a.processObjectCleanupLocked(ctx, key)
		release()
		if err != nil {
			_ = a.AbatePort.MarkObjectCleanupFailed(key, err.Error())
			continue
		}
	}
}

func (a *AbateService) scheduleCleanup() {
	select {
	case a.cleanupWake <- struct{}{}:
	default:
	}
}

func (a *AbateService) acquireFotoContext(ctx context.Context, abateID int) (func(), *domain.ContextoFotoAbate, error) {
	release, err := a.acquireObjectLocks(ctx, []string{fmt.Sprintf("identity:abate:%d", abateID)})
	if err != nil {
		return nil, nil, err
	}
	contexto, err := a.AbatePort.GetContextoFoto(abateID)
	if err != nil || contexto == nil {
		release()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, domain.ErrAbateNaoEncontrado
	}
	return release, contexto, nil
}

func (a *AbateService) processObjectCleanupLocked(ctx context.Context, objectKey string) error {
	count, err := a.AbatePort.CountFotosByObjectKey(objectKey)
	if err != nil {
		return err
	}
	if count == 0 {
		storageCtx, cancel := storageContext(ctx)
		err = a.StoragePort.Delete(storageCtx, objectKey)
		cancel()
		if err != nil {
			return err
		}
	}
	return a.AbatePort.DeletePendingObjectCleanup(objectKey)
}

func (a *AbateService) acquireObjectLocks(ctx context.Context, keys []string) (func(), error) {
	keys = uniqueSorted(keys)
	releases := make([]func() error, 0, len(keys))
	for _, key := range keys {
		release, err := a.AbatePort.AcquireObjectLock(ctx, key)
		if err != nil {
			for i := len(releases) - 1; i >= 0; i-- {
				_ = releases[i]()
			}
			return nil, err
		}
		releases = append(releases, release)
	}
	return func() {
		for i := len(releases) - 1; i >= 0; i-- {
			_ = releases[i]()
		}
	}, nil
}

func uniqueSorted(keys []string) []string {
	unique := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if key != "" {
			unique[key] = struct{}{}
		}
	}
	result := make([]string, 0, len(unique))
	for key := range unique {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func storageContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, storageTimeout)
}
