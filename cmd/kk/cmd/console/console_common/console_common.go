package console_common

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	"strings"
)

type Cluster struct {
	Kind              string `yaml:"kind,omitempty" json:"kind,omitempty"`
	ApiVersion        string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	metav1.ObjectMeta `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	Spec kubekeyapiv1alpha2.ClusterSpec `yaml:"spec,omitempty" json:"spec,omitempty"`
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketWriter struct {
	WsConn *websocket.Conn
}

//var clientConn *websocket.Conn

func SetupCaptureBuffer() *WebSocketWriter {
	captureBuffer := &WebSocketWriter{}
	readerOut, writerOut, _ := os.Pipe()
	os.Stdout = writerOut
	outReader := bufio.NewReader(readerOut)
	// 新开线程，监听管道信息，并输出到websocket
	go func() {
		for {
			line, _, err := outReader.ReadLine()
			if err != nil {
				break
			}
			if captureBuffer.WsConn != nil {
				captureBuffer.WsConn.WriteMessage(websocket.TextMessage, line)
			}
		}
	}()
	return captureBuffer
}

func FormatErrorMessage(err error) string {
	msg := err.Error()
	if !strings.HasPrefix(msg, "error: ") {
		msg = fmt.Sprintf("error: %s", msg)
	}
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	return msg
}
