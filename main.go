package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
)

type Handler func(url *url.URL) error

type viaDispatch struct {
	commands []*viaCommand
}

func (vd *viaDispatch) AddCommand(verb string, handler Handler) *viaCommand {
	vc := &viaCommand{verb: verb, handler: handler}
	vd.commands = append(vd.commands, vc)
	return vc
}

func (vd *viaDispatch) Serve(args ...string) error {

	// TODO implement default command handling

	if len(args) == 0 {
		return errors.New("missing args")
	}

	cmdArg := args[0]

	for _, cmd := range vd.commands {
		if cmd.verb == cmdArg {
			if u, err := cmd.parse(args...); err == nil {
				return cmd.handler(u)
			} else {
				return err
			}
		}
	}

	return errors.New("not a valid command verb: " + cmdArg)
}

type viaCommand struct {
	verb       string
	handler    Handler
	parameters []*viaParameter
}

func (vc *viaCommand) parse(args ...string) (*url.URL, error) {
	return new(url.URL), nil
}

type viaParameter struct {
	title        string
	opts         []viaOption
	values       []string
	defaultValue string
}

type viaOption = int

const (
	OptDefault viaOption = iota
	OptEnvVar
	OptMultipleValues
	OptRequired
	OptBoolean
)

func (vc *viaCommand) AddParameter(title string, opts ...viaOption) *viaParameter {

	// TODO: check existing default parameters in that command and other consistency checks

	vp := &viaParameter{
		title: title,
		opts:  opts,
	}

	vc.parameters = append(vc.parameters, vp)

	return vp
}

func (vp *viaParameter) SetValues(values ...string) *viaParameter {
	vp.values = values
	return vp
}

func (vp *viaParameter) SetDefaultValue(value string) *viaParameter {
	vp.defaultValue = value
	return vp
}

func main() {

	vd := new(viaDispatch)

	vd.AddCommand("backup", echoHandler)
	cleanup := vd.AddCommand("cleanup", echoHandler)

	cleanup.AddParameter("id", OptDefault, OptMultipleValues)
	cleanup.AddParameter("slug", OptMultipleValues)
	cleanup.AddParameter("os", OptMultipleValues, OptEnvVar).SetValues("Windows", "macOS", "Linux")
	cleanup.AddParameter("lang-code", OptMultipleValues, OptEnvVar).SetValues("en", "ru")
	cleanup.AddParameter("download-types", OptMultipleValues, OptEnvVar).SetValues("installer", "dlc", "extra")
	cleanup.AddParameter("no-patches", OptEnvVar, OptBoolean)
	cleanup.AddParameter("downloads-layout", OptEnvVar).SetValues("flat", "sharded").SetDefaultValue("flat")
	cleanup.AddParameter("all", OptBoolean)
	cleanup.AddParameter("test", OptBoolean)

	if err := vd.Serve(os.Args[1:]...); err != nil {
		panic(err)
	}

}

func echoHandler(u *url.URL) error {
	fmt.Println(u)
	return nil
}
