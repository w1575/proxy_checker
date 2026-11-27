package proxy_checker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"proxy_checker/pkg/logger"
	"testing"
	"time"
)

func Test_ProxyFactoryJustWorks(t *testing.T) {
	NewProxyManager(ProxyManagerSettings{})
}

//func Test_HTTPClientCloneJustWorks(t *testing.T) {
//	c := ProxyManager{httpClient: &CustomClient{}}
//	//link := &c.httpClient
//	firstLink := c.httpClient.CreateClient()
//	anotherLink := c.httpClient.CreateClient()
//	fmt.Println(&firstLink, *&firstLink, &anotherLink, *&anotherLink)
//}

func makeFileWithGivenProxies(proxies map[string]string, namePattern string) (string, error) {
	var (
		err            error
		tmpFileHandler *os.File
	)
	tmpFileHandler, err = os.CreateTemp(os.TempDir(), "makeFileWithGivenProxies")

	if err != nil {
		return "", err
	}

	defer tmpFileHandler.Close()
	for _, val := range proxies {
		_, err = tmpFileHandler.WriteString(val + "\n")
		if err != nil {
			return "", err
		}
	}
	return tmpFileHandler.Name(), err
}

var processTestFileDataProvider = map[string]string{
	"http://127.0.0.1:8981": "http://127.0.0.1:8981", //: true,
	"http://127.0.0.1:8982": "http://127.0.0.1:8982", //: true,
	"http://127.0.0.1:8983": "http://127.0.0.1:8983", //: false,
	"http://127.0.0.1:8984": "http://127.0.0.1:8984", //: true,
	"http://127.0.0.1:8985": "http://127.0.0.1:8985", //: false,
	"http://127.0.0.1:8986": "http://127.0.0.1:8986", //: true,
} // ор. просто делаю чтобы быстрее чем со срезами работало.

//func Test_rawProxyProducer(t *testing.T) {
//	filePath, err := makeFileWithGivenProxies(processTestFileDataProvider, "--")
//	t.Cleanup(func() {
//		os.Remove(filePath)
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	pm := NewProxyManager(ProxyManagerSettings{})
//	ctx := context.Background()
//	fileHandler, handlerErr := os.OpenFile(filePath, os.O_RDONLY, 0644)
//
//	if handlerErr != nil {
//		t.Fatal(handlerErr)
//	}
//
//	testChan := make(chan string, 6)
//
//	pm.rawProxyProducer(ctx, fileHandler, testChan)
//
//	for line := range testChan {
//		_, ok := processTestFileDataProvider[line]
//		if !ok {
//			t.Errorf("по какой-то причине прокси %s нет в списке", line)
//		}
//	}
//}

// Она хранит ожидание - валидный прокси или нет, и время выполнения (тестирование таймаута)
type proxyExecData struct {
	isValid     bool
	executeTime time.Duration
}
type PseudoHTTPClient struct {
	proxiesMap      map[string]proxyExecData
	currentProxy    string
	responseBody    string
	responseTimeSec int
}

func (p *PseudoHTTPClient) CloseIdleConnections() {}

func (p *PseudoHTTPClient) SetProxy(proxy string) {
	p.currentProxy = proxy
}

func (p *PseudoHTTPClient) shouldReturn(response string, responseTimeSec int) {
	p.responseBody = response
	p.responseTimeSec = responseTimeSec
}

func (p *PseudoHTTPClient) DoAndReturnBody(r *http.Request) (string, error) {
	respData := p.proxiesMap[p.currentProxy]

	if respData.executeTime > DefaultResponseTimeoutInSeconds {
		return "", fmt.Errorf("долго ждем, товарищи")
	}
	if p.responseBody != "" {
		return p.currentProxy, nil
	}
	return "", nil
}

func (p *PseudoHTTPClient) CreateClient() CreatableHttpClient {
	return p
}

var proxyCheckerDataProvider = map[string]proxyExecData{
	"http://127.0.0.1:8080": {isValid: true, executeTime: DefaultResponseTimeoutInSeconds + 1},
	//ваилдный
	"http://127.0.0.2:8079": {isValid: false, executeTime: 5},
	// не проходит по таймауту
	"http//127.0.0.1:800": {isValid: false, executeTime: 1},
	// двоеточие
	"http://127.0.0:8080": {isValid: false, executeTime: 1},
	// последний октет (не помню как правильно )
	"http://266.0.0.1:8080": {isValid: false, executeTime: 1},
	// за все возможные рамки приличия выходит
	"http://0.0.0.1": {isValid: false, executeTime: 1},
	// ой
}

func Test_proxyChecker(t *testing.T) {
	client := &PseudoHTTPClient{}
	client.proxiesMap = proxyCheckerDataProvider

	inChan := make(chan string, len(proxyCheckerDataProvider))
	outChan := make(chan proxyCheckResult, len(proxyCheckerDataProvider))

	p := NewProxyManager(ProxyManagerSettings{})
	p.httpClient = client
	//p.logger = logger.NewConsoleLogger()
	p.logger = logger.NewNullLogger()
	bgCtxt := context.Background()
	//ctx, cancel := context.WithTimeout(bgCtxt, time.Second*DefaultResponseTimeoutInSeconds)
	//defer cancel()
	for ip, _ := range proxyCheckerDataProvider {
		inChan <- ip
	}
	close(inChan)
	p.proxyChecker(bgCtxt, inChan, outChan)
	for result := range outChan {
		proxyRaw := result.proxyData.String()
		expectation := proxyCheckerDataProvider[proxyRaw].isValid
		if result.valid != expectation {
			t.Errorf("Ожидалось, что %s будет %t", proxyRaw, expectation)
		}
	}
}

func Test_DownloadProxiesFromSources(t *testing.T) {

}
