package httpwrap

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

const DefaultClientTimeout = 10 * time.Second

// Client is a wrapper around http.Client that provides simplified HTTP methods.
type Client struct {
	httpClient  *http.Client
	proxy       string
	bearerToken string
}

// NewClient creates a new Client with the specified timeout.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultClientTimeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 10,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

// DoRequest sends an HTTP request with the given method, URL, body, and headers.
func (c *Client) DoRequest(method, url string, bodyReader io.Reader, headers Header) (io.ReadCloser, int, error) {
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, -1, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	// Default Content-Type
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, -1, err
	}
	if resp.StatusCode >= 300 {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logrus.Errorf("error closing response body: %v\n", err)
			}
		}(resp.Body)
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, -1, fmt.Errorf("error reading response: %w", err)
		}
		httpErr := HTTPError{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       respBody,
			Err:        fmt.Errorf("HTTP %d: %s ", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
		httpErr.Log()
		return nil, -1, httpErr
	}
	limitCount := -1
	if resp.Header.Get("X-Rate-Limit-Remaining") != "" {
		//convert the string to an int
		limitCount, err = strconv.Atoi(resp.Header.Get("X-Rate-Limit-Remaining"))
		if err != nil {
			limitCount = -1
		}
	}
	return resp.Body, limitCount, nil
}

// Get sends an HTTP GET request.
func (c *Client) Get(baseURL string, urlParams url.Values, headers map[string]string, obj any) (any, int, error) {
	// Convert map[string]interface{} to url.Values
	// Parse the base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, -1, fmt.Errorf("invalid base URL: %w", err)
	}

	// Set the query parameters
	parsedURL.RawQuery = urlParams.Encode()

	respBody, limitCount, err := c.DoRequest(http.MethodGet, parsedURL.String(), nil, headers)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logrus.Errorf("error closing response body: %v\n", err)
		}
	}(respBody)
	if err != nil {
		return nil, limitCount, err
	}
	if obj == nil {
		obj = make(map[string]interface{})
	}
	err = json.NewDecoder(respBody).Decode(&obj)
	if err != nil {
		return nil, limitCount, err
	}

	return obj, limitCount, nil
}

// Post sends an HTTP POST request with a JSON body.
func (c *Client) Post(url string, body interface{}, headers map[string]string, obj any) (any, int, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, -1, err
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}
	respBody, limitCount, err := c.DoRequest(http.MethodPost, url, bodyReader, headers)
	if err != nil {
		return nil, limitCount, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logrus.Errorf("error closing response body: %v\n", err)
		}
	}(respBody)

	if obj == nil {
		obj = make(map[string]interface{})
	}
	err = json.NewDecoder(respBody).Decode(&obj)
	if err != nil {
		return nil, limitCount, err
	}
	return obj, limitCount, nil
}

// SetTimeout sets the timeout for the singleton http.Client.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetProxy sets the proxy for the singleton http.Client.
func (c *Client) SetProxy(proxyAddr string) error {
	if proxyAddr == "" {
		c.httpClient.Transport = &http.Transport{
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			DialContext: (&net.Dialer{
				Timeout: c.httpClient.Timeout,
			}).DialContext,
		}
	} else if strings.HasPrefix(proxyAddr, "http") {
		urlproxy, err := url.Parse(proxyAddr)
		if err != nil {
			return err
		}
		c.httpClient.Transport = &http.Transport{
			Proxy:        http.ProxyURL(urlproxy),
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			DialContext: (&net.Dialer{
				Timeout: c.httpClient.Timeout,
			}).DialContext,
		}
		c.proxy = proxyAddr
	} else if strings.HasPrefix(proxyAddr, "socks5") {
		baseDialer := &net.Dialer{
			Timeout:   c.httpClient.Timeout,
			KeepAlive: c.httpClient.Timeout,
		}
		proxyURL, err := url.Parse(proxyAddr)
		if err != nil {
			return err
		}

		// username password
		username := proxyURL.User.Username()
		password, _ := proxyURL.User.Password()

		// ip and port
		host := proxyURL.Hostname()
		port := proxyURL.Port()

		dialSocksProxy, err := proxy.SOCKS5("tcp", host+":"+port, &proxy.Auth{User: username, Password: password}, baseDialer)
		if err != nil {
			return errors.New("error creating socks5 proxy :" + err.Error())
		}
		if contextDialer, ok := dialSocksProxy.(proxy.ContextDialer); ok {
			dialContext := contextDialer.DialContext
			c.httpClient.Transport = &http.Transport{
				DialContext: dialContext,
			}
		} else {
			return errors.New("failed type assertion to DialContext")
		}
		c.proxy = proxyAddr
		return nil
	} else {
		return errors.New("only support http(s) or socks5 protocol")
	}
	return nil
}

func (c *Client) GetCookies(url *url.URL) []*http.Cookie {
	return c.httpClient.Jar.Cookies(url)
}

func (c *Client) SetCookies(url *url.URL, cookies []*http.Cookie) {
	c.httpClient.Jar.SetCookies(url, cookies)
}

func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

func (c *Client) WithJar() *Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		logrus.Errorf("error creating cookie jar: %v\n", err)
		return c
	}
	c.httpClient.Jar = jar
	return c
}

// WithBearerToken sets the Bearer Token for the client.
func (c *Client) WithBearerToken(token string) *Client {
	c.bearerToken = token
	c.httpClient.Transport = &BearerTransport{
		Transport: c.httpClient.Transport,
		Token:     token,
	}
	return c
}
