package types

import (
	"encoding/base64"
	"log"
	"strconv"
	"sync"
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
	defer r.mu.Unlock()

	log.Printf("Registering client:  %v with pointer: %v", client.ID, &client)
	r.Clients[client] = &ClientInfo{
		Connected:   true,
		MediaTracks: make(map[string]*TrackInfo),
	}
	r.handleCreateOffer(client)
	dispatchKeyFrame(r)
	handleChatMessageNoLock(r, utils.Marshal(r.ToResponse()))
}

func dispatchKeyFrame(r *Room) {
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
		client.mu.RLock()

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

		client.mu.RUnlock()
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
	case "webrtc-tracks":
		handleUserStateUpdate(r, msg)
	case "chat-message":
		handleChatMessage(r, payload)
	case "webrtc-answer", "webrtc-ice-candidate", "webrtc-offer":
		handleWebRTCEvent(r, msg)
	case "stop-share":
		handleStopShare(r, msg)
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

		_, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		})
		if err != nil {
			log.Printf("Failed to add transceiver: %v", err)
			return
		}

		// Accept one audio and one video track incoming
		// for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		// 	if _, err := pc.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
		// 		Direction: webrtc.RTPTransceiverDirectionRecvonly,
		// 	}); err != nil {
		// 		log.Printf("Failed to add transceiver: %v", err)
		// 		return
		// 	}
		// }

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

		pc.OnTrack(func(track *webrtc.TrackRemote, reciever *webrtc.RTPReceiver) {
			log.Printf("Received track of kind %s from client %d with id %s", track.Kind().String(), clientID, track.ID())

			localTrack, err := webrtc.NewTrackLocalStaticRTP(
				track.Codec().RTPCodecCapability,
				track.ID(),
				track.StreamID(),
			)
			if err != nil {
				log.Printf("Failed to create local track: %v", err)
				return
			}

			r.mu.Lock()
			clientInfo, exists := r.Clients[client]
			if exists {
				clientInfo.MediaTracks[track.ID()] = &TrackInfo{
					Track: localTrack,
					ID:    track.ID(),
					Kind:  track.Kind().String(),
				}
				r.Clients[client] = clientInfo
			}
			r.mu.Unlock()

			r.forwardTrackToOthers(client, localTrack, track)
		})

	}

	for otherClient, clientInfo := range r.Clients {
		if otherClient == client {
			continue
		}
		for _, trackInfo := range clientInfo.MediaTracks {
			_, err := pc.AddTrack(trackInfo.Track)
			if err != nil {
				log.Printf("Failed to add track for client %s: %v", otherClient.ID, err)
			}
		}
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Printf("Failed to create offer: %v", err)
		return
	}

	// modifiedSDP := strings.ReplaceAll(offer.SDP, "a=setup:actpass", "a=setup:passive")
	// offer.SDP = modifiedSDP

	if err = pc.SetLocalDescription(offer); err != nil {
		log.Printf("Failed to set local description: %v", err)
		return
	}

	offerMessage := Message{
		Type:     "webrtc-offer",
		SenderID: clientID,
		Offer:    &offer,
	}

	log.Println("Sending initial Offer to: ", clientID)
	sendToClient(r, client, utils.Marshal(offerMessage))
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

	r.mu.Lock()
	clientInfo, exists := r.Clients[sender]
	if exists {
		// Create a map to keep track of the current active tracks
		activeTracks := make(map[string]bool)
		for _, transceiver := range pc.GetTransceivers() {
			if transceiver.Receiver() != nil {
				track := transceiver.Receiver().Track()
				if track != nil {
					activeTracks[track.ID()] = true
				}
			}
		}

		// Check which tracks are no longer active and remove them
		for trackID := range clientInfo.MediaTracks {
			if !activeTracks[trackID] {
				log.Printf("Track %s removed for client %d", trackID, senderID)
				delete(clientInfo.MediaTracks, trackID)
			}
		}

		r.Clients[sender] = clientInfo
	}
	r.mu.Unlock()

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

		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			track.Codec().RTPCodecCapability,
			track.ID(),
			track.StreamID(),
		)
		if err != nil {
			log.Printf("Failed to create local track: %v", err)
			return
		}

		r.mu.Lock()
		clientInfo, exists := r.Clients[sender]
		if exists {
			clientInfo.MediaTracks[track.ID()] = &TrackInfo{
				Track: localTrack,
				ID:    track.ID(),
				Kind:  kind,
			}
			r.Clients[sender] = clientInfo
		}
		r.mu.Unlock()

		r.forwardTrackToOthers(sender, localTrack, track)
	})

	log.Printf("Peer Connection offer State: ", pc.SignalingState())
	if err := pc.SetRemoteDescription(*msg.Offer); err != nil {
		log.Printf("Failed to set remote description: %v", err)
		return
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		log.Printf("Failed to create answer: %v", err)
		return
	}

	// modifiedSDP := strings.ReplaceAll(answer.SDP, "a=setup:active", "a=setup:actpass")
	// answer.SDP = modifiedSDP

	if err = pc.SetLocalDescription(answer); err != nil {
		log.Printf("Failed to set local description: %v", err)
		return
	}

	answerMessage := Message{
		Type:     "webrtc-answer",
		SenderID: senderID,
		Answer:   &answer,
	}
	log.Printf("Handled offer, sending answer", pc.SignalingState())
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
	log.Printf("Peer Connection State: ", pc.SignalingState())
}

