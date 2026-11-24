package proxy_checker

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DefaultThreads                  = 128
	DefaultResponseTimeoutInSeconds = 2

	FilesBasePath       = ""
	FilePathRaw         = FilesBasePath + "/raw.txt"
	FilePathValid       = FilesBasePath + "/valid.txt"
	FilePathInvalid     = FilesBasePath + "/invalid.txt"
	DefaultIpCheckerAPI = "https://api.ipify.org?format=json"
)

type i18nFn = func(s string) string

type ProxyManager struct {
	validAndInvalidAppend bool // true сохраняет старые данные IP адресов в valid и invalid  файлах
	__                    i18nFn
	logger                Logger
	httpClient            CreatableHttpClient
}

func NewProxyManager() *ProxyManager {
	c := NewCustomClient()
	return &ProxyManager{
		httpClient: c,
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
	__ i18nFn // функция-перевода
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

// ProcessProxiesFile проверяет данный ему файл, который содержит данные прокси серверов.
func (p *ProxyManager) ProcessProxiesFile(filePath string) {
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

	producerWg.Wait()       // продюсер асинхронно выполняется, поэтому ждем пока что он завершится
	close(processedProxyCh) // только потом закрываем канал
	processersWg.Wait()
	close(processedProxyCh)
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
	for {
		select {
		case <-ctx.Done():
		case proxyRaw, ok := <-inChan:
			if !ok {
				close(outChan)
				return
			}
			proxyData, proxyErr := FromString(proxyRaw)
			if proxyErr != nil {
				p.logger.Fail(proxyErr.Error())
				break
			} else {
				func() { // тупо чтобы defer можно было заиспользовать и returnы
					timeOutCtx, cancel := context.WithTimeout(ctx, time.Second*2)
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

func getIPCheckerURL() string {
	return DefaultIpCheckerAPI
}

func invalidFilePath() string {
	return FilePathInvalid
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
				_, err := validHandler.WriteString(dataToWrite.proxyData.String())
				if err != nil {
					p.logger.Fail(err.Error())
					return
				}
			} else {
				_, err := validHandler.WriteString(dataToWrite.proxyData.String())
				if err != nil {
					p.logger.Fail(err.Error())
					return
				}
			}
		}
	}

}

func (p *ProxyManager) DownloadProxiesFromSources() {

}

func (p *ProxyManager) DownloadProxiesFromSourcesAndProcesses() {

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
