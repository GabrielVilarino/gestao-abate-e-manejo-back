package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
)

type abatePortStub struct {
	findByID       func(int) (*domain.Abate, error)
	find           func(domain.FiltroAbate) ([]domain.Abate, error)
	updateDados    func(int, domain.DadosGeraisAbate) error
	deleteAbate    func(int) ([]string, error)
	contextoFoto   func(int) (*domain.ContextoFotoAbate, error)
	findFotoHash   func(int, string, string) (*domain.FotoAbate, error)
	createFoto     func(*domain.FotoAbate) error
	findFotoID     func(int) (*domain.FotoAbate, error)
	deleteFoto     func(int) (*domain.FotoAbate, error)
	countFotoKey   func(string) (int, error)
	enqueueCleanup func(string) error
	pendingCleanup func(int) ([]string, error)
	deleteCleanup  func(string) error
	markFailed     func(string, string) error
	acquireLock    func(context.Context, string) (func() error, error)
	findByIDs      func(context.Context, []int) ([]domain.Abate, error)
}

func (s abatePortStub) CreateAbate(*domain.Abate) error { return nil }
func (s abatePortStub) FindAbateByID(id int) (*domain.Abate, error) {
	if s.findByID == nil {
		return nil, nil
	}
	return s.findByID(id)
}
func (s abatePortStub) FindAbates(f domain.FiltroAbate) ([]domain.Abate, error) {
	return s.find(f)
}
func (s abatePortStub) FindAbatesByIDs(ctx context.Context, ids []int) ([]domain.Abate, error) {
	return s.findByIDs(ctx, ids)
}
func (s abatePortStub) UpdateDadosGeraisAbate(id int, v domain.DadosGeraisAbate) error {
	return s.updateDados(id, v)
}
func (s abatePortStub) UpdateEtapaFazenda(int, domain.EtapaFazenda) error         { return nil }
func (s abatePortStub) UpdateEtapaFrigorifico(int, domain.EtapaFrigorifico) error { return nil }
func (s abatePortStub) Delete(id int) ([]string, error)                           { return s.deleteAbate(id) }
func (s abatePortStub) GetContextoFoto(id int) (*domain.ContextoFotoAbate, error) {
	return s.contextoFoto(id)
}
func (s abatePortStub) FindFotoByHash(id int, etapa, hash string) (*domain.FotoAbate, error) {
	return s.findFotoHash(id, etapa, hash)
}
func (s abatePortStub) CreateFoto(v *domain.FotoAbate) error { return s.createFoto(v) }
func (s abatePortStub) FindFotoByID(id int) (*domain.FotoAbate, error) {
	return s.findFotoID(id)
}
func (s abatePortStub) DeleteFoto(id int) (*domain.FotoAbate, error) { return s.deleteFoto(id) }
func (s abatePortStub) CountFotosByObjectKey(key string) (int, error) {
	return s.countFotoKey(key)
}
func (s abatePortStub) EnqueueObjectCleanup(key string) error {
	return s.enqueueCleanup(key)
}
func (s abatePortStub) FindPendingObjectCleanup(limit int) ([]string, error) {
	if s.pendingCleanup == nil {
		return nil, nil
	}
	return s.pendingCleanup(limit)
}
func (s abatePortStub) DeletePendingObjectCleanup(key string) error {
	return s.deleteCleanup(key)
}
func (s abatePortStub) MarkObjectCleanupFailed(key, message string) error {
	if s.markFailed == nil {
		return nil
	}
	return s.markFailed(key, message)
}
func (s abatePortStub) AcquireObjectLock(ctx context.Context, key string) (func() error, error) {
	if s.acquireLock == nil {
		return func() error { return nil }, nil
	}
	return s.acquireLock(ctx, key)
}

type storagePortStub struct {
	upload   func(context.Context, string, []byte, string) error
	download func(context.Context, string) (io.ReadCloser, error)
	delete   func(context.Context, string) error
}

type abateReportGeneratorStub struct {
	generate func(context.Context, []output.AbateReport) ([]byte, error)
}

