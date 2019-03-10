package main

import (
	"context"
	"time"

	"github.com/CyCoreSystems/ari"
	"github.com/CyCoreSystems/ari-proxy/client"
	"github.com/CyCoreSystems/ari/ext/play"
	"github.com/inconshreveable/log15"
)

var log = log15.New("APP", "GO-IVR")

// IVR TYPE
type IVR struct {
	channelHandle *ari.ChannelHandle
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cl, err := connect(ctx, "ivr")
	if err != nil {
		log.Error("Failed to build ARI client", "error", err)
		return
	}

	// setup app

	log.Info("Listening for new calls")
	sub := cl.Bus().Subscribe(nil, "StasisStart")

	for {
		select {
		case e := <-sub.Events():
			v := e.(*ari.StasisStart)
			log.Debug("Got stasis start", "channel", v.Channel.ID)
			go func() {
				ivr := &IVR{
					channelHandle: cl.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)),
				}
				ivr.Start(ctx)
			}()
		case <-ctx.Done():
			return
		}
	}
}

func connect(ctx context.Context, appName string) (ari.Client, error) {
	c, err := client.New(ctx,
		client.WithApplication(appName),
		client.WithURI("nats://localhost:4222"),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Start IVR
func (i *IVR) Start(ctx context.Context) {
	h := i.channelHandle
	defer h.Hangup()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Debug("Running app", "channel", h.ID())

	end := h.Subscribe(ari.Events.StasisEnd)
	defer end.Cancel()

	// End the app when the channel goes away
	go func() {
		<-end.Events()
		cancel()
	}()

	i.GetChanVar("EXTEN")
	i.Dial("12345", 30*time.Second)
	soundfile := "sound:demo-congrats"
	i.Play(ctx, soundfile)
	return

}

// Play Sound file
func (i *IVR) Play(ctx context.Context, f string) error {
	log.Debug("Playing ", "Sound", f)
	if err := play.Play(ctx, i.channelHandle, play.URI(f)).Err(); err != nil {
		return err
	}
	return nil
}

// Answer Call
func (i *IVR) Answer() error {
	log.Debug("Answering Call ", "channel", i.channelHandle.ID())
	if err := i.channelHandle.Answer(); err != nil {
		return err
	}
	return nil
}

// HangUp Call
func (i *IVR) HangUp() error {
	log.Debug("Hainging UP Channel ", "channel", i.channelHandle.ID())
	if err := i.channelHandle.Hangup(); err != nil {
		return err
	}
	return nil
}

// GetChanVar Event
func (i *IVR) GetChanVar(name string) (string, error) {
	value, err := i.channelHandle.GetVariable(name)
	if err != nil {
		return "", err
	}

	log.Debug("Get Variable ", name, value)
	return value, nil
}

// Dial Number
func (i *IVR) Dial(number string, time time.Duration) error {
	if err := i.channelHandle.Dial(number, time); err != nil {
		return err
	}
	return nil
}
