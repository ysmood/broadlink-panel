package main

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/mixcode/broadlink"
	g "github.com/ysmood/gokit"
)

const dbRoot = "db"

type device struct {
	d broadlink.Device
}

type actionData struct {
	Actions []string
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

func (dev *device) learn(name string) error {
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
			return dev.save(name, &actionData{IRcode: ircode})
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
	}

	return nil
}

func (dev *device) sendAction(name string) error {
	errs := g.Retry(20, 50*time.Millisecond, func() {
		if !g.FileExists(dev.path(name)) {
			return
		}

		var data actionData
		err := g.ReadJSON(dev.path(name), &data)
		if err != nil || data.IRcode == nil {
			panic(err)
		}

		g.E(dev.d.SendIRRemoteCode(data.IRcode, 1))
	})
	if errs != nil {
		return errs[0].(error)
	}
	return nil
}

func (dev *device) path(name string) string {
	return filepath.Join(dbRoot, name+".json")
}

func (dev *device) addGroup(name string, actions []string) error {
	return dev.save(name, &actionData{Actions: actions})
}

func (dev *device) save(name string, data *actionData) error {
	p := dev.path(name)

	if g.FileExists(p) {
		return errors.New("name already exists")
	}

	return g.OutputFile(p, data, nil)
}

func (dev *device) delete(name string) error {
	return g.Remove(dev.path(name))
}

func (dev *device) list() (map[string]actionData, error) {
	ps, err := g.Glob([]string{dev.path("*")}, nil)
	if err != nil {
		return nil, err
	}

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
