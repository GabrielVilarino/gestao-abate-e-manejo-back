package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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
	AbatePort   output.AbatePort
	StoragePort output.StoragePort
	cleanupWake chan struct{}
}

func NewAbateService(abatePort output.AbatePort, storagePort output.StoragePort) *AbateService {
	return &AbateService{AbatePort: abatePort, StoragePort: storagePort, cleanupWake: make(chan struct{}, 1)}
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
