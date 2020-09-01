package server

import (
	"log"
	"net/http"
	model "xx/model"

	"github.com/gorilla/websocket"
)

//var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan model.Message)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var chatRooms = make([]chatRoom, 0, 5)

type chatRoom struct {
	roomNumber        string
	clientConnections map[*websocket.Conn]bool
}

func newChatRoom(roomNr string, conn map[*websocket.Conn]bool) chatRoom {

	return chatRoom{
		roomNumber:        roomNr,
		clientConnections: conn,
	}
}

func StartServer() {

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("http server stated on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()

	//clients[ws] = true

	for {
		var msg model.Message
		err := ws.ReadJSON(&msg)

		switch action := msg.Action; action {
		case "create":
			createChatRoom(msg.RoomNumber, make(map[*websocket.Conn]bool), ws)
		case "join":
			joinChatRoom(msg.RoomNumber, ws)
		case "leave":
			leaveChatRoom(msg.RoomNumber, ws)
		case "send":
			broadcast <- msg
		default:
		}

		if err != nil {
			log.Printf("error: %v", err)
			break
		}
	}
}

func createChatRoom(roomNr string, client map[*websocket.Conn]bool, conn *websocket.Conn) {
	client[conn] = true
	chatRoom := newChatRoom(roomNr, client)
	chatRooms = append(chatRooms, chatRoom)
	log.Println("chatRoom created ")
}

func joinChatRoom(roomNr string, conn *websocket.Conn) {
	for _, chatRoom := range chatRooms {
		if chatRoom.roomNumber == roomNr {
			chatRoom.clientConnections[conn] = true
		}
	}
}

func leaveChatRoom(roomNr string, conn *websocket.Conn) {
	for _, chatRoom := range chatRooms {
		if chatRoom.roomNumber == roomNr {
			_, ok := chatRoom.clientConnections[conn]
			if ok {
				delete(chatRoom.clientConnections, conn)
			}
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for _, chatRoom := range chatRooms {
			if chatRoom.roomNumber == msg.RoomNumber {
				for client := range chatRoom.clientConnections {
					err := client.WriteJSON(msg.Text)
					if err != nil {
						log.Printf("error: %v", err)
						client.Close()
					}
				}
			}
		}
	}
}
