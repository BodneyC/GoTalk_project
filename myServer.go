package main
//---------------------------------------------------------------------
import (
	"fmt"
	"strconv"
	"github.com/golang/glog"
	"net"
	"sync"
	"bufio"
	"time"
	"flag"
	"os" //os.Exit
)
//---------------------------------------------------------------------
var (
	connections []net.Conn
	listener net.Listener
	quit chan bool
	room ServerRoom
	CONN_PROT = "tcp"
	CONN_HOST = "localhost"
)
//---------------------------------------------------------------------
type UserInfo struct {
	userID int
	incoming chan string
	outgoing chan string
	readBuf *bufio.Reader
	writeBuf *bufio.Writer
	userMutex sync.Mutex
}
//---------------------------------------------------------------------
func NewUser(id int, conn net.Conn) *UserInfo{
	readBuf := bufio.NewReader(conn)
	writeBuf := bufio.NewWriter(conn)

	newUser := &UserInfo{
		userID: id,
		incoming: make(chan string),
		outgoing: make(chan string),
		readBuf: readBuf,
		writeBuf: writeBuf,
	}
	glog.Infoln("User added, ID: ", id)
	return newUser // Return to room.Join()
}
//---------------------------------------------------------------------
type ServerRoom struct {
	userTrack map[int]*UserInfo // Map for delete() (tracking UserInfo objects)
	userID int // Tracking users, and used as key for delete()^
	entrance chan net.Conn
	incoming chan string
	outgoing chan string
	roomMutex sync.Mutex // For reserving use by a single GoRoutine per object
}
//---------------------------------------------------------------------
func NewRoom() *ServerRoom {
	newroom := &ServerRoom{
		userTrack: make(map[int]*UserInfo),
		userID: -1,
		entrance: make(chan net.Conn),
		incoming: make(chan string),
		outgoing: make(chan string),
	}
	glog.Infoln("New room created")
	return newroom
}
//---------------------------------------------------------------------
func (room *ServerRoom) Join(conn net.Conn) {
	room.roomMutex.Lock()
	defer room.roomMutex.Unlock()

	newuserID := room.userID + 1
	room.userID = newuserID

	user := NewUser(newuserID, conn)

	// Check to see if key exists ('_' would retrieve the value of the key)
	_, key := room.userTrack[newuserID]
	if !key {
		room.userTrack[newuserID] = user
	}

}
//---------------------------------------------------------------------
func (room *ServerRoom) Broadcast(input string) {
	for _, user := range room.userTrack {
		user.outgoing <- input
	}
}
//---------------------------------------------------------------------
func (room *ServerRoom) Listen() {
	go func() {
		for {
			select {
			case input := <-room.incoming:
				glog.Infoln("RECIEVED: " + input)
				room.Broadcast(input)
			case connIn := <-room.entrance:
				glog.Infoln("New connection to join")
				room.Join(connIn)
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
		// 'connection' into 'room.entrance' causes 'Join' via 'Listen'
		room.entrance <- connection
	}

}
//---------------------------------------------------------------------
func init() {
	flag.Parse()
	flag.Lookup("log_dir").Value.Set("D:\\Users\\BenJC\\Documents\\1_Current\\Programming\\GO\\src\\chatThing\\logs")
  flag.Lookup("alsologtostderr").Value.Set("true")

	connections = make([]net.Conn, 0, 10)
	quit = make(chan bool)
}
//---------------------------------------------------------------------