func (s abateReportGeneratorStub) Generate(ctx context.Context, reports []output.AbateReport) ([]byte, error) {
	return s.generate(ctx, reports)
}

func (s storagePortStub) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	return s.upload(ctx, key, data, contentType)
}
func (s storagePortStub) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.download(ctx, key)
}
func (s storagePortStub) Delete(ctx context.Context, key string) error {
	return s.delete(ctx, key)
}

func TestAbateServiceRejectsInvertedPeriod(t *testing.T) {
	start := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, -1)
	service := NewAbateService(abatePortStub{find: func(domain.FiltroAbate) ([]domain.Abate, error) {
		t.Fatal("repositório não deveria ser chamado")
		return nil, nil
	}}, storagePortStub{})
	_, err := service.FindAbates(domain.FiltroAbate{DataInicio: &start, DataFim: &end})
	if !errors.Is(err, domain.ErrPeriodoInvalido) {
		t.Fatalf("erro=%v", err)
	}
}

func TestAbateServiceUploadUsesDurableIntentAndHashKey(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 1}
	const hash = "289c77a179831309c3d4505952b0e5ff6bb91c0b4c6f68a9da8b1ed6d38c5e64"
	key := "7/42/FAZENDA/" + hash
	steps := make([]string, 0)
	port := abatePortStub{
		contextoFoto: func(int) (*domain.ContextoFotoAbate, error) {
			return &domain.ContextoFotoAbate{ProprietarioID: 7, NumeroLote: 42}, nil
		},
		findFotoHash: func(int, string, string) (*domain.FotoAbate, error) { return nil, nil },
		enqueueCleanup: func(got string) error {
			steps = append(steps, "enqueue")
			if got != key {
				t.Fatal(got)
			}
			return nil
		},
		createFoto: func(foto *domain.FotoAbate) error {
			steps = append(steps, "create")
			if foto.ObjectKey != key || foto.ContentType != "image/png" || foto.Tamanho != int64(len(data)) {
				t.Fatalf("foto=%+v", foto)
			}
			foto.ID = 11
			return nil
		},
		deleteCleanup: func(got string) error { steps = append(steps, "done"); return nil },
	}
	storage := storagePortStub{upload: func(ctx context.Context, got string, content []byte, contentType string) error {
		steps = append(steps, "upload")
		if got != key || !bytes.Equal(content, data) || contentType != "image/png" {
			t.Fatal("upload inválido")
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("upload sem timeout")
		}
		return nil
	}}
	foto, err := NewAbateService(port, storage).UploadFotoAbate(context.Background(), 7, "fazenda", "boi.png", data)
	if err != nil || foto.ID != 11 {
		t.Fatalf("foto=%+v err=%v", foto, err)
	}
	if got := len(steps); got != 4 || steps[0] != "enqueue" || steps[1] != "upload" || steps[2] != "create" || steps[3] != "done" {
		t.Fatalf("ordem=%v", steps)
	}
}

func TestAbateServiceCompensatesCreateFotoFailure(t *testing.T) {
	data := []byte{0xff, 0xd8, 0xff, 0xe0, 0, 0, 0, 0}
	createErr := errors.New("falha ao inserir")
	deleted := false
	port := abatePortStub{
		contextoFoto: func(int) (*domain.ContextoFotoAbate, error) {
			return &domain.ContextoFotoAbate{ProprietarioID: 1, NumeroLote: 1}, nil
		},
		findFotoHash:   func(int, string, string) (*domain.FotoAbate, error) { return nil, nil },
		enqueueCleanup: func(string) error { return nil },
		createFoto:     func(*domain.FotoAbate) error { return createErr },
		countFotoKey:   func(string) (int, error) { return 0, nil },
		deleteCleanup:  func(string) error { return nil },
	}
	storage := storagePortStub{
		upload: func(context.Context, string, []byte, string) error { return nil },
		delete: func(context.Context, string) error { deleted = true; return nil },
	}
	_, err := NewAbateService(port, storage).UploadFotoAbate(context.Background(), 1, "FAZENDA", "x.jpg", data)
	if !errors.Is(err, createErr) || !deleted {
		t.Fatalf("err=%v deleted=%v", err, deleted)
	}
}

