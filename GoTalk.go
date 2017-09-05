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
	cmd_CONN_HOST = flag.String("ip", "127.0.0.1", "Host IP")
	cmd_CONN_PORT = flag.String("p", "6666", "Host port")
	cmd_USER_NAME = flag.String("n", "BodneyC", "Username")
)
//----------------------------------------------------------------------------------------------
type Client struct {
	outgoing chan string
	writeBuf *bufio.Reader
	conn net.Conn
}
//----------------------------------------------------------------------------------------------
func NewClient(conn net.Conn) *Client {
	writeBuf := bufio.NewReader(os.Stdin)

	client := &Client{
		outgoing: make(chan string),
		writeBuf: writeBuf,
		conn: conn,
	}
	return client
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

	for {
		fmt.Print("<", *cmd_USER_NAME, "> ")
		select {
		case message := <-client.outgoing:
			fmt.Fprintf(conn,  "<" + *cmd_USER_NAME + "> " + message)
		}
	}
}
//----------------------------------------------------------------------------------------------
func init() {
	flag.Parse()
}
//----------------------------------------------------------------------------------------------
