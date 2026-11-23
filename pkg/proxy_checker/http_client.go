package proxy_checker

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
)

type CustomClient struct {
	http.Client
}

func NewCustomClient() *CustomClient {
	c := CustomClient{
		//Transport: &http.Transport{}
	}
	c.Timeout = DefaultResponseTimeoutInSeconds

	return &c
}

func (c *CustomClient) DoAndReturnBody(r *http.Request) (string, error) {
	resp, err := c.Client.Do(r)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		cErr := Body.Close()
		if cErr != nil {
			panic(cErr) // lol
		}
	}(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func (c *CustomClient) SetProxy(proxy string) {
	proxyUrl, _ := url.Parse(proxy)
	//t := c.Transport
	c.Transport = &http.Transport{
		Proxy:             http.ProxyURL(proxyUrl),
		ForceAttemptHTTP2: false,
		TLSNextProto:      make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}
}
