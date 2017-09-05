/*----------------------------------------------------------------------------------------------
INCOMING DATA FLOW:							OUTGOING DATA FLOW:
	user.Read()								// No need to cycle users so:
		user.incoming <- line				user.Write()
	room.Join									output := <-user.outgoing
		room.incoming <- data
	room.Listen()
		room.Broadcast(data)
	room.Broadcast(data) (send to all)
		user.outgoing <- data
	user.Write()
		user.writer.WriteString(data)
---------------------------------------------------------------------------------------------*/
package main // Executables must always be called 'main'
//----------------------------------------------------------------------------------------------
import (
	"fmt"
	"strconv"
	"github.com/golang/glog"
	"net"
	"sync"
	"bufio" //Read/Write buffers
	"time"
	"flag"
	"strings"
	"os" //os.Exit
	"io" //io.EOF (For user killing)
)
//----------------------------------------------------------------------------------------------
var (
	connections []net.Conn
	listener net.Listener
	room ServerRoom
	port int
	CONN_PROT = "tcp"
	CONN_HOST = "localhost"
)
//----------------------------------------------------------------------------------------------
type UserInfo struct {
	room *ServerRoom
	userID int
	conn net.Conn
	incoming chan string
	outgoing chan string
	readBuf *bufio.Reader
	writeBuf *bufio.Writer
	killUserConnection chan bool
	userMutex sync.Mutex
}
//----------------------------------------------------------------------------------------------
func NewUser(id int, room *ServerRoom, conn net.Conn) *UserInfo{
	readBuf := bufio.NewReader(conn)
	writeBuf := bufio.NewWriter(conn)

	newUser := &UserInfo{
		room: room, // Needed for deletion of 'user' in 'room'...
		userID: id, // ...indexed by id
		conn: conn,
		incoming: make(chan string),
		outgoing: make(chan string),
		readBuf: readBuf,
		writeBuf: writeBuf,
		killUserConnection: make(chan bool),
	}
	glog.Infoln("User added, ID: ", id)
	return newUser // Return to room.Join()
}
//----------------------------------------------------------------------------------------------
func (user *UserInfo) RemoveUser() {
	// Handling of 'user' within 'room'
	room := *user.room
	room.roomMutex.Lock()
	defer room.roomMutex.Unlock()
	delete(room.userTrack, user.userID)

	// Handling of 'user' as 'user'
	user.userMutex.Lock()
	defer user.userMutex.Unlock()
	user.readBuf = nil
	user.writeBuf = nil
	user.conn.Close() // Delete net.Conn for 'user'
	user.killUserConnection <- true
	glog.Infoln("User", user.userID, "removed from room")
}
//----------------------------------------------------------------------------------------------
func (user *UserInfo) Write() {
	for {
		select {
		case <-user.killUserConnection:
			return
		case output := <-user.outgoing:
			user.writeBuf.WriteString(output)
			user.writeBuf.Flush()
		}
	}
}
//----------------------------------------------------------------------------------------------
func (user *UserInfo) Read() {
	for {
		select {
		case <-user.killUserConnection:
			return
		default: //'default' is non-blocking so it executed if the 'select' isn't present
			// Return 'line' and error_MSG. Reads until delimiter '\n'
			input, er_read := user.readBuf.ReadString('\n')
			if er_read != nil {
				if er_read == io.EOF {
					glog.Info("User ", user.userID, " left the room")
					user.RemoveUser()
				} else {
					glog.Errorln("Reading failed (User: ", user.userID, ")")
					time.Sleep(100 * time.Millisecond)
					user.RemoveUser()
				}
			}
			user.incoming <- input
		}
	}
}
//----------------------------------------------------------------------------------------------
func (user *UserInfo) Listen() {
	go user.Read()
	go user.Write()
}
//----------------------------------------------------------------------------------------------
type ServerRoom struct {
	userTrack map[int]*UserInfo // Map for delete() (tracking UserInfo objects)
	userID int // Tracking users, and used as key for delete()^
	entrance chan net.Conn
	incoming chan string
	outgoing chan string
	roomMutex sync.Mutex // For reserving use by a single GoRoutine per object
}
//----------------------------------------------------------------------------------------------
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
//----------------------------------------------------------------------------------------------
func (room *ServerRoom) Join(conn net.Conn) {
	room.roomMutex.Lock()
	defer room.roomMutex.Unlock()

	newuserID := room.userID + 1
	room.userID = newuserID

	user := NewUser(newuserID, room, conn)
	user.Listen()

	// Check to see if key exists ('_' would retrieve the value of the key)
	_, key := room.userTrack[newuserID]
	if !key {
		room.userTrack[newuserID] = user
	}

	go func() {
		for {
			select {
			case <-user.killUserConnection:
				return
			case input := <-user.incoming:
				// Links to room.Listen() and then room.Broadcast()
				room.incoming <- input
			}
		}
	} ()
}
//----------------------------------------------------------------------------------------------
func (room *ServerRoom) Broadcast(input string) {
	for _, user := range room.userTrack {
		user.outgoing <- input
	}
}
//----------------------------------------------------------------------------------------------
func (room *ServerRoom) Listen() {
	go func() {
		for {
			select {
			case input := <-room.incoming:
				temp := strings.Replace(input, "\n", "", 1)
				glog.Info("RECIEVED: " + temp)
				room.Broadcast(input)
			case connIn := <-room.entrance:
				glog.Infoln("New connection to join")
				room.Join(connIn)
			}
		}
	} ()
}
//----------------------------------------------------------------------------------------------
func main() {
	defer func() {
		glog.Info("Closing listener safely")
		for i, connection := range connections {
			if connection != nil {
				glog.Infof("Closing connections ", i)
				connection.Close()
			}
		}
		glog.Flush()
		listener.Close()
	}()

	fmt.Print("Please enter port number to listen on:")
	fmt.Scan(&port)

	listener, er_lis := net.Listen(CONN_PROT, CONN_HOST + ":" + strconv.Itoa(port))

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
//----------------------------------------------------------------------------------------------
// init() is ran before main()
func init() {
	// Parse commandline arguments (needed for glog)
	flag.Parse()
	// Altering commandline arguments
	flag.Lookup("log_dir").Value.Set(".\\logs")
	flag.Lookup("alsologtostderr").Value.Set("true")
	//Potentially unneeded vv
	connections = make([]net.Conn, 0, 10)
}
//----------------------------------------------------------------------------------------------
