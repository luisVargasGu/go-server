package types

import (
	"encoding/base64"
	"log"
	"strconv"
	"sync"
	"time"
	"user/server/services/utils"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

var iceServers = []webrtc.ICEServer{

	{

		URLs: []string{"stun:stun.l.google.com:19302"},
	},
}

var config = webrtc.Configuration{

	ICEServers: iceServers,
}

var trackLocals = LocalTracks{}

type LocalTracks struct {
	mu     sync.RWMutex
	tracks map[string]*webrtc.TrackLocalStaticRTP
}

type Room struct {
	mu        sync.RWMutex
	ID        int                     `json:"id"`
	Name      string                  `json:"name"`
	ChannelID int                     `json:"channel_id"`
	Clients   map[*Client]*ClientInfo `json:"-"`
	Bus       *EventBus               `json:"-"`
}

type RoomInfo struct {
	RoomID   int        `json:"roomId"`
	RoomName string     `json:"roomName"`
	Users    []UserInfo `json:"users"`
}

type RoomInfoMessage struct {
	Type    string   `json:"type"`
	Payload RoomInfo `json:"payload"`
}

type RoomsResponse struct {
	Rooms []*Room `json:"rooms"`
}

type RoomStore interface {
	GetRoomsInChannel(channelID int) ([]*Room, error)
	CreateRoom(room *Room) error
	DeleteRoom(roomID int) (*Room, error)
}

func (r *Room) Run() {
	// Create event channels
	registerCh := make(chan Event, 100)
	broadcastCh := make(chan Event, 100)
	unRegisterCh := make(chan Event, 100)

	// Subscribe to events
	r.Bus.Subscribe(EventRegister, registerCh)
	r.Bus.Subscribe(EventBroadcast, broadcastCh)
	r.Bus.Subscribe(EventUnregister, unRegisterCh)

	go func() {
		for range time.NewTicker(time.Second * 3).C {
			dispatchKeyFrame(r)
		}
	}()

	for {
		select {
		case event := <-registerCh:
			r.handleRegister(event.Payload.(*Client))
		case event := <-unRegisterCh:
			r.handleUnregister(event.Payload.(*Client))
		case event := <-broadcastCh:
			r.handleBroadcast(event.Payload.([]byte))
		}
	}
}

func (r *Room) handleRegister(client *Client) {
	r.mu.Lock()
	log.Printf("Registering client:  %v with pointer: %v", client.ID, &client)
	r.Clients[client] = &ClientInfo{
		Connected:   true,
		MediaTracks: make(map[string]*TrackInfo),
	}
	r.mu.Unlock()
	r.handleCreateOffer(client)
}

func dispatchKeyFrame(r *Room) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for client := range r.Clients {
		for _, receiver := range client.PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = client.PeerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}

func (r *Room) handleUnregister(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("Un-registering client:  %v with pointer: %v", client.ID, &client)
	if info, ok := r.Clients[client]; ok {
		for key, _ := range info.MediaTracks {
			removeTrackNoLock(r, client, key)
		}
		delete(r.Clients, client)
	}
	handleChatMessageNoLock(r, utils.Marshal(r.ToResponse()))
	close(client.Send)
}

func (r *Room) ToResponse() RoomInfoMessage {
	users := make([]UserInfo, 0, len(r.Clients))
	for client, info := range r.Clients {
		var avatar string
		if len(client.Avatar) > 0 {
			avatar = base64.StdEncoding.EncodeToString(client.Avatar)
		}

		tracks := make([]TrackInfo, 0, len(info.MediaTracks))
		for _, trackInfo := range info.MediaTracks {
			tracks = append(tracks, *trackInfo)
		}

		users = append(users, UserInfo{
			ID:              client.ID,
			Name:            client.Username,
			Avatar:          avatar,
			IsMicEnabled:    client.MicEnabled,
			IsVideoEnabled:  client.VideoEnabled,
			IsScreenEnabled: client.ScreenEnabled,
			Tracks:          tracks,
		})
	}

	return RoomInfoMessage{
		Type: "room-updated",
		Payload: RoomInfo{
			RoomID:   r.ID,
			RoomName: r.Name,
			Users:    users,
		},
	}
}

func (r *Room) handleBroadcast(payload []byte) {
	msg := Message{}
	if err := utils.Unmarshal(payload, &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	switch msg.Type {
	// TODO: handle tracks separately as we'll have to update client state on the hub
	// TODO: this is only used for toggleMicSharing
	case "webrtc-tracks":
		handleUserStateUpdate(r, msg)
	case "track-metadata":
		handleTrackMetadata(r, msg)
	case "chat-message":
		handleChatMessage(r, payload)
	case "webrtc-answer", "webrtc-ice-candidate", "webrtc-offer":
		handleWebRTCEvent(r, msg)
	default:
		log.Printf("Unhandled message type: %s", msg.Type)
	}
}

func (r *Room) handleCreateOffer(client *Client) {
	client.mu.Lock()
	defer client.mu.Unlock()
	clientID, err := strconv.Atoi(client.ID)
	if err != nil {
		log.Printf("Error parsing client id: %v", err)
		return
	}

	pc := client.PeerConnection
	if pc == nil {
		var err error
		pc, err = webrtc.NewPeerConnection(config)
		if err != nil {
			log.Printf("Failed to create PeerConnection: %v", err)
			return
		}

		_, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		})
		if err != nil {
			log.Printf("Failed to add transceiver: %v", err)
			return
		}

		_, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		})
		if err != nil {
			log.Printf("Failed to add transceiver: %v", err)
			return
		}

		_, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		})
		if err != nil {
			log.Printf("Failed to add transceiver: %v", err)
			return
		}

		client.PeerConnection = pc
		pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			if candidate != nil {
				iceCandidate := candidate.ToJSON()
				iceCandidateMessage := Message{
					Type:      "webrtc-ice-candidate",
					SenderID:  clientID,
					Candidate: &iceCandidate,
				}

				sendToClient(r, client, utils.Marshal(iceCandidateMessage))
			}
		})

		pc.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {

			switch p {
			case webrtc.PeerConnectionStateFailed:
				if err := pc.Close(); err != nil {
					log.Printf("Failed to close PeerConnection: %v", err)
				}
			case webrtc.PeerConnectionStateClosed:
				r.signalPeerConnections()
			default:
			}
		})

		pc.OnTrack(func(track *webrtc.TrackRemote, reciever *webrtc.RTPReceiver) {
			log.Printf("Received track of kind %s from client %d with id %s", track.Kind().String(), clientID, track.ID())

			r.mu.Lock()
			clientInfo, exists := r.Clients[client]
			if !exists {
				log.Printf("Client info not found for client %d", clientID)
				r.mu.Unlock()
				return
			}

			// Look up the existing TrackInfo by track ID
			trackInfo, trackExists := clientInfo.MediaTracks[track.StreamID()]
			if !trackExists {
				log.Printf("TrackInfo not found for track ID %s; and stream ID %s.", track.ID(), track.StreamID())
				r.mu.Unlock()
				return
			}
			r.mu.Unlock()

			trackLocal := addTrack(r, track)
			trackInfo.Track = trackLocal
			defer removeTrack(r, client, trackLocal)

			buf := make([]byte, 1500)
			rtpPkt := &rtp.Packet{}

			for {
				i, _, err := track.Read(buf)
				if err != nil {
					return
				}

				if err = rtpPkt.Unmarshal(buf[:i]); err != nil {
					log.Printf("Failed to unmarshal incoming RTP packet: %v", err)
					return
				}

				rtpPkt.Extension = false
				rtpPkt.Extensions = nil

				if err = trackLocal.WriteRTP(rtpPkt); err != nil {
					return
				}
			}
		})

	}

	r.signalPeerConnections()
}