func TestAbateServiceDoesNotDeleteSharedObject(t *testing.T) {
	deleted := false
	port := abatePortStub{
		findFotoID:    func(int) (*domain.FotoAbate, error) { return &domain.FotoAbate{ObjectKey: "shared"}, nil },
		deleteFoto:    func(int) (*domain.FotoAbate, error) { return &domain.FotoAbate{ObjectKey: "shared"}, nil },
		countFotoKey:  func(string) (int, error) { return 1, nil },
		deleteCleanup: func(string) error { return nil },
	}
	storage := storagePortStub{delete: func(context.Context, string) error { deleted = true; return nil }}
	if err := NewAbateService(port, storage).DeleteFotoAbate(context.Background(), 4); err != nil {
		t.Fatal(err)
	}
	if deleted {
		t.Fatal("objeto compartilhado foi apagado")
	}
}

func TestAbateServiceRetainsOutboxWhenR2DeleteFails(t *testing.T) {
	cleanupRemoved := false
	port := abatePortStub{
		findFotoID: func(int) (*domain.FotoAbate, error) {
			return &domain.FotoAbate{ObjectKey: "pending"}, nil
		},
		deleteFoto: func(int) (*domain.FotoAbate, error) {
			return &domain.FotoAbate{ObjectKey: "pending"}, nil
		},
		deleteCleanup: func(string) error { cleanupRemoved = true; return nil },
	}
	err := NewAbateService(port, storagePortStub{}).DeleteFotoAbate(context.Background(), 4)
	if err != nil {
		t.Fatalf("exclusão confirmada deve retornar sucesso: %v", err)
	}
	if cleanupRemoved {
		t.Fatal("outbox removida apesar da falha no R2")
	}
}

func TestAbateServiceCleanupContinuesAfterFailure(t *testing.T) {
	attempted := make([]string, 0, 2)
	marked := make([]string, 0, 1)
	completed := make([]string, 0, 1)
	port := abatePortStub{
		pendingCleanup: func(int) ([]string, error) { return []string{"old", "next"}, nil },
		countFotoKey:   func(string) (int, error) { return 0, nil },
		markFailed: func(key, message string) error {
			marked = append(marked, key)
			if message == "" {
				t.Fatal("erro não registrado")
			}
			return nil
		},
		deleteCleanup: func(key string) error { completed = append(completed, key); return nil },
	}
	storage := storagePortStub{delete: func(_ context.Context, key string) error {
		attempted = append(attempted, key)
		if key == "old" {
			return errors.New("R2 indisponível")
		}
		return nil
	}}
	NewAbateService(port, storage).retryPendingCleanup(context.Background())
	if len(attempted) != 2 || attempted[0] != "old" || attempted[1] != "next" {
		t.Fatalf("tentativas=%v", attempted)
	}
	if len(marked) != 1 || marked[0] != "old" {
		t.Fatalf("falhas registradas=%v", marked)
	}
	if len(completed) != 1 || completed[0] != "next" {
		t.Fatalf("tarefas concluídas=%v", completed)
	}
}

