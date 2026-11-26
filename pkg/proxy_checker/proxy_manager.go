package proxy_checker

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DefaultThreads                  = 1
	DefaultResponseTimeoutInSeconds = 2

	FilesBasePath   = ""
	FilePathRaw     = FilesBasePath + "raw.txt"
	FilePathValid   = FilesBasePath + "valid.txt"
	FilePathInvalid = FilesBasePath + "invalid.txt"
	FilePathSources = FilesBasePath + "sources.txt"

	DefaultIpCheckerAPI = "https://api.ipify.org?format=json"
)

type i18nFn = func(s string) string

type ProxyManager struct {
	validAndInvalidAppend bool // true сохраняет старые данные IP адресов в valid и invalid  файлах
	__                    i18nFn
	logger                Logger
	httpClient            CreatableHttpClient
}

func NewProxyManager(settings ProxyManagerSettings) *ProxyManager {
	c := NewCustomClient()
	return &ProxyManager{
		__:         settings.TransFn,
		httpClient: c,
		logger:     settings.Logger,
	}
}

// AYYY LAMAO
type CreatableHttpClient interface {
	CloseIdleConnections()
	SetProxy(poxyRaw string)
	DoAndReturnBody(r *http.Request) (string, error)
	CreateClient() CreatableHttpClient
}

type ProxyManagerSettings struct {
	TransFn i18nFn // функция-перевода
	Logger  Logger
}

//// CreateClient "создает" новый клиент. По факту это клон. Но не важно
//func (p *ProxyManager) CreateClient() HttpClient {
//	return p.httpClient
//}

func (p *ProxyManager) SetValidAndInvalidAppend(v bool) {
	p.validAndInvalidAppend = v
}

type proxyCheckResult struct {
	valid     bool
	proxyData Data
}

type Logger interface {
	Warning(message string)
	Info(message string)
	Success(message string)
	Fail(message string)
}

// ProcessProxiesFile1 проверяет данный ему файл, который содержит необработанные данные прокси серверов.
func (p *ProxyManager) ProcessProxiesFile1(filePath string) {
	// зафиксировали время старта
	startTime := time.Now()

	ctx := context.Background()
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	// продюсер читает прокси из файла
	// продюсер отправляет прокси в канал rawProxy
	sourceHandler, handlerErr := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if handlerErr != nil {
		p.logger.Fail(handlerErr.Error())
		return
	}
	defer sourceHandler.Close()
	rawProxyChannel := make(chan string, DefaultThreads)

	producerWg := sync.WaitGroup{}
	producerWg.Go(func() {
		p.rawProxyProducer(ctx, sourceHandler, rawProxyChannel)
	})

	// worker принимает данные из канала rawProxy
	// отправляет в processedProxy
	processedProxyCh := make(chan proxyCheckResult, DefaultThreads)
	processersWg := sync.WaitGroup{}
	processersWg.Add(DefaultThreads)

	for range DefaultThreads {
		go func() {
			p.proxyChecker(ctx, rawProxyChannel, processedProxyCh)
			processersWg.Done()
		}()
	}

	// ну и consumer-writer записывает данные в резудьтирующий файл
	resultWriterWG := sync.WaitGroup{}
	resultWriterWG.Go(func() {
		p.resultsWriter(ctx, processedProxyCh)
	})

	producerWg.Wait() // продюсер асинхронно выполняется, поэтому ждем пока что он завершится
	//close(processedProxyCh) // только потом закрываем канал
	processersWg.Wait()
	//close(processedProxyCh)
	resultWriterWG.Wait()
	// зафиксировали время завершения
	p.logger.Info(
		fmt.Sprintf("Файл со списком прокси был обработан за %v", time.Since(startTime)*time.Second))
}

func (p *ProxyManager) rawProxyProducer(
	ctx context.Context,
	fileHandler *os.File,
	rawChan chan<- string,
) {
	sc := bufio.NewScanner(fileHandler)
	defer close(rawChan)
	for sc.Scan() {
		select {
		case <-ctx.Done():
			p.logger.Fail(ctx.Err().Error())
			return
		case rawChan <- sc.Text():
		}
	}
	return
}