func (r *Room) forwardTrackToOthers(sender *Client, localTrack *webrtc.TrackLocalStaticRTP, remoteTrack *webrtc.TrackRemote) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for client := range r.Clients {
		if client == sender {
			continue
		}

		forwardPC := client.PeerConnection
		if forwardPC == nil {
			continue
		}

		log.Println("Forwarding track to: ", client.ID)
		_, err := forwardPC.AddTrack(localTrack)
		if err != nil {
			log.Printf("Failed to add track to PeerConnection for client %s: %v", client.ID, err)
		}

		handleRenegotiation(r, client)
		client.Send <- utils.Marshal(r.ToResponse())
		go handleTrackForwarding(remoteTrack, localTrack)
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
	log.Println("Added Ice candidate for client", senderID)
	log.Println("Peer Connection State: ", sender.PeerConnection.ConnectionState())
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

func handleRenegotiation(r *Room, client *Client) {
	clientID, err := strconv.Atoi(client.ID)
	if err != nil {
		log.Printf("Error parsing client id: %v", err)
		return
	}

	pc := client.PeerConnection
	if pc == nil {
		log.Printf("PeerConnection does not exist for client %d. Cannot renegotiate.", clientID)
		return
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Printf("Failed to create offer: %v", err)
		return
	}

	if err = pc.SetLocalDescription(offer); err != nil {
		log.Printf("Failed to set local description: %v", err)
		return
	}

	offerMessage := Message{
		Type:     "webrtc-renegotiation",
		SenderID: clientID,
		Offer:    &offer,
	}
	sendToClient(r, client, utils.Marshal(offerMessage))
}

func handleTrackForwarding(rt *webrtc.TrackRemote, lt *webrtc.TrackLocalStaticRTP) {
	buf := make([]byte, 1500)
	rtpPkt := &rtp.Packet{}

	for {
		n, _, readErr := rt.Read(buf)
		if readErr != nil {
			log.Printf("Error reading RTP packets from track: %v", readErr)
			return
		}

		if unMarshalErr := rtpPkt.Unmarshal(buf[:n]); unMarshalErr != nil {
			log.Printf("Failed to unmarshal incoming RTP packet: %v", unMarshalErr)
			return
		}

		if writeErr := lt.WriteRTP(rtpPkt); writeErr != nil {
			log.Printf("Error writting RTP packets to local track: %v", writeErr)
			return
		}
	}
}

func handleStopShare(r *Room, msg Message) {
	senderID := msg.SenderID
	trackID := msg.TrackID

	if trackID == "" {
		log.Printf("TrackID missing in stop screen share request from sender %d", senderID)
		return
	}

	sender := r.GetClientByID(strconv.Itoa(senderID))
	if sender == nil {
		log.Printf("Sender not found: sender=%d", senderID)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	sender.mu.Lock()
	defer sender.mu.Unlock()

	clientInfo, exists := r.Clients[sender]
	if !exists {
		log.Printf("Client info not found for sender %d", senderID)
		return
	}

	trackInfo, ok := clientInfo.MediaTracks[trackID]
	if !ok || trackInfo == nil {
		log.Printf("No screen-sharing track found for sender %d with track ID %s", senderID, trackID)
		return
	}

	log.Printf("Stopping track for sender %d, track ID %s", senderID, trackID)
	removeTrackNoLock(r, sender, trackID)
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
