package main

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/mixcode/broadlink"
	"github.com/ysmood/kit"
	g "github.com/ysmood/kit"
)

const dbRoot = "db"

type device struct {
	d broadlink.Device
}

type actionData struct {
	Actions []string
	Icon    string
	IRcode  []byte
}

func newDevice() (*device, error) {
	devs, err := broadlink.DiscoverDevices(100*time.Millisecond, 0)

	d := devs[0]

	myname := "home"         // Your local machine's name.
	myid := make([]byte, 15) // Must be 15 bytes long.
	// Fill myid[] with some unique ID for your local machine.

	d.Auth(myid, myname) // d.ID and d.AESKey will be updated on success.

	return &device{d}, err
}

func (dev *device) learn(name string, icon string) error {
	var ircode []byte

	// Enter capturing mode.
	dev.d.StartCaptureRemoteControlCode()

	// Point a remote controller toward the device and press a button to have some signal.

	// Poll captured data. (Certainly you can do much better than this ;p)
	for i := 0; i < 30; i++ {
		var err error
		_, ircode, err = dev.d.ReadCapturedRemoteControlCode()
		if err == nil {
			g.Log("learned:", name)
			return dev.save(name, &actionData{IRcode: ircode, Icon: icon})
		}
		if err != broadlink.ErrNotCaptured {
			return err // real error
		}
		time.Sleep(time.Second)
		continue
	}

	return errors.New("learn timeout")
}

func (dev *device) send(name string) error {
	var data actionData
	err := g.ReadJSON(dev.path(name), &data)
	if err != nil {
		return err
	}

	if data.IRcode != nil {
		return dev.sendAction(name)
	}

	for _, action := range data.Actions {
		dev.sendAction(action)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (dev *device) sendAction(name string) error {
	sleeper := kit.BackoffSleeper(50*time.Millisecond, 50*time.Millisecond, nil)
	return g.Retry(context.Background(), sleeper, func() (bool, error) {
		if !g.FileExists(dev.path(name)) {
			return false, nil
		}

		var data actionData
		err := g.ReadJSON(dev.path(name), &data)
		if err != nil || data.IRcode == nil {
			panic(err)
		}

		g.Log(dev.d.SendIRRemoteCode(data.IRcode, 1))
		return true, nil
	})
}

func (dev *device) path(name string) string {
	return filepath.Join(dbRoot, name+".json")
}

func (dev *device) addGroup(name string, icon string, actions []string) error {
	if len(actions) == 0 {
		return errors.New("group empty")
	}
	return dev.save(name, &actionData{Actions: actions, Icon: icon})
}

func (dev *device) save(name string, data *actionData) error {
	p := dev.path(name)

	if g.FileExists(p) {
		return errors.New("name already exists")
	}

	return g.OutputFile(p, data, nil)
}

func (dev *device) rename(from, to string) error {
	return g.Move(dev.path(from), dev.path(to), nil)
}

func (dev *device) delete(name string) error {
	return g.Remove(dev.path(name))
}

func (dev *device) list() (map[string]actionData, error) {
	ps := g.Walk(dev.path("*")).MustList()

	list := map[string]actionData{}
	for _, p := range ps {
		var data actionData
		err := g.ReadJSON(p, &data)
		if err != nil {
			return nil, err
		}

		b := filepath.Base(p)
		list[strings.TrimSuffix(b, filepath.Ext(b))] = data
	}

	return list, nil
}