func (p *ProxyManager) proxyChecker(
	ctx context.Context,
	inChan <-chan string,
	outChan chan<- proxyCheckResult,
) {
	defer close(outChan)
	for {
		select {
		case <-ctx.Done():
		case proxyRaw, ok := <-inChan:
			if !ok {
				return
			}
			proxyData, proxyErr := FromString(proxyRaw)
			if proxyErr != nil {
				p.logger.Fail(proxyErr.Error())
				break
			} else {
				func() { // тупо чтобы defer можно было заиспользовать и returnы
					timeOutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
					defer cancel()

					req, reqErr := http.NewRequestWithContext(
						timeOutCtx,
						http.MethodGet,
						getIPCheckerURL(),
						nil,
					)

					if reqErr != nil {
						p.logger.Fail(reqErr.Error())
						return
					}

					client := p.httpClient.CreateClient()
					defer client.CloseIdleConnections()
					client.SetProxy(proxyRaw)
					body, err := client.DoAndReturnBody(req)
					if err != nil {
						return // ????
					}

					result := proxyCheckResult{}
					result.proxyData = *proxyData
					result.valid = strings.Contains(body, proxyRaw)
					outChan <- result
				}()
			}
		}
	}
}

// sourceFileHandler обрабатывает файл и посылает каждый прокси на проверку в канал, который возвращает функция.
func (p *ProxyManager) sourceFileHandler(ctx context.Context, path string) (chan string, *sync.WaitGroup) {
	var (
		fileHandler *os.File
		err         error
	)
	outChan := make(chan string, DefaultThreads)

	sourceWG := &sync.WaitGroup{}
	//sourceWG.Add(1)
	sourceWG.Go(func() {
		//defer sourceWG.Done()
		defer close(outChan)
		fileHandler, err = os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			p.logger.Fail(err.Error())
		}
		defer fileHandler.Close()
		sc := bufio.NewScanner(fileHandler)
		for sc.Scan() {
			txt := sc.Text()
			select {
			case <-ctx.Done():
				//p.logger.Fail(ctx.Err().Error())
				return
			case outChan <- txt:
				//fmt.Println(txt)
			}
		}
	})

	return outChan, sourceWG
}

func (p *ProxyManager) rawProxyChecker(ctx context.Context, rawChan <-chan string) (chan proxyCheckResult, *sync.WaitGroup) {
	outChan := make(chan proxyCheckResult, DefaultThreads)
	wg := &sync.WaitGroup{}

	wg.Add(DefaultThreads)
	for range DefaultThreads {
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case proxy, isOpen := <-rawChan:
				// А вот этот момент я упустил. Не знали (или я забыл?!) что он сойдет с ума и будет до посинения "получать" данные пустые
				if !isOpen {
					return
				}
				proxyData, err := FromString(proxy)

				if err != nil {
					p.logger.Fail(err.Error())
				}

				res := proxyCheckResult{
					valid:     rand.Intn(1) == 1,
					proxyData: *proxyData,
				}
				outChan <- res
			}
		}()
	}

	return outChan, wg
}

func (pm *ProxyManager) ProcessProxiesFile(ctx context.Context, filePath string) {
	// зафиксировали время старта
	//startTime := time.Now()

	//var cancel context.CancelFunc
	//defer cancel()

	var (
		rawProxyChan    chan string
		sourceHandlerWG *sync.WaitGroup
	)

	rawProxyChan, sourceHandlerWG = pm.sourceFileHandler(ctx, filePath)

	var (
		checkedProxyChan chan proxyCheckResult
		checkerWG        *sync.WaitGroup
	)
	checkedProxyChan, checkerWG = pm.rawProxyChecker(ctx, rawProxyChan)

	writerWg := sync.WaitGroup{}

	writerWg.Go(func() {
		var (
			err                                  error
			validFileHandler, invalidFileHandler *os.File
		)
		validFileHandler, err = os.OpenFile(validFilePath(), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			panic(pm.__("Не могу записать в валидный файл"))
		}
		defer validFileHandler.Close()

		invalidFileHandler, err = os.OpenFile(invalidFilePath(), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			panic(pm.__("Не могу записать в невалидный файл"))
		}
		defer invalidFileHandler.Close()

		select {
		case <-ctx.Done():
			return
		case checkedData := <-checkedProxyChan:
			if checkedData.valid {
				_, err = validFileHandler.WriteString(checkedData.proxyData.String() + "\n")
				if err != nil {
					pm.logger.Warning(pm.__("Не могу записать данные в валидный файл"))
				}
			} else {
				_, err = invalidFileHandler.WriteString(checkedData.proxyData.String() + "\n")
				if err != nil {
					pm.logger.Warning(pm.__("Не могу записать данные в невалидный файл"))
				}
			}
		}
	})

	sourceHandlerWG.Wait() // дождались продюсерра проксей
	//close(rawProxyChan)     // закрыли его канал

	checkerWG.Wait()        // подождали все воркеры-чекеры проксей
	close(checkedProxyChan) // закрыли канал чекеров
	writerWg.Wait()         // ну записунчик должен будет закрыть по контексту поэтому просто подождем если вдруг он
	// отстает
	pm.logger.Info("")
}

