package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	webPort               int
	adminPort             int
	defaultRingBufferSize = 30
	requestRingBuffer     ringBuffer
	serverInfo            serverInfoType
)

//go:embed templates
var content embed.FS

type serverInfoType struct {
	Hostname string
}

func (sit *serverInfoType) MarshalJSON() ([]byte, error) {
	return json.Marshal(sit)
}

type ringBuffer struct {
	Buffer []historicRequest
	Size   int
	Index  int
	mux    sync.Mutex
}

type historicRequest struct {
	Ready      bool
	RemoteAddr string
	Request    string
	RequestURI string
	DateTime   string
	Method     string
}

type dataResponse struct {
	Requests   []historicRequest
	ServerInfo serverInfoType
}

func (buf *ringBuffer) Add(req historicRequest) {
	buf.mux.Lock()
	buf.Index = (buf.Index + 1) % buf.Size
	buf.Buffer[buf.Index] = req
	buf.mux.Unlock()
}

func (buf *ringBuffer) Get() []historicRequest {
	buf.mux.Lock()

	var reqs []historicRequest

	for i := range buf.Buffer {
		newIndex := buf.Index - i

		if newIndex < 0 {
			newIndex = buf.Size + newIndex
		}

		curReq := buf.Buffer[newIndex]
		if curReq.Ready {
			reqs = append(reqs, curReq)
		}
	}

	buf.mux.Unlock()

	return reqs
}

func webServer() {
	aWebServer := http.NewServeMux()
	aWebServer.HandleFunc("/", webHandler)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", webPort))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("  Web Server listening on port %d\n", webPort)
	log.Fatal(http.Serve(listener, aWebServer))
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	buf := new(bytes.Buffer)
	r.Write(buf)

	fmt.Fprintf(buf, "%s\n", body)

	fullRequest := buf.String()

	path := r.URL.EscapedPath()

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("WEB: [%s] %s:\n\n", r.RemoteAddr, path))
	sb.WriteString(fmt.Sprintf("%s\n", fullRequest))
	log.Print(sb.String())

	w.Header().Add("X-WEBECHO-HOSTNAME", serverInfo.Hostname)
	fmt.Fprintf(w, "%s\n", fullRequest)

	req := historicRequest{}
	req.Ready = true
	req.RemoteAddr = r.RemoteAddr
	req.RequestURI = r.URL.String()
	req.DateTime = time.Now().Format("2006.01.02 15:04:05")
	req.Method = r.Method
	req.Request = fullRequest

	requestRingBuffer.Add(req)
}

func adminServer() {
	aAdmServer := http.NewServeMux()

	fs, _ := fs.Sub(content, "templates")
	aAdmServer.Handle("/", http.FileServer(http.FS(fs)))

	aAdmServer.HandleFunc("/api/", adminHandler)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", adminPort))
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Admin Server listening on port %d\n", adminPort)
	log.Fatal(http.Serve(listener, aAdmServer))
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.EscapedPath()

	log.Printf("ADM: [%s] %s\n", r.RemoteAddr, path)

	if strings.EqualFold("/api/data", path) {
		data := dataResponse{}
		data.Requests = requestRingBuffer.Get()
		data.ServerInfo = serverInfo

		dataBytes, _ := json.Marshal(data)
		dataStr := string(dataBytes)

		fmt.Fprint(w, dataStr)
		fmt.Print(dataStr)

	} else {
		// http.Error(w, "What you trying to do man?", http.StatusNotFound)
		w.WriteHeader(http.StatusNotFound)
	}
}

func init() {
	flag.IntVar(&webPort, "webport", 8080, "web server listen port")
	flag.IntVar(&adminPort, "adminport", 8081, "admin server listen port")

	requestRingBuffer = ringBuffer{}
	requestRingBuffer.Buffer = make([]historicRequest, defaultRingBufferSize)
	requestRingBuffer.Size = defaultRingBufferSize
	requestRingBuffer.Index = 0

	serverInfo = serverInfoType{}
	serverInfo.Hostname, _ = os.Hostname()
}

func main() {
	flag.Parse()

	go webServer()
	go adminServer()

	select {}
}
