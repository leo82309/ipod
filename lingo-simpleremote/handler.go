package simpleremote

import (
	"log"
	"sync"

	"github.com/leo82309/ipod"
	"github.com/leo82309/ipod/mpd"
)

type DeviceSimpleRemote interface {
}

var (
	mpdClient *mpd.Client
	mpdMutex  sync.Mutex
)

func getMpdClient() (*mpd.Client, error) {
	mpdMutex.Lock()
	defer mpdMutex.Unlock()
	if mpdClient == nil {
		client, err := mpd.NewClient("127.0.0.1:6600")
		if err != nil {
			return nil, err
		}
		mpdClient = client
	}
	return mpdClient, nil
}

func HandleSimpleRemote(req *ipod.Command, tr ipod.CommandWriter, dev DeviceSimpleRemote) error {
	switch msg := req.Payload.(type) {
	case *ContextButtonStatus:
		log.Printf("SimpleRemote: received %s", msg.State.String())
		client, err := getMpdClient()
		if err != nil {
			log.Panic("could not get mpd client")
			return err
		}

		switch {
		case msg.State&ContextButtonMask(ContextButtonPlayPause) != 0:
			client.Pause(mpd.CurrentStatus.State == "play")
		case msg.State&ContextButtonMask(ContextButtonNextTrack) != 0:
			client.Next()
		case msg.State&ContextButtonMask(ContextButtonPreviousTrack) != 0:
			client.Previous()
		}
	default:
		_ = msg
	}
	return nil
}