func getIPCheckerURL() string {
	return DefaultIpCheckerAPI
}

func invalidFilePath() string {
	return FilePathValid
}

func validFilePath() string {
	return FilePathInvalid
}

func (p *ProxyManager) resultsWriter(
	ctxt context.Context,
	inChan <-chan proxyCheckResult,
) {
	validHandler, errIn := os.OpenFile(validFilePath(), os.O_WRONLY|os.O_CREATE, 0644)
	if errIn != nil {
		p.logger.Fail(errIn.Error())
		return
	}
	defer validHandler.Close()

	invalidHandler, errVal := os.OpenFile(invalidFilePath(), os.O_WRONLY|os.O_CREATE, 0644)
	if errVal != nil {
		p.logger.Fail(errVal.Error())
		return
	}

	defer invalidHandler.Close()
	for {
		select {
		case <-ctxt.Done():
		case dataToWrite := <-inChan:
			if dataToWrite.valid {
				_, err := validHandler.WriteString(dataToWrite.proxyData.String() + "\n")
				if err != nil {
					p.logger.Fail(err.Error())
					return
				}
			} else {
				_, err := validHandler.WriteString(dataToWrite.proxyData.String() + "\n")
				if err != nil {
					p.logger.Fail(err.Error())
					return
				}
			}
		}
	}
}

// sourcesPath получение путя к файлу в котором все источники находятся.
func (p *ProxyManager) sourcesPath() string {
	return FilePathSources
}
func (p *ProxyManager) rwPath() string {
	return FilePathRaw
}

// DownloadProxiesFromSourcesToRawFile просто скачивает данные проксей из всех источников в RAW файл
func (p *ProxyManager) DownloadProxiesFromSourcesToRawFile(truncateRawFile bool) error {
	var (
		err                         error
		rawFileHandler, fileHandler *os.File
	)
	fileHandler, err = os.OpenFile(p.sourcesPath(), os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	defer fileHandler.Close()

	fileMode := os.O_WRONLY | os.O_CREATE

	if truncateRawFile {
		fileMode += os.O_TRUNC
	}

	rawFileHandler, err = os.OpenFile(p.rwPath(), fileMode, 0644)

	if err != nil {
		return err
	}
	defer rawFileHandler.Close()

	sourcesScan := bufio.NewScanner(fileHandler)
	for sourcesScan.Scan() {
		var localErr error

		sourceLink := sourcesScan.Text()
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		req, localErr := http.NewRequestWithContext(ctx, http.MethodGet, sourceLink, nil)
		if localErr != nil {
			return localErr
		}
		body, localErr := p.httpClient.DoAndReturnBody(req)
		if localErr != nil {
			return localErr
		}

		_, localErr = rawFileHandler.WriteString(body)
		if localErr != nil {
			return localErr
		}
	}

	return err
}

func (p *ProxyManager) DownloadFromSourcesAndProcess() {
	err := p.DownloadProxiesFromSourcesToRawFile(false)
	if err != nil {
		return
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	p.ProcessProxiesFile(ctx, getRawFilePath())
}

func getRawFilePath() string {
	return FilePathRaw
}

// CountOfValid возвращает количество рабочих прокси.
func (p *ProxyManager) CountOfValid() int {
	return 0
}

// CountOfInvalid возвращает число не рабочих прокси.
func (p *ProxyManager) CountOfInvalid() int {
	return 0
}

// CountOfUnprocessed возвращает общее число прокси.
func (p *ProxyManager) CountOfUnprocessed() int {
	return 0
}
