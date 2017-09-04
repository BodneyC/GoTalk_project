package main
//---------------------------------------------------------------------
import (
	"fmt"
	"strconv"
	"github.com/golang/glog"
	"net"
	"time"
	"flag"
	"os" //os.Exit
)
//---------------------------------------------------------------------
var (
	connections []net.Conn
	listener net.Listener
	quit chan bool
)
//---------------------------------------------------------------------
const (
	CONN_PROT = "tcp"
	CONN_HOST = "localhost"
)
//---------------------------------------------------------------------
func init() {
	flag.Parse()
	flag.Lookup("log_dir").Value.Set("D:\\Users\\BenJC\\Documents\\1_Current\\Programming\\GO\\src\\chatThing\\logs")
  flag.Lookup("alsologtostderr").Value.Set("true")

	connections = make([]net.Conn, 0, 10)
	quit = make(chan bool)
}
//---------------------------------------------------------------------
func serverInit() {
	for {
		connection, er_acc := listener.Accept()
		if er_acc != nil {
			glog.Error("Could not accept connection " + strconv.Itoa(len(connections) - 1))
			connection.Close()
			time.Sleep(time.Millisecond * 100)
			continue
		} else {
			connections = append(connections, connection)
			glog.Info("Connection " + strconv.Itoa(len(connections) - 1) + " accepted")
		}
		// go handleconnection(connection, len(connections) - 1)
	}
}
//---------------------------------------------------------------------
// func handleconnection(conn net.Conn, id int) error
type ChatRoom struct {
	entrance chan net.Conn
	incoming chan string
	outgoing chan string
}
//---------------------------------------------------------------------
func NewRoom() *ChatRoom {
	newroom := &ChatRoom{
		entrance: make(chan net.Conn),
		incoming: make(chan string),
		outgoing: make(chan string),
	}
	return newroom
}
//---------------------------------------------------------------------
func (room *ChatRoom) Join(conn net.Conn) {

}
//---------------------------------------------------------------------
func (room *ChatRoom) Broadcast(input string) {

}
//---------------------------------------------------------------------
func (room *ChatRoom) Listen() {
	go func() {
		for {
			select {
			case input := <-room.incoming:
				glog.Infoln("RECIEVED: " + input)
				room.Broadcast(input)
			case connection := <-room.entrance:
				glog.Infoln("New connection to join")
				room.Join(connection)
			}
		}
	} ()
}
//---------------------------------------------------------------------
func main() {
	var port int
	defer glog.Flush()

	defer func() {
		glog.Info("Closing listener safely")
		for i, connection := range connections {
			if connection != nil {
				glog.Infof("Closing connections ", i)
				connection.Close()
			}
		}
	}()

	fmt.Println("Please enter port number to listen on:")
	fmt.Scan(&port)

	listener, er_lis := net.Listen(CONN_PROT, CONN_HOST + ":" + strconv.Itoa(port))
	defer listener.Close()

	if er_lis != nil {
		glog.Fatalf("Fatal error in listener init (Port: %s)", strconv.Itoa(port))
		os.Exit(1)
	} else {
		glog.Info("Listener init (Port: ", strconv.Itoa(port), ")")
	}

	room := NewRoom()
	room.Listen()

	serverInit()

}
//---------------------------------------------------------------------
