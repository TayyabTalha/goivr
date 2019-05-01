package main

import (
	"context"
	"time"

	"github.com/CyCoreSystems/ari"
	"github.com/CyCoreSystems/ari-proxy/client"
	"github.com/CyCoreSystems/ari/ext/audiouri"
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

	// i.GetChanVar("EXTEN")
	// err := i.Dial("SIP/12345", 30*time.Second)
	// if err != nil {
	// 	log.Error("Failed to dial", "Error", err)
	// }
	// soundfile := "sound:demo-congrats"
	// i.Play(ctx, soundfile)
	// req := ari.OriginateRequest{
	// 	Endpoint: "SIP/1002",
	// 	Timeout:  10,
	// 	CallerID: "Exten <1001>",
	// 	App:      "ARI-APP",
	// }
	// i.channelHandle.Originate(req)
	// t, _ := time.ParseDuration("10s")
	// i.Wait(ctx, t)
	// i.PlayDigits(ctx, "1256*", "")
	i.PlayNumber(ctx, 45678)
	return

}

// PlayRec Play Recoding
func (i *IVR) PlayRec(ctx context.Context, rec string) error {
	log.Debug("Playing ", "Recoding", rec)
	uri := audiouri.RecordingURI(rec)
	if err := i.Play(ctx, uri); err != nil {
		return err
	}
	return nil
}

// PlayNumber Play number
func (i *IVR) PlayNumber(ctx context.Context, num int) error {
	log.Debug("Playing ", "Number", num)
	uri := audiouri.NumberURI(num)
	if err := i.Play(ctx, uri); err != nil {
		return err
	}
	return nil
}

// Wait for given duration
func (i *IVR) Wait(ctx context.Context, t time.Duration) error {
	log.Debug("Waiting ", "Duration", t)
	uri := audiouri.WaitURI(t)
	for _, sound := range uri {
		if err := i.Play(ctx, sound); err != nil {
			return err
		}
	}
	return nil
}

// PlayDigits Play Digits
func (i *IVR) PlayDigits(ctx context.Context, d string, h string) error {
	log.Debug("Playing ", "Digits", d, "Hash", h)
	uri := audiouri.DigitsURI(d, h)
	for _, sound := range uri {
		if err := i.Play(ctx, sound); err != nil {
			return err
		}
	}
	return nil
}

// PlayDuration Play Duration 
func (i *IVR) PlayDuration(ctx context.Context, t time.Duration) error {
	log.Debug("Playing ", "Duration", t)
	d := audiouri.DurationURI(t)
	for _, sound := range d {
		if err := i.Play(ctx, sound); err != nil {
			return err
		}
	}
	return nil
}

// PlayDateTime Play Date TIME 
func (i *IVR) PlayDateTime(ctx context.Context, t time.Time) error {
	log.Debug("Playing ", "Time", t)
	d := audiouri.DateTimeURI(t)
	for _, sound := range d {
		if err := i.Play(ctx, sound); err != nil {
			return err
		}
	}
	return nil
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
	log.Debug("Dialing ", "number", number, "timeout", time)
	if err := i.channelHandle.Dial(number, time); err != nil {
		return err
	}
	return nil
}
