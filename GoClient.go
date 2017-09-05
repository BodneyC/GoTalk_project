package main
//----------------------------------------------------------------------------------------------
import (
	"fmt"
	"net"
	"flag"
	"bufio"
	"time"
	"os"
)
//----------------------------------------------------------------------------------------------
var (
	CONN_PROT = "tcp"
	cmd_CONN_HOST = flag.String("h", "127.0.0.1", "Host IP")
	cmd_CONN_PORT = flag.String("p", "6666", "Host port")
	cmd_USER_NAME = flag.String("n", "BodneyC", "Username")
	quit = make(chan bool)
)
//----------------------------------------------------------------------------------------------
type Client struct {
	incoming chan string
	outgoing chan string
	readBuf *bufio.Reader
	writeBuf *bufio.Reader
	conn net.Conn
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
		message, er_reader := client.readBuf.ReadString('\n')
		if er_reader == nil {
			client.incoming <- message
		}
	}
}
//----------------------------------------------------------------------------------------------
func (client *Client) Writer() {
	for {
		message, er_writer := client.writeBuf.ReadString('\n')
		if er_writer == nil {
			client.outgoing <- message
		}
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
	fmt.Print("<", *cmd_USER_NAME, "> ")

	for {
		select {
		case message := <-client.incoming:
			fmt.Print("\n" + message)
			fmt.Print("<", *cmd_USER_NAME, "> ")
		case message := <-client.outgoing:
			fmt.Fprintf(conn, "<" + *cmd_USER_NAME + "> " + message)
		}
	}
}
//----------------------------------------------------------------------------------------------
func init() {
	flag.Parse()
}
//----------------------------------------------------------------------------------------------
