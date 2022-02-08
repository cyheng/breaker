package plugin

import (
	"breaker/pkg/netio"
	"io"
	"net"
	"net/http"
	"strings"
)

type HttpProxy struct {
}

func (p *HttpProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodConnect {
		p.ConnectHandler(rw, req)
	} else {
		p.HttpHandler(rw, req)
	}
}
func (hp *HttpProxy) HttpHandler(rw http.ResponseWriter, req *http.Request) {
	transport := http.DefaultTransport

	// step 1
	outReq := new(http.Request)
	*outReq = *req // this only does shallow copies of maps

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outReq.Header.Set("X-Forwarded-For", clientIP)
	}

	// call request
	res, err := transport.RoundTrip(outReq)

	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer res.Body.Close()
	// copy header
	for key, value := range res.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}

	rw.WriteHeader(res.StatusCode)
	io.Copy(rw, res.Body)
}
func (hp *HttpProxy) ConnectHandler(rw http.ResponseWriter, req *http.Request) {
	hj, ok := rw.(http.Hijacker)
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	client, _, err := hj.Hijack()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	remote, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		http.Error(rw, "Failed", http.StatusBadRequest)
		client.Close()
		return
	}
	client.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	go netio.StartTunnel(remote, client)
}
