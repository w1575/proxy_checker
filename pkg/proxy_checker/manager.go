package proxy_checker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"proxy_checker/pkg/logger"
	"strings"
	"sync"
	"time"
	//"ttv_watcher/internal/components/logger"
)

// TODO: Тут сильная наркомания получилась. Нужно немного переделать
// Всё в бардак превратилось
const (
	PathDir                         = "files/proxy"
	FilePathRaw                     = PathDir + "/raw.txt"
	FilePathValid                   = PathDir + "/valid.txt"
	FilePathInvalid                 = PathDir + "/invalid.txt"
	FileDownloadedRawFromSourceSite = PathDir + "/git.txt"

	DirPermissions = 0755

	LinkIpApi = "https://api.ipify.org?format=json"
	// TODO: не очень. нужно например в env вынести, но в таком случае формат может отличаться. хзхзхзх
	DefaultTimeout = 2 * time.Second
	DefaultThreads = 128

	LinkGithubRaw = "https://raw.githubusercontent.com/iplocate/free-proxy-list/refs/heads/main/all-proxies.txt"
)

type Manager struct {
	dirPath,
	rawFilePath,
	validFilePath string
	Logger Logger // да приложение все больше напоминает франкенштейна
	client HttpClient
}

func NewManager() *Manager {
	p := &Manager{
		dirPath:       PathDir,
		rawFilePath:   FilePathRaw,
		validFilePath: FilePathValid,
		Logger:        logger.NewConsoleLogger(),
		client:        NewCustomClient(),
	}

	exist, err := isPathExist(PathDir)
	if err != nil {
		return nil
	}

	if !exist {
		err1 := os.Mkdir(PathDir, DirPermissions)
		if err1 != nil {
			return nil
		}
	}

	return p
}

func (m *Manager) SetRawFilePath(path string) {
	m.rawFilePath = path
}

func (m *Manager) GetRawFilePath() string {
	if m.rawFilePath != "" {
		return m.rawFilePath
	}

	return FilePathRaw
}

func (m *Manager) SetClient(c HttpClient) {
	m.client = c
}

func (m *Manager) UpdateProxiesFromSourceSite() {
	_ = os.Remove(FileDownloadedRawFromSourceSite)
	_ = os.Remove(FilePathInvalid)
	_ = os.Remove(FilePathValid)
	_ = os.Remove(FilePathRaw)
	// плеватюшки
	m.Logger.Info("Скачиваем файл с гитхаба.")
	m.DownloadProxiesFromSourceSite()
	err := os.Rename(FileDownloadedRawFromSourceSite, FilePathRaw)
	if err != nil {
		return
	}
	m.Logger.Info("Начинаем проверку прокси серверов.")
	m.ProcessRawFile()
}

func (m *Manager) DownloadProxiesFromSourceSite() {
	m.Logger.Info("Начали скачивать файл со списком прокси с гитхаба.")
	var err error
	var resp *http.Response
	c := http.Client{}
	resp, err = c.Get(LinkGithubRaw)
	defer resp.Body.Close()
	if err != nil {
		m.Logger.Fail(fmt.Sprintf("при скачивании списка прокси произошла ошибка: %s", err))
		return
	}
	respScanner := bufio.NewScanner(resp.Body)

	var writeToFile *os.File
	writeToFile, err = os.OpenFile(FileDownloadedRawFromSourceSite, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		m.Logger.Fail(fmt.Sprintf("при скачивании списка прокси произошла ошибка: %s", err))
		return
	}

	defer writeToFile.Close()
	for respScanner.Scan() {
		proxyRaw := respScanner.Text() + "\n"
		_, err = writeToFile.WriteString(proxyRaw)
		if err != nil {
			m.Logger.Fail(fmt.Sprintf("При записи прокси в файл произошла ошибка: %s", err))
			return
		}
	}

	return
}

type Logger interface {
	Warning(message string)
	Info(message string)
	Success(message string)
	Fail(message string)
}

func isPathExist(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, err
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func (m *Manager) isPathExist() (bool, error) {
	return isPathExist(m.dirPath)
}

func createFile(path string) error {
	f, err := os.Create(path)
	defer f.Close()

	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) CreateValidFile() error {
	return createFile(FilePathValid)
}

// checkAvailability непосредственно проверяет работать прокси или нет
func (m *Manager) checkAvailability(ctx context.Context, proxy string) (*Data, error) {
	var err error
	var d *Data
	d, err = FromString(proxy)

	if d == nil || err != nil {
		return &Data{}, err
	}

	var req, reqE = http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		LinkIpApi,
		nil,
	)

	client := NewCustomClient() // SOLID? Да пошёл он этот солид

	if reqE != nil {
		panic(reqE) //TODO без паники
	}

	client.SetProxy(proxy)

	body, err := client.DoAndReturnBody(req)
	client.CloseIdleConnections()

	if !strings.Contains(body, d.IP) {
		err = fmt.Errorf("увы, получен скорее всего пустой ответ")
	}

	m.Logger.Info(fmt.Sprintf("Получили ответ от API %s", body))

	return d, err
}