func (r *Room) signalPeerConnections() {
	r.mu.Lock()
	trackLocals.mu.Lock()
	defer func() {
		r.mu.Unlock()
		trackLocals.mu.Unlock()
		dispatchKeyFrame(r)
	}()

	attemptSync := func() (tryAgain bool) {
		for client, clientInfo := range r.Clients {
			log.Println("Creating Offer to: ", client.ID)
			if client.PeerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				client.mu.Lock()
				clientInfo.Connected = false
				clientInfo.MediaTracks = make(map[string]*TrackInfo)
				client.PeerConnection = nil
				client.mu.Unlock()
				return true
			}

			// Map to track which senders or receivers are already accounted for
			existingSenders := map[string]bool{}
			existingStreamIDs := map[string]bool{}

			// Handle existing senders
			for _, sender := range client.PeerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				streamID := sender.Track().StreamID()
				existingSenders[sender.Track().ID()] = true
				existingStreamIDs[streamID] = true

				// Check existence in trackLocals using StreamID as the key
				trackLocal, ok := trackLocals.tracks[streamID]
				if !ok {
					log.Println("Removing outdated track:", sender.Track().ID(), "StreamID:", streamID)
					if err := client.PeerConnection.RemoveTrack(sender); err != nil {
						log.Println("Error removing track:", err)
						return true
					}
					continue
				}

				// Ensure the StreamID matches
				if trackLocal.StreamID() != streamID {
					log.Println("Removing mismatched track:", sender.Track().ID(), "StreamID:", streamID)
					if err := client.PeerConnection.RemoveTrack(sender); err != nil {
						log.Println("Error removing track:", err)
						return true
					}
				}
			}

			// Handle existing receivers to avoid loopback
			for _, receiver := range client.PeerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}
				streamID := receiver.Track().StreamID()
				existingStreamIDs[streamID] = true
			}

			// Add missing tracks by checking StreamID and TrackID
			for _, trackInfo := range trackLocals.tracks {
				if _, ok := existingStreamIDs[trackInfo.StreamID()]; !ok {
					log.Println("Adding new track with StreamID:", trackInfo.StreamID())
					if _, err := client.PeerConnection.AddTrack(trackInfo); err != nil {
						log.Println("Error adding track:", err)
						return true
					}
				}
			}

			offer, err := client.PeerConnection.CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = client.PeerConnection.SetLocalDescription(offer); err != nil {
				return true
			}

			clientID, err := strconv.Atoi(client.ID)
			if err != nil {
				return true
			}
			offerMessage := Message{
				Type:     "webrtc-offer",
				SenderID: clientID,
				Offer:    &offer,
			}
			client.Send <- utils.Marshal(r.ToResponse())

			sendToClient(r, client, utils.Marshal(offerMessage))
		}

		return
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
			go func() {
				time.Sleep(time.Second * 3)
				r.signalPeerConnections()
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}
}

