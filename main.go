package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

var defaultRingBufferSize = 30

var requestRingBuffer ringBuffer

var serverInfo serverInfoType

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

		// curReq := buf.Buffer[(buf.Index-i)%buf.Size]
		curReq := buf.Buffer[newIndex]
		if curReq.Ready {
			reqs = append(reqs, curReq)
		}
	}

	buf.mux.Unlock()

	return reqs
}

func webServer(port string) {
	aWebServer := http.NewServeMux()
	aWebServer.HandleFunc("/", webHandler)

	listener, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Web Server listening on port %s\n", port)
	log.Fatal(http.Serve(listener, aWebServer))
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
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

	log.Printf("WEB: [%s] %s:\n\n", r.RemoteAddr, path)

	// r.Write(log.Writer())
	fmt.Printf("%s\n", fullRequest)

	// r.Write(w)
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

func adminServer(port string) {
	aAdmServer := http.NewServeMux()

	// box := packr.NewBox("./templates")
	fs, _ := fs.Sub(content, "templates")
	aAdmServer.Handle("/", http.FileServer(http.FS(fs)))

	aAdmServer.HandleFunc("/api/", adminHandler)

	listener, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Admin Server listening on port %s\n", port)
	log.Fatal(http.Serve(listener, aAdmServer))
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.EscapedPath()

	log.Printf("ADM: [%s] %s\n", r.RemoteAddr, path)

	// if strings.EqualFold("/api/exit", path) {
	// 	log.Println("Admin Exit Yar")
	// 	fmt.Fprintf(w, "Exiting...\n\n")
	// 	wg.Done()

	// } else
	// if strings.EqualFold("/api/reqs", path) {
	// 	reqs, _ := json.Marshal(requestRingBuffer.Get())

	// 	fmt.Fprintf(w, string(reqs))
	// 	// fmt.Println(string(reqs))

	// } else
	if strings.EqualFold("/api/data", path) {
		data := dataResponse{}
		data.Requests = requestRingBuffer.Get()
		data.ServerInfo = serverInfo

		// rb, _ := json.Marshal(requestRingBuffer.Get())
		// data.Requests = string(rb)

		// si, _ := json.Marshal(serverInfo)
		// data.ServerInfo = string(si)

		dataBytes, _ := json.Marshal(data)
		dataStr := string(dataBytes)

		fmt.Fprintf(w, dataStr)
		fmt.Printf(dataStr)

	} else {
		// http.Error(w, "What you trying to do man?", http.StatusNotFound)
		w.WriteHeader(http.StatusNotFound)
	}
}

func getPorts() (string, string) {
	webPort := "5080"
	adminPort := "5081"

	if len(os.Args) >= 3 { // <prog> <port1> <port2>
		webPort = os.Args[1]
		adminPort = os.Args[2]
	} else {
		log.Printf("Using default web [%s] and admin [%s] ports, to use different ports run: \n\n\t%s <webPort> <adminPort>\n\n", webPort, adminPort, filepath.Base(os.Args[0]))
	}

	return webPort, adminPort
}

func init() {
	requestRingBuffer = ringBuffer{}
	requestRingBuffer.Buffer = make([]historicRequest, defaultRingBufferSize)
	requestRingBuffer.Size = defaultRingBufferSize
	requestRingBuffer.Index = 0

	serverInfo = serverInfoType{}
	serverInfo.Hostname, _ = os.Hostname()
}

func main() {
	webPort, adminPort := getPorts()

	wg.Add(1)
	go webServer(webPort)
	go adminServer(adminPort)

	wg.Wait()
	time.Sleep(time.Second * 1)
}
