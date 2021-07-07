package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/marv2097/siprocket"
	"github.com/pires/go-proxyproto"
)

func RunServer(addr string, sslAddr string, ssl map[string]string, proxyp string) chan error {
	errs := make(chan error)

	if proxyp == "" {
		go func() {
			fmt.Printf("Echo server listening on port %s.\n", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				errs <- err
			}

		}()

		go func() {
			fmt.Printf("Echo server listening on ssl port %s.\n", sslAddr)
			if err := http.ListenAndServeTLS(sslAddr, ssl["cert"], ssl["key"], nil); err != nil {
				errs <- err
			}
		}()

	} else {
		// enable proxy protocol on both sockets
		go func() {
			fmt.Printf("WS server with proxy proto listening on port %s.\n", addr)

			tcpSrv := http.Server{
				Addr: addr,
			}

			tcpln, err := net.Listen("tcp", addr)
			if err != nil {
				errs <- err
			}

			proxyListener := &proxyproto.Listener{
				Listener:          tcpln,
				ReadHeaderTimeout: 10 * time.Second,
			}
			defer proxyListener.Close()

			if err := tcpSrv.Serve(proxyListener); err != nil {
				errs <- err
			}

		}()

		// TODO: refactor
		go func() {
			fmt.Printf("WSS server with proxy proto listening on port %s.\n", sslAddr)

			tlsSrv := http.Server{
				Addr: sslAddr,
			}

			tlsln, err := net.Listen("tcp", sslAddr)
			if err != nil {
				errs <- err
			}

			proxyListener := &proxyproto.Listener{
				Listener:          tlsln,
				ReadHeaderTimeout: 10 * time.Second,
			}
			defer proxyListener.Close()

			if err := tlsSrv.ServeTLS(proxyListener, ssl["cert"], ssl["key"]); err != nil {
				errs <- err
			}

		}()

	}

	return errs
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool {
		return true
	},
	Subprotocols: []string{"sip"},
}

func handler(wr http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s | %s %s\n", req.RemoteAddr, req.Method, req.URL)
	if websocket.IsWebSocketUpgrade(req) {
		serveWebSocket(wr, req)
	} else {
		log.Fatalf("Can not establish ws connection\n")
	}
}

func parseSipRequest(msg []byte, remoteAddr string) []byte {
	sip := siprocket.Parse(msg)
	via := sip.Via[0]
	remAddr := strings.Split(remoteAddr, ":")

	sipTo := "<sip:" + string(sip.To.User) + "@" + string(sip.To.Host) + ">"
	sipFrom := "<sip:" + string(sip.From.User) + "@" + string(sip.From.Host) + ">"

	response := "SIP/2.0 200 OK\r\n" +
		"Via: " + string(via.Src) + ";received=" + remAddr[0] + "\r\n" +
		"From: " + sipFrom + ";tag=" + string(sip.From.Tag) + "\r\n" +
		"To: " + sipTo + ";tag=37GkEhwl6" + "\r\n" + // fake To tag
		"Call-ID: " + string(sip.CallId.Value) + "\r\n" +
		"CSeq: " + string(sip.Cseq.Id) + " REGISTER\r\n" +
		"Contact: " + string(sip.Contact.Src) + "\r\n" +
		"Content-Length: 0\r\n\r\n"

	fmt.Print(response + "\n")

	return []byte(response)
}

func serveWebSocket(wr http.ResponseWriter, req *http.Request) {
	connection, err := upgrader.Upgrade(wr, req, nil)
	if err != nil {
		fmt.Printf("%s | %s\n", req.RemoteAddr, err)
		return
	}

	defer connection.Close()
	fmt.Printf("%s | upgraded to websocket\n", req.RemoteAddr)

	var message []byte

	if err == nil {
		var messageType int
		var resp []byte

		for {
			messageType, message, err = connection.ReadMessage()
			if err != nil {
				break
			}

			if messageType == websocket.TextMessage {
				fmt.Printf("%s | txt | %s\n", req.RemoteAddr, message)

				// parse SIP msg
				resp = parseSipRequest(message, req.RemoteAddr)

			} else { // we don't expect to receive bin requests..
				fmt.Printf("%s | bin | %d byte(s)\n", req.RemoteAddr, len(message))
			}

			err = connection.WriteMessage(messageType, resp)
			if err != nil {
				break
			}
		}
	}

	if err != nil {
		fmt.Printf("%s | %s\n", req.RemoteAddr, err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sslport := os.Getenv("SSLPORT")
	if sslport == "" {
		sslport = "8443"
	}

	proxyp := os.Getenv("PROXYP")

	http.HandleFunc("/", http.HandlerFunc(handler))

	errs := RunServer(":"+port, ":"+sslport, map[string]string{
		"cert": "cert.pem",
		"key":  "key.pem",
	}, proxyp)

	erros := <-errs

	for {
		if erros != nil {
			log.Printf("Could not start serving service due to (error: %s)", erros)
		}
	}

}