func (r *Room) handleWebRTCOffer(msg Message) {
	senderID := msg.SenderID
	sender := r.GetClientByID(strconv.Itoa(senderID))

	if sender == nil {
		log.Printf("Sender not found: sender=%d", senderID)
		return
	}

	sender.mu.Lock()
	defer sender.mu.Unlock()

	pc := sender.PeerConnection
	if pc == nil {
		log.Printf("PeerConnection does not exist for client %d. Cannot renegotiate.", senderID)
		return
	}

	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			iceCandidate := candidate.ToJSON()
			iceCandidateMessage := Message{
				Type:      "webrtc-ice-candidate",
				SenderID:  senderID,
				Candidate: &iceCandidate,
			}

			sendToClient(r, sender, utils.Marshal(iceCandidateMessage))
		}
	})

	pc.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := pc.Close(); err != nil {
				log.Printf("Failed to close PeerConnection: %v", err)
			}
		case webrtc.PeerConnectionStateClosed:
			r.signalPeerConnections()
		default:
		}
	})

	pc.OnTrack(func(track *webrtc.TrackRemote, reciever *webrtc.RTPReceiver) {
		var kind string
		if msg.IsVideoEnabled != nil {
			kind = "video"
			sender.VideoEnabled = *msg.IsVideoEnabled
			log.Printf("Updated Video Enabled for client %v: %v", sender.ID, *msg.IsVideoEnabled)
		}
		if msg.IsMicEnabled != nil {
			kind = "audio"
			sender.MicEnabled = *msg.IsMicEnabled
			log.Printf("Updated Mic Enabled for client %v: %v", sender.ID, *msg.IsMicEnabled)
		}
		if msg.IsScreenEnabled != nil {
			kind = "screen"
			sender.ScreenEnabled = *msg.IsScreenEnabled
			log.Printf("Updated Screen Enabled for client %v: %v", sender.ID, *msg.IsScreenEnabled)
		}
		log.Printf("Received track of kind %s from client %d with id %s", kind, senderID, track.ID())

		r.mu.Lock()
		clientInfo, exists := r.Clients[sender]
		if !exists {
			log.Printf("Client info not found for client %d", senderID)
			r.mu.Unlock()
			return
		}

		// Look up the existing TrackInfo by track ID
		trackInfo, trackExists := clientInfo.MediaTracks[track.StreamID()]
		if !trackExists {
			log.Printf("TrackInfo not found for track ID %s; and stream ID %s.", track.ID(), track.StreamID())
			r.mu.Unlock()
			return
		}
		r.mu.Unlock()

		trackLocal := addTrack(r, track)
		trackInfo.Track = trackLocal
		defer removeTrack(r, sender, trackLocal)

		buf := make([]byte, 1500)
		rtpPkt := &rtp.Packet{}

		for {
			i, _, err := track.Read(buf)
			if err != nil {
				return
			}

			if err = rtpPkt.Unmarshal(buf[:i]); err != nil {
				log.Printf("Failed to unmarshal incoming RTP packet: %v", err)
				return
			}

			rtpPkt.Extension = false
			rtpPkt.Extensions = nil

			if err = trackLocal.WriteRTP(rtpPkt); err != nil {
				return
			}
		}

	})

	if err := pc.SetRemoteDescription(*msg.Offer); err != nil {
		log.Printf("Failed to set remote description: %v", err)
		return
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		log.Printf("Failed to create answer: %v", err)
		return
	}

	if err = pc.SetLocalDescription(answer); err != nil {
		log.Printf("Failed to set local description: %v", err)
		return
	}

	answerMessage := Message{
		Type:     "webrtc-answer",
		SenderID: senderID,
		Answer:   &answer,
	}
	log.Println("Handled offer, sent answer to ", senderID)
	sendToClient(r, sender, utils.Marshal(answerMessage))
}

