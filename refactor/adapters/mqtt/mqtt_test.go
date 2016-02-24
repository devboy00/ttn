// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"reflect"
	"testing"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	//"github.com/TheThingsNetwork/ttn/utils/pointer"
)

func TestMQTTSend(t *testing.T) {
	tests := []struct {
		Desc      string          // Test Description
		Packet    []byte          // Handy representation of the packet to send
		Recipient []testRecipient // List of recipient to send

		WantData     []byte  // Expected Data on the recipient
		WantResponse []byte  // Expected Response from the Send method
		WantError    *string // Expected error nature returned by the Send method
	}{}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		// Generate new adapter
		// Generate reception servers

		// Operate
		// Send data to recipients
		// Retrieve data from servers

		// Check
		// Check if data has been received
		// Check if response is valid
		// Check if error is valid

		// Clean
		// Disconnect clients
	}
}

// ----- TYPE utilities
type testRecipient struct {
	Response  []byte
	TopicUp   string
	TopicDown string
}

type testPacket struct {
	payload []byte
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p testPacket) MarshalBinary() ([]byte, error) {
	if p.payload == nil {
		return nil, errors.New(ErrInvalidStructure, "Fake error")
	}

	return p.payload, nil
}

// String implements the core.Packet interface
func (p testPacket) String() string {
	return string(p.payload)
}

// ----- BUILD utilities
func createAdapter(t *testing.T) (*MQTT.Client, core.Adapter) {
	client, err := NewClient("testClient", "0.0.0.0", Tcp)
	if err != nil {
		panic(err)
	}

	adapter := NewAdapter(client, GetLogger(t, "adapter"))
	return client, adapter
}

func createServers(recipients []testRecipient) (*MQTT.Client, chan []byte) {
	client, err := NewClient("FakeServerClient", "0.0.0.0", Tcp)
	if err != nil {
		panic(err)
	}

	chresp := make(chan []byte)
	for _, r := range recipients {
		token := client.Subscribe(r.TopicUp, 2, func(client *MQTT.Client, msg MQTT.Message) {
			chresp <- msg.Payload()
		})
		if token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	return client, chresp
}

// ----- OPERATE utilities
func trySend(adapter core.Adapter, packet []byte, recipients []testRecipient) ([]byte, error) {
	// Convert testRecipient to core.Recipient using the mqtt recipient
	var coreRecipients []core.Recipient
	for _, r := range recipients {
		coreRecipients = append(coreRecipients, NewRecipient(r.TopicUp, r.TopicDown))
	}

	// Try send the packet
	chresp := make(chan struct {
		Data  []byte
		Error error
	})
	go func() {
		data, err := adapter.Send(testPacket{packet}, coreRecipients...)
		chresp <- struct {
			Data  []byte
			Error error
		}{data, err}
	}()

	select {
	case resp := <-chresp:
		return resp.Data, resp.Error
	case <-time.After(time.Millisecond * 200):
		return nil, nil
	}
}

// ----- CHECK utilities
func checkErrors(t *testing.T, want *string, got error) {
	if got == nil {
		if want == nil {
			Ok(t, "Check errors")
			return
		}
		Ko(t, "Expected error to be {%s} but got nothing", *want)
		return
	}

	if want == nil {
		Ko(t, "Expected no error but got {%v}", got)
		return
	}

	if got.(errors.Failure).Nature == *want {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%s} but got {%v}", *want, got)
}

func checkResponses(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check responses")
		return
	}
	Ko(t, "Received response does not match expectations.\nWant: %s\nGot:  %s", string(want), string(got))
}

func checkData(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check data")
		return
	}
	Ko(t, "Received data does not match expectations.\nWant: %s\nGot:  %s", string(want), string(got))
}