type IsProxyValid struct {
	proxy     string
	valid     bool
	proxyData *Data
}

func (i IsProxyValid) String() string {
	return fmt.Sprintf("IP: %s, VALID: %t", i.proxy, i.valid)
}

func rawProxiesProducer(ctx context.Context, proxyRawCh chan<- string) {
	rawFile, err := os.OpenFile(FilePathRaw, os.O_RDONLY, 0644)
	if err != nil {
		panic("Raw file не удалось открыть")
	}
	defer rawFile.Close()
	rawScanner := bufio.NewScanner(rawFile)
	for rawScanner.Scan() {
		proxy := rawScanner.Text()
		select {
		case <-ctx.Done():
			return
		case proxyRawCh <- proxy:
		}
	}
	close(proxyRawCh) // в целом пускай от него мы и пляшем.
}

func (m *Manager) proxiesChecker(
	ctx context.Context,
	proxyRawCh <-chan string,
	resultsChan chan IsProxyValid,
) {
	var (
		workerErr error
		proxyData *Data
	)
	for {
		select {
		case <-ctx.Done():
			return
		case proxy, ok := <-proxyRawCh:
			if !ok {
				//close(resultsChan)
				return
			}
			availableContext, cancelAvailable := context.WithTimeout(ctx, time.Second*2)
			proxyData, workerErr = m.checkAvailability(availableContext, proxy)
			cancelAvailable()
			select {
			case <-ctx.Done():
				return
			case resultsChan <- IsProxyValid{proxyData: proxyData, valid: workerErr == nil}:
			}
		}
	}
}

func (m *Manager) ProcessRawFile() {
	timeNow := time.Now()
	var err error

	proxyRawCh := make(chan string, DefaultThreads)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultsChan := make(chan IsProxyValid, DefaultThreads)

	go rawProxiesProducer(ctx, proxyRawCh) // запустили продюсера

	workersWG := sync.WaitGroup{}
	workersWG.Add(DefaultThreads)

	for range DefaultThreads {
		go func() {
			m.proxiesChecker(ctx, proxyRawCh, resultsChan)
			workersWG.Done()
		}() // небольшой велосипед для того, чтобы можно было ниже понять, когда зыкрывать канал
	}

	// я выше хотел сделать вейтгруп, даже загрузился. но тут подумал, а можно же просто подождать
	// закрытие канала. ЕСЛИ ЧЕСТНО Я УЖЕ НИ В ЧЕМ НЕ УВЕРЕН!

	invalidFile, err := os.OpenFile(FilePathInvalid, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		panic(err) // TODO: сделать обработку ошибки
	}

	defer invalidFile.Close()

	validFile, err := os.OpenFile(FilePathValid, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		panic(err) // TODO: сделать обработку ошибки
	}

	defer validFile.Close()

	consWG := sync.WaitGroup{}
	consWG.Go(func() {
		for processedProxy := range resultsChan {
			if processedProxy.valid {
				_, err = validFile.WriteString(processedProxy.proxyData.String() + "\n")
				if err != nil {
					return
				}
			} else {
				_, err2 := invalidFile.WriteString(processedProxy.proxyData.String() + "\n")
				if err2 != nil {
					return
				}
			}
		}
	})
	workersWG.Wait() //
	close(resultsChan)
	consWG.Wait()

	fmt.Printf("Прошло времени с момента запуска %d \n", time.Since(timeNow)/time.Second)
}

func (m *Manager) rawIpsProducer(rawIpChan chan<- string) {
	rawFile, err := os.OpenFile(FilePathRaw, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		m.Logger.Fail(err.Error())
		return
	}
	defer rawFile.Close()
	scan := bufio.NewScanner(rawFile)
	for scan.Scan() {
		proxy := scan.Text()
		rawIpChan <- proxy
	}

	close(rawIpChan)
}

func (m *Manager) GetRawProxyCount() int {
	return getLinesCount(FilePathRaw) // FilePathRaw
}

func (m *Manager) GetValidProxyCount() int {
	return getLinesCount(FilePathValid) // FilePathRaw
}

func (m *Manager) GetInvalidProxyCount() int {
	return getLinesCount(FilePathInvalid) // FilePathRaw
}

func getLinesCount(path string) int {
	file, _ := os.Open(path) // TODO обработать ошибку
	reader := bufio.NewReader(file)
	linesCount := 0
	for {
		str, err := reader.ReadString('\n')
		//fmt.Println(str)
		if str != "" {
			linesCount++
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return 0
		}

	}
	return linesCount
}

type HttpClient interface {
	DoAndReturnBody(req *http.Request) (string, error)
	//DoAndReturnBodyWithContext(ctx context.Context, req *http.Request) (string, error)
	//DoWithContext(ctx context.Context, r *http.Request) (string, error)
	SetProxy(proxy string)
}
