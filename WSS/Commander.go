package WSS

import (
	"server/Auth"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var wssID2Conn = make(map[primitive.ObjectID]*websocket.Conn) // map of clients [id] -> *websocket.Conn
var wssConn2ID = make(map[*websocket.Conn]primitive.ObjectID) // map of clients [*websocket.Conn] -> id

var Handler = websocket.New(func(c *websocket.Conn) {
	var (
		wssMsg WSSMessage
		err    error
	)
	for {
		wssMsg = WSSMessage{}
		err = c.ReadJSON(&wssMsg)
		if err != nil {
			break
		}
		userID := Auth.SESS[wssMsg.Token].USERID
		wssMsg.Token = ""
		if !userID.IsZero() {
			wssID2Conn[userID] = c
			wssConn2ID[c] = userID
			continue
		}
		// If connection is not authorized, send error message
		if wssConn2ID[c].IsZero() {
			c.WriteJSON(fiber.Map{"type": "error", "message": "unauthorized"})
			continue
		}
		// Send to admin channel if target is not set, or not sent
		if wssMsg.TargetID.IsZero() || wssID2Conn[wssMsg.TargetID] == nil {
			c.WriteJSON(fiber.Map{"type": "error", "message": "Target is not available, or not connected! Please try later."})
			continue
		}
		receiverID := wssMsg.TargetID
		wssMsg.TargetID = wssConn2ID[c]
		// Response to sender(admin)
		if isResponse(wssMsg) && Auth.IsAdmin(receiverID) {
			if err = wssID2Conn[receiverID].WriteJSON(wssMsg); err != nil {
				c.WriteJSON(fiber.Map{"type": "error", "message": "Receiver is not available, or not connected! Please try later."})
			}
			continue
		}
		// If non-admin sends command to another user
		if !Auth.IsAdmin(wssMsg.TargetID) {
			c.WriteJSON(fiber.Map{"type": "error", "message": "not allowed"})
			continue
		}
		// Admin (became wssMessage.targetID) sends command to user targetID
		if err = wssID2Conn[receiverID].WriteJSON(wssMsg); err != nil {
			c.WriteJSON(fiber.Map{"type": "error", "message": "Target is offline"})
		}
	}
	// remove client from the maps
	disID := wssConn2ID[c]
	if !disID.IsZero() {
		// set user state to offline
		delete(Auth.SESS, Auth.LINU[disID])
		delete(Auth.LINU, disID)
		// remove session of that user
		delete(wssID2Conn, disID)
		delete(wssConn2ID, c)
	}
})

var Upgrader = func(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

type WSSMessage struct {
	Token    string             `json:"token,omitempty"`
	Type     string             `json:"type,omitempty"`
	TargetID primitive.ObjectID `json:"targetid,omitempty"`
	Message  string             `json:"message,omitempty"`
	Dir      string             `json:"dir,omitempty"`
}

const (
	mtError      = "error"   // error message
	mtSuccess    = "success" // success message
	mtCMDRes     = "cmdres"  // command response
	mtCMD        = "cmd"     // command
)

func isResponse(msg WSSMessage) bool {
	return msg.Type == mtError || msg.Type == mtSuccess || msg.Type == mtCMDRes
}
