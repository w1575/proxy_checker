package proxy_checker

import (
	"context"
	"os"
	"testing"
)

func Test_ProxyFactoryJustWorks(t *testing.T) {
	NewProxyManager()
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

func Test_rawProxyProducer(t *testing.T) {
	filePath, err := makeFileWithGivenProxies(processTestFileDataProvider, "--")
	t.Cleanup(func() {
		os.Remove(filePath)
	})
	if err != nil {
		t.Fatal(err)
	}

	pm := NewProxyManager()
	ctx := context.Background()
	fileHandler, handlerErr := os.OpenFile(filePath, os.O_RDONLY, 0644)

	if handlerErr != nil {
		t.Fatal(handlerErr)
	}

	testChan := make(chan string, 6)

	pm.rawProxyProducer(ctx, fileHandler, testChan)

	for line := range testChan {
		_, ok := processTestFileDataProvider[line]
		if !ok {
			t.Errorf("по какой-то причине прокси %s нет в списке", line)
		}
	}
}
