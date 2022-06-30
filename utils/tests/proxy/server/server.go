package server

import (
	"crypto/tls"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests/proxy/server/certificate"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	clilog "github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

type httpResponse func(rw http.ResponseWriter, req *http.Request)

func handleReverseProxyHttps(reverseProxy *httputil.ReverseProxy) httpResponse {
	return func(rw http.ResponseWriter, req *http.Request) {
		clilog.Info("*********************************************************")
		clilog.Info("Scheme:  ", "HTTPS")
		clilog.Info("Host:    ", req.Host)
		clilog.Info("Method:  ", req.Method)
		clilog.Info("URI:     ", req.RequestURI)
		clilog.Info("Agent:   ", req.UserAgent())
		clilog.Info("*********************************************************")
		reverseProxy.ServeHTTP(rw, req)
	}
}

func getReverseProxyHandler(targetUrl string) (*httputil.ReverseProxy, error) {
	clilog.Info("Reverse proxy URL:", targetUrl)
	var err error
	var target *url.URL
	target, err = url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}
	origHost := target.Host
	d := func(req *http.Request) {
		req.URL.Host = origHost
		req.Host = origHost
		req.URL.Scheme = target.Scheme
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	proxyErrLogger := log.New(os.Stdout, "PROXY-LOGGER", log.Ldate|log.Ltime|log.Lshortfile)
	p := &httputil.ReverseProxy{Director: d, Transport: tr, ErrorLog: proxyErrLogger}
	return p, nil
}

func removeProxyHeaders(r *http.Request) {
	r.RequestURI = "" // this must be reset when serving a request with the client
	// If no Accept-Encoding header exists, Transport will add the headers it can accept
	// and would wrap the response body with the relevant reader.
	r.Header.Del("Accept-Encoding")
	// curl can add that, see
	// https://jdebp.eu./FGA/web-proxy-connection-header.html
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")
	// Connection, Authenticate and Authorization are single hop Header:
	// http://www.w3.org/Protocols/rfc2616/rfc2616.txt
	// 14.10 Connection
	//   The Connection general-header field allows the sender to specify
	//   options that are desired for that particular connection and MUST NOT
	//   be communicated by proxies over further connections.
	r.Header.Del("Connection")
}

func copyHeaders(dst, src http.Header) {
	for k := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func httpProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/" {
		w.WriteHeader(200)
	} else {
		removeProxyHeaders(r)
		t := &http.Transport{}
		resp, err := t.RoundTrip(r)
		if err != nil {
			clilog.Error(err)
		}
		origBody := resp.Body
		defer origBody.Close()
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			clilog.Error(err)
		}
		if err := resp.Body.Close(); err != nil {
			clilog.Error("Can't close response body %v", err)
		}
	}
}

type testProxy struct {
}

func (t *testProxy) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.RequestURI == "/" {
		responseWriter.WriteHeader(200)
	} else {
		host := request.URL.Host
		request.Host = "https://" + request.Host
		targetSiteCon, err := net.Dial("tcp", host)
		if err != nil {
			clilog.Error(err)
			return
		}
		hij, ok := responseWriter.(http.Hijacker)
		if !ok {
			http.Error(responseWriter, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}
		proxyClient, _, err := hij.Hijack()
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
			return
		}

		proxyClient.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
		targetTCP, targetOK := targetSiteCon.(*net.TCPConn)
		proxyClientTCP, clientOK := proxyClient.(*net.TCPConn)
		if targetOK && clientOK {
			go copyAndClose(targetTCP, proxyClientTCP)
			go copyAndClose(proxyClientTCP, targetTCP)
		}
	}
}

func copyAndClose(dst, src *net.TCPConn) {
	if _, err := io.Copy(dst, src); err != nil {
		clilog.Error(err)
	}
	dst.CloseWrite()
	src.CloseRead()
}

func GetProxyHttpPort() string {
	port := "8099"
	if httpPort := os.Getenv("PROXY_HTTP_PORT"); httpPort != "" {
		port = httpPort
	}
	return port
}

func prepareHTTPSHandling(handler *httputil.ReverseProxy) (httpsMux *http.ServeMux, absPathCert, absPathKey string, err error) {
	// We can use the same handler for both HTTP and HTTPS
	httpsMux = http.NewServeMux()
	httpsMux.HandleFunc("/", handleReverseProxyHttps(handler))
	absPathCert, absPathKey, err = CreateNewServerCertificates()
	return
}

// Creates a server cerf file and cert key file.
// Returns the absolute path of these two files.
func CreateNewServerCertificates() (certFilePath, keyCertFilePath string, err error) {
	certFilePath, err = filepath.Abs(certificate.CertFile)
	if errorutils.CheckError(err) != nil {
		return
	}
	keyCertFilePath, err = filepath.Abs(certificate.KeyFile)
	if errorutils.CheckError(err) != nil {
		return
	}

	if _, err = os.Stat(certFilePath); os.IsNotExist(err) {
		err = certificate.CreateNewCert(certFilePath, keyCertFilePath)
		if err != nil {
			return
		}
	}
	errorutils.CheckError(err)
	return
}

func GetProxyHttpsPort() string {
	port := "1024"
	if httpPort := os.Getenv(tests.HttpsProxyEnvVar); httpPort != "" {
		port = httpPort
	}
	return port
}

func startHttpsReverseProxy(proxyTarget string, requestClientCerts bool) error {
	handler, err := getReverseProxyHandler(proxyTarget)
	if err != nil {
		return err
	}
	// Starts a new Go routine
	httpsMux, absPathCert, absPathKey, err := prepareHTTPSHandling(handler)
	if err != nil {
		return err
	}

	if requestClientCerts {
		server := &http.Server{
			Addr:    ":" + GetProxyHttpsPort(),
			Handler: httpsMux,
			TLSConfig: &tls.Config{
				ClientAuth: tls.RequireAnyClientCert,
			},
		}
		err = server.ListenAndServeTLS(absPathCert, absPathKey)
	} else {
		err = http.ListenAndServeTLS(":"+GetProxyHttpsPort(), absPathCert, absPathKey, httpsMux)
	}

	return err
}

func StartLocalReverseHttpProxy(artifactoryUrl string, requestClientCerts bool) {
	if artifactoryUrl == "" {
		artifactoryUrl = "http://localhost:8081/artifactory/"
	}
	artifactoryUrl = utils.AddTrailingSlashIfNeeded(artifactoryUrl)
	err := startHttpsReverseProxy(artifactoryUrl, requestClientCerts)
	// Since this function is always executed in its own go routine,
	// we panic when an error occurs.
	panic(err)
}

func StartHttpProxy() {
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", httpProxyHandler)
	port := GetProxyHttpPort()
	err := http.ListenAndServe(":"+port, httpMux)
	if err != nil {
		panic(err)
	}
}

func StartHttpsProxy() {
	port := GetProxyHttpsPort()
	err := http.ListenAndServe(":"+port, &testProxy{})
	if err != nil {
		panic(err)
	}
}