func (r *Room) handleWebRTCAnswer(msg Message) {
	senderID := msg.SenderID
	sender := r.GetClientByID(strconv.Itoa(senderID))

	if sender == nil {
		log.Printf("Sender not found: sender=%d", senderID)
		return
	}

	sender.mu.Lock()
	defer sender.mu.Unlock()

	pc := sender.PeerConnection
	if pc == nil {
		log.Printf("PeerConnection not found for sender=%d", senderID)
		return
	}

	if err := pc.SetRemoteDescription(*msg.Answer); err != nil {
		log.Printf("Failed to set remote description: %v", err)
		return
	}

	log.Printf("Successfully set answer remote description for client %d", senderID)
}

func addTrack(r *Room, t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	trackLocals.mu.Lock()
	defer func() {
		trackLocals.mu.Unlock()
		r.signalPeerConnections()
	}()

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		panic(err)
	}

	if trackLocals.tracks == nil {
		trackLocals.tracks = make(map[string]*webrtc.TrackLocalStaticRTP)
	}
	trackLocals.tracks[t.StreamID()] = trackLocal
	return trackLocal
}

func removeTrack(r *Room, client *Client, t *webrtc.TrackLocalStaticRTP) {
	r.mu.Lock()
	trackLocals.mu.Lock()
	info := r.Clients[client]
	defer func() {
		r.mu.Unlock()
		trackLocals.mu.Unlock()
		r.signalPeerConnections()
	}()

	if info != nil {
		_, exists := info.MediaTracks[t.StreamID()]
		if exists {
			delete(info.MediaTracks, t.StreamID())
		}
	}

	_, exists := trackLocals.tracks[t.StreamID()]
	if exists {
		log.Println("Removing track")
		delete(trackLocals.tracks, t.StreamID())
	}
}

func (r *Room) handleWebRTCIceCandidate(msg Message) {
	senderID := msg.SenderID
	sender := r.GetClientByID(strconv.Itoa(senderID))

	if sender == nil {
		log.Printf("Sender not found: sender=%d", senderID)
		return
	}

	if msg.Candidate == nil {
		log.Println("Received nil ICE candidate")
		return
	}

	sender.mu.Lock()
	if sender.PeerConnection == nil {
		log.Println("PeerConnection has not been initialized")
		sender.mu.Unlock()
		return
	}
	if err := sender.PeerConnection.AddICECandidate(*msg.Candidate); err != nil {
		log.Printf("Error adding ICE candidate: %v", err)
	}
	sender.mu.Unlock()
}

