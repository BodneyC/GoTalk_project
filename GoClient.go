// package main
// //----------------------------------------------------------------------------------------------
// import (
// 	"fmt"
// 	"net"
// 	"flag"
// 	"bufio"
// 	"time"
// 	"os"
// )
// //----------------------------------------------------------------------------------------------
// var (
// 	CONN_PROT = "tcp"
// 	cmd_CONN_HOST = flag.String("h", "127.0.0.1", "Host IP")
// 	cmd_CONN_PORT = flag.String("p", "6666", "Host port")
// )
// //----------------------------------------------------------------------------------------------
// func main() {
// 	conn, er_dial := net.DialTimeout(CONN_PROT, *cmd_CONN_HOST + ":" + *cmd_CONN_PORT, time.Second * 5)
// 	if er_dial != nil {
// 		fmt.Println("Could not connect to ", cmd_CONN_HOST, " via port ", cmd_CONN_PORT, ".\nError: ", er_dial)
// 	}
// 	defer conn.Close()
//
// 	for {
// 		input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
// 		message, _ := bufio.NewReader(conn).ReadString('\n')
//
// 		if input != nil {
// 			// fmt.Print("Text to send: ")
// 			// input, _ := readBuf.ReadString('\n')
// 			conn.Write([]byte(input))
// 		} else if message != nil {
// 			conn.Read([]byte(message))
// 			fmt.Println("MESSAGE IN: ", message)
// 			// fmt.Println("Message from server: " + message)
// 		} else {}
// 	}
// }
// //----------------------------------------------------------------------------------------------
// func init() {
// 	flag.Parse()
// }
// //----------------------------------------------------------------------------------------------



package main
//----------------------------------------------------------------------------------------------
import (
	"fmt"
	"net"
	"flag"
	"bufio"
	"time"
	"sync"
	"os"
)
//----------------------------------------------------------------------------------------------
var (
	CONN_PROT = "tcp"
	cmd_CONN_HOST = flag.String("h", "127.0.0.1", "Host IP")
	cmd_CONN_PORT = flag.String("p", "6666", "Host port")
	quit = make(chan bool)
)
//----------------------------------------------------------------------------------------------
type Client struct {
	incoming chan string
	outgoing chan string
	readBuf *bufio.Reader
	writeBuf *bufio.Reader
	conn net.Conn
	clientMutex sync.Mutex
}
//----------------------------------------------------------------------------------------------
func NewClient(conn net.Conn) *Client {
	readBuf := bufio.NewReader(conn)
	writeBuf := bufio.NewReader(os.Stdin)

	client := &Client{
		incoming: make(chan string),
		outgoing: make(chan string),
		readBuf: readBuf,
		writeBuf: writeBuf,
		conn: conn,
	}
	return client
}
//----------------------------------------------------------------------------------------------
func (client *Client) Reader() {
	for {
		message, _ := client.readBuf.ReadString('\n')
		client.incoming <- message
	}
}
//----------------------------------------------------------------------------------------------
func (client *Client) Writer() {
	for {
		message, _ := client.writeBuf.ReadString('\n')
		client.outgoing <- message
	}
}
//----------------------------------------------------------------------------------------------
func (client *Client) IO() {
	go client.Reader()
	go client.Writer()
}
//----------------------------------------------------------------------------------------------
func main() {

	conn, er_dial := net.DialTimeout(CONN_PROT, *cmd_CONN_HOST + ":" + *cmd_CONN_PORT, time.Second * 5)
	if er_dial != nil {
		fmt.Println("Could not connect to ", *cmd_CONN_HOST, " via port ", *cmd_CONN_PORT, ".\nError: ", er_dial)
		return
	}
	defer conn.Close()

	client := NewClient(conn)
	client.IO()
	fmt.Print("MESSAGE OUT: ")

	for {
		select {
		case message := <-client.incoming:
			fmt.Println("\nMESSAGE IN: " + message)
			fmt.Print("MESSAGE OUT: ")
		case message := <-client.outgoing:
			fmt.Fprintf(conn, message)
		}
	}
}
//----------------------------------------------------------------------------------------------
func init() {
	flag.Parse()
}
//----------------------------------------------------------------------------------------------
