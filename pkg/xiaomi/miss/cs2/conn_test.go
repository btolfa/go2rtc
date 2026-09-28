package cs2

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

// The cs2 command channel is big-endian on the wire (see marshalCmd), so
// ReadCommand must decode it the same way. If it reads little-endian the
// command constants used by the miss client never match and every response
// (auth, motor, ...) is silently dropped.
func TestReadCommandBigEndian(t *testing.T) {
	conn := &Conn{
		channels: [4]*dataChannel{newDataChannel(0, 1), nil, nil, nil},
	}

	tests := []struct {
		name string
		cmd  uint32
		data []byte
	}{
		{name: "encoded", cmd: 0x1001, data: []byte(`{"ret":0,"angle":46,"elevation":50}`)},
		{name: "motor", cmd: 0x113, data: []byte(`{"angle":1,"elevation":1,"ret":0}`)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw := binary.BigEndian.AppendUint32(nil, test.cmd)
			raw = append(raw, test.data...)

			// dataChannel.Push expects a 4-byte big-endian size prefix.
			frame := binary.BigEndian.AppendUint32(nil, uint32(len(raw)))
			frame = append(frame, raw...)
			require.NoError(t, conn.channels[0].Push(frame))

			cmd, data, err := conn.ReadCommand()
			require.NoError(t, err)
			require.Equal(t, test.cmd, cmd)
			require.Equal(t, test.data, data)
		})
	}
}