func handleTrackMetadata(r *Room, msg Message) {
	log.Println("Handling track metadata:", msg)

	// Find the client by sender ID
	client := r.GetClientByID(strconv.Itoa(msg.SenderID))
	if client == nil {
		log.Printf("Client not found for sender ID: %d", msg.SenderID)
		return
	}

	// Acquire lock to modify room state
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure the client exists in the room's state
	info, exists := r.Clients[client]
	if !exists {
		log.Printf("Client state not found for client ID: %v", client.ID)
		return
	}

	// Validate message fields
	if msg.TrackType == "" || msg.TrackID == "" || msg.StreamID == "" {
		log.Println("Invalid track metadata received; missing required fields")
		return
	}

	// Check if the client's MediaTracks map is initialized
	if info.MediaTracks == nil {
		info.MediaTracks = make(map[string]*TrackInfo)
	}

	// Remove any existing track with the same Kind
	for trackID, trackInfo := range info.MediaTracks {
		if trackInfo.Kind == msg.TrackType {
			// Delete the existing track
			delete(info.MediaTracks, trackID)
			log.Printf("Deleted old track with ID %s of kind %s for client %s", trackID, msg.TrackType, client.ID)
			break
		}
	}

	// Add the new track
	newTrack := &TrackInfo{
		Kind:     msg.TrackType,
		ID:       msg.TrackID,
		StreamID: msg.StreamID,
	}
	info.MediaTracks[msg.StreamID] = newTrack
	log.Printf("New track added for client %s: %+v", client.ID, *newTrack)

	// Notify all clients except the sender about the updated room state
	roomState := r.ToResponse()
	for otherClient := range r.Clients {
		if otherClient.ID == strconv.Itoa(msg.SenderID) {
			continue
		}
		otherClient.Send <- utils.Marshal(roomState)
	}
}

func handleUserStateUpdate(r *Room, msg Message) {
	log.Println("Handling user state updates: ", msg)
	// Find the client by sender ID
	client := r.GetClientByID(strconv.Itoa(msg.SenderID))
	if client == nil {
		log.Printf("Client not found for sender ID: %d", msg.SenderID)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure the client exists in the room's state
	_, exists := r.Clients[client]
	if !exists {
		log.Printf("Client state not found for client: %v", client.ID)
		return
	}

	// Update the client's media state
	if msg.IsMicEnabled != nil {
		client.MicEnabled = *msg.IsMicEnabled
		log.Printf("Updated MicEnabled for client %v: %v", client.ID, *msg.IsMicEnabled)
	}

	if msg.IsVideoEnabled != nil {
		client.VideoEnabled = *msg.IsVideoEnabled
		log.Printf("Updated VideoEnabled for client %v: %v", client.ID, *msg.IsVideoEnabled)
	}

	if msg.IsScreenEnabled != nil {
		client.ScreenEnabled = *msg.IsScreenEnabled
		log.Printf("Updated ScreenEnabled for client %v: %v", client.ID, *msg.IsScreenEnabled)
	}

	// Notify all clients except sender about the updated state
	roomState := r.ToResponse()
	for client := range r.Clients {
		if client.ID == strconv.Itoa(msg.SenderID) {
			continue
		}
		client.Send <- utils.Marshal(roomState)
	}
}

func removeTrackNoLock(r *Room, client *Client, trackID string) {
	info := r.Clients[client]
	trackInfo := info.MediaTracks[trackID]

	for otherClient := range r.Clients {
		pc := otherClient.PeerConnection
		if pc == nil {
			continue
		}

		for _, sender := range pc.GetSenders() {
			if sender.Track() == trackInfo.Track {
				if err := pc.RemoveTrack(sender); err != nil {
					log.Printf("Failed to remove track from PeerConnection for client %s: %v", otherClient.ID, err)
				}
				break
			}
		}
	}

	delete(info.MediaTracks, trackID)
	r.Clients[client] = info
}

func handleChatMessageNoLock(r *Room, msg []byte) {
	for client := range r.Clients {
		select {
		case client.Send <- msg:
		default:
			r.mu.RUnlock()
			r.mu.Lock()
			delete(r.Clients, client)
			close(client.Send)
			r.mu.Unlock()
			r.mu.RLock()
		}
	}
}

func handleChatMessage(r *Room, msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handleChatMessageNoLock(r, msg)
}

func handleWebRTCEvent(r *Room, message Message) {
	switch message.Type {
	case "webrtc-answer":
		r.handleWebRTCAnswer(message)
	case "webrtc-offer":
		r.handleWebRTCOffer(message)
	case "webrtc-ice-candidate":
		r.handleWebRTCIceCandidate(message)
	default:
		log.Printf("Unkown message type: %s", message.Type)
	}
}

func sendToClient(r *Room, client *Client, message []byte) {
	select {
	case client.Send <- message:
	default:
		r.mu.Lock()
		delete(r.Clients, client)
		close(client.Send)
		r.mu.Unlock()
	}
}