func TestAbateMutationsDoNotRunPendingCleanupSynchronously(t *testing.T) {
	pendingMustNotRun := func(int) ([]string, error) {
		t.Fatal("mutação HTTP não deve varrer a fila pendente")
		return nil, nil
	}
	t.Run("delete abate", func(t *testing.T) {
		service := NewAbateService(abatePortStub{
			pendingCleanup: pendingMustNotRun,
			deleteAbate:    func(int) ([]string, error) { return []string{"key"}, nil },
		}, storagePortStub{})
		if err := service.Delete(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("delete foto", func(t *testing.T) {
		service := NewAbateService(abatePortStub{
			pendingCleanup: pendingMustNotRun,
			findFotoID:     func(int) (*domain.FotoAbate, error) { return &domain.FotoAbate{ObjectKey: "key"}, nil },
			deleteFoto:     func(int) (*domain.FotoAbate, error) { return &domain.FotoAbate{ObjectKey: "key"}, nil },
		}, storagePortStub{})
		if err := service.DeleteFotoAbate(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("upload", func(t *testing.T) {
		service := NewAbateService(abatePortStub{
			pendingCleanup: pendingMustNotRun,
			contextoFoto: func(int) (*domain.ContextoFotoAbate, error) {
				return &domain.ContextoFotoAbate{ProprietarioID: 2, NumeroLote: 4}, nil
			},
			findFotoHash:   func(int, string, string) (*domain.FotoAbate, error) { return nil, nil },
			enqueueCleanup: func(string) error { return nil },
			createFoto:     func(*domain.FotoAbate) error { return nil },
			deleteCleanup:  func(string) error { return nil },
		}, storagePortStub{upload: func(context.Context, string, []byte, string) error { return nil }})
		png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
		if _, err := service.UploadFotoAbate(context.Background(), 1, "FAZENDA", "foto.png", png); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAbateServiceBlocksPathChangeWhenPhotosExist(t *testing.T) {
	port := abatePortStub{
		updateDados: func(int, domain.DadosGeraisAbate) error { return domain.ErrDadosGeraisComFotos },
	}
	err := NewAbateService(port, storagePortStub{}).UpdateDadosGeraisAbate(1, domain.DadosGeraisAbate{FazendaID: 1, NumeroLote: 3})
	if !errors.Is(err, domain.ErrDadosGeraisComFotos) {
		t.Fatalf("err=%v", err)
	}
}

func TestAbateServiceDownloadPropagatesContextWithTimeout(t *testing.T) {
	want := []byte("imagem")
	service := NewAbateService(abatePortStub{
		findFotoID: func(id int) (*domain.FotoAbate, error) {
			return &domain.FotoAbate{ID: id, ObjectKey: "private/key"}, nil
		},
	}, storagePortStub{download: func(ctx context.Context, key string) (io.ReadCloser, error) {
		if key != "private/key" {
			t.Fatal(key)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("download sem timeout")
		}
		return io.NopCloser(bytes.NewReader(want)), nil
	}})
	_, reader, err := service.DownloadFotoAbate(context.Background(), 8)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	got, err := io.ReadAll(reader)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("data=%q err=%v", got, err)
	}
}

func TestAbateServiceSerializesUploadAndPathUpdateByIdentity(t *testing.T) {
	var guard sync.Mutex
	locks := map[string]*sync.Mutex{}
	lockFor := func(key string) *sync.Mutex {
		guard.Lock()
		defer guard.Unlock()
		if locks[key] == nil {
			locks[key] = &sync.Mutex{}
		}
		return locks[key]
	}
	uploadEntered := make(chan struct{})
	finishUpload := make(chan struct{})
	updateAttempted := make(chan struct{})
	updateDone := make(chan struct{})
	port := abatePortStub{
		contextoFoto: func(int) (*domain.ContextoFotoAbate, error) {
			return &domain.ContextoFotoAbate{ProprietarioID: 2, NumeroLote: 4}, nil
		},
		findFotoHash:   func(int, string, string) (*domain.FotoAbate, error) { return nil, nil },
		enqueueCleanup: func(string) error { return nil },
		createFoto:     func(*domain.FotoAbate) error { return nil },
		deleteCleanup:  func(string) error { return nil },
		acquireLock: func(_ context.Context, key string) (func() error, error) {
			lock := lockFor(key)
			lock.Lock()
			return func() error { lock.Unlock(); return nil }, nil
		},
		updateDados: func(int, domain.DadosGeraisAbate) error {
			close(updateAttempted)
			lock := lockFor("identity:abate:7")
			lock.Lock()
			lock.Unlock()
			close(updateDone)
			return nil
		},
	}
	storage := storagePortStub{upload: func(context.Context, string, []byte, string) error {
		close(uploadEntered)
		<-finishUpload
		return nil
	}}
	service := NewAbateService(port, storage)
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	uploadResult := make(chan error, 1)
	go func() {
		_, err := service.UploadFotoAbate(context.Background(), 7, "FAZENDA", "foto.png", png)
		uploadResult <- err
	}()
	<-uploadEntered
	go func() { _ = service.UpdateDadosGeraisAbate(7, domain.DadosGeraisAbate{}) }()
	<-updateAttempted
	select {
	case <-updateDone:
		t.Fatal("update concluiu enquanto upload mantinha o lock de identidade")
	default:
	}
	close(finishUpload)
	if err := <-uploadResult; err != nil {
		t.Fatal(err)
	}
	<-updateDone
}

func TestAbateServiceGeneratesReportWithCalculationsAndPhotos(t *testing.T) {
	observation := "  Animais sem ocorrências.  "
	abate := domain.Abate{
		ID: 7, NomeProprietario: "Dono", NomeFazenda: "Fazenda",
		DadosGeraisAbate: domain.DadosGeraisAbate{Observacao: &observation},
		EtapaFazenda: domain.EtapaFazenda{
			PesoTotal: 1200,
			QuantidadeAnimal: []domain.QtdDenticao{
				{QtdDenticao: domain.DenticaoZero, QtdAnimais: 1},
				{QtdDenticao: domain.DenticaoDois, QtdAnimais: 2},
			},
			Fotos: []domain.FotoAbate{{ID: 10}},
		},
		EtapaFrigorifico: domain.EtapaFrigorifico{
			PesoTotal: 600, Balancao: 900,
			AcabamentoCarcaca: []domain.AcabamentoCarcaca{
				{Acabamento: domain.AcabamentoCarcacaMediano, QtdAnimais: 2},
			},
			ClassificacaoFrigorifico: []domain.ClassificacaoFrigorifico{
				{Classificacao: domain.ClassificacaoFrigorificoBoiMedioNormal, QtdAnimais: 3},
			},
			DistribuicaoPeso: []domain.DistribuicaoPeso{
				{Classificacao: domain.FaixaDistribuicaoPeso20A21Ponto9, QtdAnimais: 2, PesoTotal: 450},
			},
			Fotos: []domain.FotoAbate{{ID: 11}},
		},
	}
	port := abatePortStub{
		findByIDs: func(_ context.Context, ids []int) ([]domain.Abate, error) {
			if len(ids) != 1 || ids[0] != 7 {
				t.Fatalf("ids=%v", ids)
			}
			return []domain.Abate{abate}, nil
		},
		findFotoID: func(id int) (*domain.FotoAbate, error) {
			return &domain.FotoAbate{ID: id, ObjectKey: string(rune(id)), NomeOriginal: "foto.jpg", ContentType: "image/jpeg"}, nil
		},
	}
	storage := storagePortStub{download: func(_ context.Context, _ string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader([]byte("imagem"))), nil
	}}
	generator := abateReportGeneratorStub{generate: func(_ context.Context, reports []output.AbateReport) ([]byte, error) {
		if len(reports) != 1 {
			t.Fatalf("reports=%d", len(reports))
		}
		report := reports[0]
		if report.QuantidadeAnimais != 3 || report.MediaKGFrigorifico != 200 || report.MediaArrobaFrigorifico != 200.0/15 {
			t.Fatalf("médias=%+v", report)
		}
		if report.MediaKGFazenda != 400 || report.RendimentoCarcaca != .5 || report.PesoMedioBalancao != 300 {
			t.Fatalf("rendimentos=%+v", report)
		}
		if report.Esvaziamento != 100 || report.RendimentoBalancao != 2.0/3 {
			t.Fatalf("esvaziamento=%+v", report)
		}
		if len(report.Denticoes) != 5 || len(report.Acabamentos) != 6 || len(report.Classificacoes) != 4 || len(report.DistribuicoesPeso) != 4 {
			t.Fatalf("catálogos incompletos: dentições=%d acabamentos=%d classificações=%d distribuições=%d", len(report.Denticoes), len(report.Acabamentos), len(report.Classificacoes), len(report.DistribuicoesPeso))
		}
		if report.Denticoes[0].Percentual != 33.3 || report.Denticoes[2].Quantidade != 0 || report.Denticoes[2].Percentual != 0 {
			t.Fatalf("dentições=%+v", report.Denticoes)
		}
		if report.Acabamentos[2].Percentual != 66.7 || report.Acabamentos[0].Quantidade != 0 || report.Acabamentos[0].Percentual != 0 {
			t.Fatalf("percentuais=%+v", report)
		}
		if report.Classificacoes[2].Classificacao != domain.ClassificacaoFrigorificoBoiMedioNormal || report.Classificacoes[2].Quantidade != 3 || report.Classificacoes[0].Quantidade != 0 {
			t.Fatalf("classificações=%+v", report.Classificacoes)
		}
		if report.DistribuicoesPeso[1].MediaKG != 225 || report.DistribuicoesPeso[1].MediaArroba != 15 || report.DistribuicoesPeso[1].Percentual != 75 {
			t.Fatalf("distribuição=%+v", report.DistribuicoesPeso[1])
		}
		if report.DistribuicoesPeso[0].Quantidade != 0 || report.DistribuicoesPeso[0].PesoTotal != 0 || report.DistribuicoesPeso[0].MediaKG != 0 || report.DistribuicoesPeso[0].MediaArroba != 0 || report.DistribuicoesPeso[0].Percentual != 0 {
			t.Fatalf("faixa ausente deveria estar zerada: %+v", report.DistribuicoesPeso[0])
		}
		if report.Observacao != "Animais sem ocorrências." || len(report.FotosFazenda) != 1 || len(report.FotosFrigorifico) != 1 {
			t.Fatalf("dados=%+v", report)
		}
		if !strings.HasPrefix(report.FotosFazenda[0].DataURL, "data:image/jpeg;base64,") {
			t.Fatalf("data url=%s", report.FotosFazenda[0].DataURL)
		}
		return []byte("%PDF"), nil
	}}
	pdf, err := NewAbateService(port, storage, generator).GenerateAbateReport(context.Background(), []int{7})
	if err != nil || string(pdf) != "%PDF" {
		t.Fatalf("pdf=%q err=%v", pdf, err)
	}
}

func TestAbateServiceReportAvoidsDivisionByZeroAndUsesObservationFallback(t *testing.T) {
	abate := domain.Abate{ID: 1}
	port := abatePortStub{findByIDs: func(context.Context, []int) ([]domain.Abate, error) {
		return []domain.Abate{abate}, nil
	}}
	generator := abateReportGeneratorStub{generate: func(_ context.Context, reports []output.AbateReport) ([]byte, error) {
		report := reports[0]
		if len(report.Denticoes) != 5 || len(report.Acabamentos) != 6 || len(report.Classificacoes) != 4 || len(report.DistribuicoesPeso) != 4 {
			t.Fatalf("catálogos incompletos: %+v", report)
		}
		values := []float64{report.MediaKGFrigorifico, report.MediaArrobaFrigorifico, report.MediaKGFazenda, report.RendimentoCarcaca, report.PesoMedioBalancao, report.Esvaziamento, report.RendimentoBalancao}
		for _, value := range values {
			if math.IsNaN(value) || math.IsInf(value, 0) || value != 0 {
				t.Fatalf("valor inválido=%v", value)
			}
		}
		if report.Observacao != "Não há observação sobre o abate." {
			t.Fatalf("observação=%q", report.Observacao)
		}
		return []byte("pdf"), nil
	}}
	if _, err := NewAbateService(port, storagePortStub{}, generator).GenerateAbateReport(context.Background(), []int{1}); err != nil {
		t.Fatal(err)
	}
}

func TestAbateServiceReportUsesCanonicalOrderAndIgnoresValuesOutsideCatalog(t *testing.T) {
	abate := domain.Abate{
		ID: 1,
		EtapaFazenda: domain.EtapaFazenda{QuantidadeAnimal: []domain.QtdDenticao{
			{QtdDenticao: domain.DenticaoOito, QtdAnimais: 2},
			{QtdDenticao: 3, QtdAnimais: 99},
		}},
		EtapaFrigorifico: domain.EtapaFrigorifico{
			AcabamentoCarcaca: []domain.AcabamentoCarcaca{
				{Acabamento: domain.AcabamentoCarcacaMedianoUP, QtdAnimais: 2},
				{Acabamento: "FORA DO CATÁLOGO", QtdAnimais: 99},
			},
			ClassificacaoFrigorifico: []domain.ClassificacaoFrigorifico{
				{Classificacao: domain.ClassificacaoFrigorificoBoiPesado, QtdAnimais: 2},
				{Classificacao: "FORA DO CATÁLOGO", QtdAnimais: 99},
			},
			DistribuicaoPeso: []domain.DistribuicaoPeso{
				{Classificacao: domain.FaixaDistribuicaoPesoAcimaDe24, QtdAnimais: 2, PesoTotal: 600},
				{Classificacao: "FORA DO CATÁLOGO", QtdAnimais: 99, PesoTotal: 999},
			},
		},
	}
	port := abatePortStub{findByIDs: func(context.Context, []int) ([]domain.Abate, error) {
		return []domain.Abate{abate}, nil
	}}
	generator := abateReportGeneratorStub{generate: func(_ context.Context, reports []output.AbateReport) ([]byte, error) {
		report := reports[0]
		if report.QuantidadeAnimais != 2 {
			t.Fatalf("quantidade=%d", report.QuantidadeAnimais)
		}
		for index, value := range domain.DenticoesAbate() {
			if report.Denticoes[index].Denticao != value {
				t.Fatalf("ordem dentições=%+v", report.Denticoes)
			}
		}
		for index, value := range domain.AcabamentosCarcacaAbate() {
			if report.Acabamentos[index].Classificacao != value {
				t.Fatalf("ordem acabamentos=%+v", report.Acabamentos)
			}
		}
		for index, value := range domain.ClassificacoesFrigorificoAbate() {
			if report.Classificacoes[index].Classificacao != value {
				t.Fatalf("ordem classificações=%+v", report.Classificacoes)
			}
		}
		for index, value := range domain.FaixasDistribuicaoPesoAbate() {
			if report.DistribuicoesPeso[index].Classificacao != value {
				t.Fatalf("ordem distribuição=%+v", report.DistribuicoesPeso)
			}
		}
		return []byte("pdf"), nil
	}}
	if _, err := NewAbateService(port, storagePortStub{}, generator).GenerateAbateReport(context.Background(), []int{1}); err != nil {
		t.Fatal(err)
	}
}

func TestAbateServiceReportReturnsNotFoundWhenAnyIDIsMissing(t *testing.T) {
	port := abatePortStub{findByIDs: func(context.Context, []int) ([]domain.Abate, error) {
		return []domain.Abate{{ID: 1}}, nil
	}}
	_, err := NewAbateService(port, storagePortStub{}).GenerateAbateReport(context.Background(), []int{1, 2})
	if !errors.Is(err, domain.ErrAbateNaoEncontrado) {
		t.Fatalf("err=%v", err)
	}
}

func TestAbateServiceReportDeduplicatesIDsPreservingOrder(t *testing.T) {
	port := abatePortStub{findByIDs: func(_ context.Context, ids []int) ([]domain.Abate, error) {
		if !reflect.DeepEqual(ids, []int{2, 1}) {
			t.Fatalf("ids consultados=%v", ids)
		}
		return []domain.Abate{{ID: 1}, {ID: 2}}, nil
	}}
	generator := abateReportGeneratorStub{generate: func(_ context.Context, reports []output.AbateReport) ([]byte, error) {
		if len(reports) != 2 || reports[0].Abate.ID != 2 || reports[1].Abate.ID != 1 {
			t.Fatalf("relatórios=%+v", reports)
		}
		return []byte("pdf"), nil
	}}

	if _, err := NewAbateService(port, storagePortStub{}, generator).GenerateAbateReport(context.Background(), []int{2, 1, 2, 1}); err != nil {
		t.Fatal(err)
	}
}
