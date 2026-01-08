package main

import (
	"fmt"
	"net/url"
)

type Handler func(url *url.URL) error

type viaDispatch struct {
	commands []*viaCommand
}

func (vd *viaDispatch) AddCommand(verb string, handler Handler) *viaCommand {
	vc := &viaCommand{verb: verb}
	vd.commands = append(vd.commands, vc)
	return vc
}

type viaCommand struct {
	verb       string
	handler    Handler
	parameters []*viaParameter
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

	vd.AddCommand("backup", nil)
	cleanup := vd.AddCommand("cleanup", nil)

	cleanup.AddParameter("id", OptDefault, OptMultipleValues)
	cleanup.AddParameter("slug", OptMultipleValues)
	cleanup.AddParameter("os", OptMultipleValues, OptEnvVar).SetValues("Windows", "macOS", "Linux")
	cleanup.AddParameter("lang-code", OptMultipleValues, OptEnvVar).SetValues("en", "ru")
	cleanup.AddParameter("download-types", OptMultipleValues, OptEnvVar).SetValues("installer", "dlc", "extra")
	cleanup.AddParameter("no-patches", OptEnvVar, OptBoolean)
	cleanup.AddParameter("downloads-layout", OptEnvVar).SetValues("flat", "sharded").SetDefaultValue("flat")
	cleanup.AddParameter("all", OptBoolean)
	cleanup.AddParameter("test", OptBoolean)

	fmt.Println(vd)

}
