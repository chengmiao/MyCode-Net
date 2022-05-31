package tcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"gnet"
	"io"
	"net"
	"time"

	log "github.com/rs/zerolog/log"
)

const (
	// ctData denotes a binary data message.
	ctData = 0

	// ctPing denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	ctPing = 1

	// ctPong denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	ctPong = 2

	// ctClose denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the CloseMessage
	// function to format a close message payload.
	ctClose = 3
)

const (
	// like websocket
	closeStatusNormal = 1000
	// closeStatusGoingAway         = 1001 //
	// closeStatusProtocolError     = 1002
	// closeStatusUnsupportedData   = 1003
	// closeStatusFrameTooLarge     = 1004
	// closeStatusNoStatusRcvd      = 1005
	// closeStatusAbnormalClosure   = 1006
	// closeStatusBadMessageData    = 1007
	// closeStatusPolicyViolation   = 1008
	// closeStatusTooBigData        = 1009
	// closeStatusExtensionMismatch = 1010

	maxControlFramePayloadLength = 125
)

var (
	// ErrorPacketOverflow mean
	ErrorPacketOverflow = errors.New("packet too long")
	// ErrorEncryptionWord mean
	ErrorEncryptionWord = errors.New("Encryption word is wrong")
)

/* tcp package struct
0                   1                   2                   3
0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+---------------+---------------+---------------+---+-+-+-+-+-+-+
|                                               | C |           |
|                                               | T |ENCRYPTION |
|     Payload Data length(LittleEndian)         | R |   WORD    |
|                                               | L |           |
|                (24 bit)                       |(2)|    (6)    |
+---------------+---------------+---------------+-+-+-+-+-+-+-+-+
:                     Payload Data continued ...                :
+---------------------------------------------------------------+
*(CTRL bit):
00 : data
01 : ping
10 : pong
11 : close

*(describe encryption scheme):
encryption word: 0 ~ 31


type ServerMsgHeader struct {
	cmd       uint32 // cm_connected, cm_disconnected, cm_message, sm_message, sm_kickout, sm_forward, sm_broadcast, im_forward ...
	SessionID uint64 // sessionid of client
	data      []byte // MessageHeader struct
}

// client->gate
type MessageHeader struct {
	Proto     string   `protobuf:"bytes,1,opt,name=proto,proto3" json:"proto,omitempty"`
	Data      []byte   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

*/

// Adaption implement adaptor
type Adaption struct {
	RawConn        net.Conn
	EncryptionWord byte
}

// ReadMessage implement read method
func (m *Adaption) ReadMessage(sessionID int64) ([]byte, error) {
Loop:
	err := m.RawConn.SetReadDeadline(time.Now().Add(gnet.HeartbeatTime))
	if err != nil {
		return nil, err
	}

	head := make([]byte, 4)
	if _, err = io.ReadFull(m.RawConn, head); err != nil {
		return nil, err
	}

	n := binary.LittleEndian.Uint32(head)
	datalen := int(n & 0xFFFFFF)
	encryWord := byte((n >> 24) & 0x3F)
	ctrl := byte(n>>30) & 3

	if encryWord != m.EncryptionWord {
		return nil, ErrorEncryptionWord
	}
	data := make([]byte, datalen) // *不重用数据区*
	if datalen != 0 {
		_, err = io.ReadFull(m.RawConn, data)
		if err != nil {
			return nil, err
		}
	}

	switch ctrl {
	case ctData:
		return data, err

	case ctPing:
		// log.Debug().Int64("SessionID", sessionID).Msg("PING")
		if err = m.doWrite(ctPong, data); err != nil {
			return nil, err
		}
		fallthrough
	case ctPong:
		goto Loop
	default: // ctClose ...
		errcode := int(binary.LittleEndian.Uint16(data))
		return nil, fmt.Errorf("remote closed: %d", errcode)
	}
}

func (m *Adaption) doWrite(ctrl byte, data []byte) (err error) {
	err = m.RawConn.SetWriteDeadline(time.Now().Add(gnet.HeartbeatTime))
	if err == nil {
		wCache := new(bytes.Buffer)
		datalen := uint32(len(data))
		if datalen > 0xFFFFFF {
			return ErrorPacketOverflow
		}

		datalen |= uint32(ctrl<<6|m.EncryptionWord) << 24

		binary.Write(wCache, binary.LittleEndian, datalen)

		_, err = wCache.Write(data)
		if err == nil {
			_, err = wCache.WriteTo(m.RawConn)
		}
		return
	}
	return
}

// WriteMessage implement send message to remote
func (m *Adaption) WriteMessage(data []byte) error {
	return m.doWrite(ctData, data)
}

// Ping implement send ping
func (m *Adaption) Ping(msg []byte) error {
	return m.doWrite(ctPing, msg) // msg[0:maxControlFramePayloadLength])

}

// LocalAddr implement returns the local network address
func (m *Adaption) LocalAddr() net.Addr {
	return m.RawConn.LocalAddr()
}

// Close implement Adaption
func (m *Adaption) Close() error {
	p := make([]byte, 2)
	binary.LittleEndian.PutUint16(p, uint16(closeStatusNormal))
	err := m.doWrite(ctClose, p)
	if err != nil {
		return err
	}
	return m.RawConn.Close()
}

// RemoteAddr implement returns the remote network address
func (m *Adaption) RemoteAddr() net.Addr {
	return m.RawConn.RemoteAddr()
}

// NewAdapter create
func NewAdapter(conn net.Conn, encryword byte) *Adaption {
	if encryword > 0x3F {
		log.Error().Msg("encryption word over load")
		encryword &= 0x3F
	}

	m := &Adaption{
		RawConn:        conn,
		EncryptionWord: encryword,
	}
	return m
}
