package proxy_checker

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	DefaultThreads                  = 128
	DefaultResponseTimeoutInSeconds = 2

	FilesBasePath = ""
	FilePathRaw   = FilesBasePath + "/raw.txt"
)

type i18nFn = func(s string) string

type ProxyManager struct {
	validAndInvalidAppend bool // true сохраняет старые данные IP адресов в valid и invalid  файлах
	__                    i18nFn
	logger                Logger
}

func NewProxyManager() *ProxyManager {
	return &ProxyManager{}
}

type ProxyManagerSettings struct {
	__ i18nFn // функция-перевода
}

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
		p.rawProxyProducer(sourceHandler, rawProxyChannel)
	})

	// worker принимает данные из канала rawProxy
	// отправляет в processedProxy
	processedProxyCh := make(chan proxyCheckResult, DefaultThreads)
	processersWg := sync.WaitGroup{}
	processersWg.Add(DefaultThreads)

	for range DefaultThreads {
		go func() {
			p.proxyChecker(rawProxyChannel, processedProxyCh)
			processersWg.Done()
		}()
	}

	// ну и consumer-writer записывает данные в резудьтирующий файл
	resultWriterWG := sync.WaitGroup{}
	resultWriterWG.Go(func() {
		p.resultsWriter(processedProxyCh)
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

func (p *ProxyManager) rawProxyProducer(fileHandler *os.File, rawChan chan<- string) {
	sc := bufio.NewScanner(fileHandler)
	for sc.Scan() {
		txt := sc.Text()
		rawChan <- txt
	}
	close(rawChan)
}

func (p *ProxyManager) proxyChecker(inChan <-chan string, outChan chan<- proxyCheckResult) {

}

func (p *ProxyManager) resultsWriter(inChan <-chan proxyCheckResult) {

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
