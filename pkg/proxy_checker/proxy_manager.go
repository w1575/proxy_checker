package proxy_checker

import (
	"context"
	"time"
)

const (
	DefaultThreads = 128

	FilesBasePath = ""
	FilePathRaw   = FilesBasePath + "/raw.txt"
)

type i18nFn = func(s string) string

type ProxyManager struct {
	validAndInvalidAppend bool // true сохраняет старые данные IP адресов в valid и invalid  файлах
	__                    i18nFn
}

type ProxyManagerSettings struct {
	__ i18nFn // функция-перевода
}

func NewProxyManager() *ProxyManager {
	return &ProxyManager{}
}

func (p *ProxyManager) SetValidAndInvalidAppend(v bool) {
	p.validAndInvalidAppend = v
}

type IsProxyValid struct {
	valid     bool
	proxyData *Data
}

// ProcessProxiesFile проверяет данный ему файл, который содержит данные прокси серверов.
func (p *ProxyManager) ProcessProxiesFile(filePath string) {
	timeNow := time.Now()
	var err error

	proxyRawCh := make(chan string, DefaultThreads)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultsChan := make(chan IsProxyValid, DefaultThreads)

	//go p.rawProxiesProducer(ctx, proxyRawCh) // запустили продюсера
	//
	//workersWG := sync.WaitGroup{}
	//workersWG.Add(DefaultThreads)
	//
	//for range DefaultThreads {
	//	go func() {
	//		p.proxiesChecker(ctx, proxyRawCh, resultsChan)
	//		workersWG.Done()
	//	}()
	//}
	//
	//opMode := os.O_WRONLY | os.O_CREATE
	//if clearResultFiles {
	//	opMode = opMode | os.O_TRUNC
	//}
	//
	//invalidFile, err := os.OpenFile(FilePathInvalid, opMode, 0644)
	//if err != nil {
	//	panic(err) // TODO: сделать обработку ошибки
	//}
	//
	//defer invalidFile.Close()
	//
	//validFile, err := os.OpenFile(FilePathValid, opMode, 0644)
	//if err != nil {
	//	panic(err) // TODO: сделать обработку ошибки
	//}
	//
	//defer validFile.Close()
	//
	//consWG := sync.WaitGroup{}
	//consWG.Go(func() {
	//	for processedProxy := range resultsChan {
	//		if processedProxy.valid {
	//			_, err = validFile.WriteString(processedProxy.proxyData.String() + "\n")
	//			if err != nil {
	//				return
	//			}
	//		} else {
	//			_, err2 := invalidFile.WriteString(processedProxy.proxyData.String() + "\n")
	//			if err2 != nil {
	//				return
	//			}
	//		}
	//	}
	//})
	//workersWG.Wait() //
	//close(resultsChan)
	//consWG.Wait()
	//
	//fmt.Printf(m.__("Прошло времени с момента запуска %d \n"), time.Since(timeNow)/time.Second)
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
