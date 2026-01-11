package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
)

type Handler func(url *url.URL) error

type viaServer struct {
	commands []*viaCommand
}

func (vs *viaServer) AddCommand(verb string, handler Handler) *viaCommand {
	vc := &viaCommand{verb: verb, handler: handler}
	vs.commands = append(vs.commands, vc)
	return vc
}

type viaRequest struct {
	command    *viaCommand
	parameters []*viaParameter
	query      map[string][]string
	err        error
}

func (vr *viaRequest) ToUrl() *url.URL {
	return new(url.URL)
}

type viaParseFunc func(req *viaRequest, args ...string) viaParseFunc

func (vs *viaServer) Serve(args ...string) error {

	// TODO implement default command handling

	if len(args) == 0 {
		return errors.New("missing args")
	}

	req := new(viaRequest)
	req.query = make(map[string][]string)

	for _, cmd := range vs.commands {
		if cmd.verb == args[0] {
			req.command = cmd
			if u, err := cmd.parse(req, args[1:]...); err == nil {
				return cmd.handler(u)
			} else {
				return err
			}
		}
	}

	return errors.New("not a valid command verb: " + args[0])
}

type viaCommand struct {
	verb       string
	handler    Handler
	parameters []*viaParameter
}

func (vc *viaCommand) parse(req *viaRequest, args ...string) (*url.URL, error) {

	for f := vc.defaultValuesOrParameter(req, args...); f != nil; f = f(req, args...) {
	}

	return req.ToUrl(), nil
}

func (vc *viaCommand) defaultValuesOrParameter(req *viaRequest, args ...string) viaParseFunc {
	if len(args) == 0 {
		// TODO: check if command actually expects any
		req.err = errors.New("no args for default values or parameter")
		return nil
	}

	switch strings.HasPrefix(args[0], "-") {
	case false:
		for _, p := range vc.parameters {
			if slices.Contains(p.opts, OptDefault) {
				req.parameters = append(req.parameters, p)
				if value, err := p.checkValue(args[0]); err == nil {
					req.query[p.title] = []string{value}
					return vc.valueOrParameter(req, args[1:]...)
				} else {
					req.err = err
					return nil
				}
			}
		}
	case true:
		maybeParamTitle := strings.TrimPrefix(args[0], "-")
		for _, p := range vc.parameters {
			if p.title == maybeParamTitle {
				req.parameters = append(req.parameters, p)
				req.query[p.title] = nil
				return vc.valueOrParameter(req, args[1:]...)
			}
		}
		req.err = errors.New("parameter " + args[0] + " not found")
		return nil
	}

	return nil
}

func (vc *viaCommand) valueOrParameter(req *viaRequest, args ...string) viaParseFunc {
	if len(args) == 0 {
		return nil
	}

	switch strings.HasPrefix(args[0], "-") {
	case false:
		if len(req.parameters) > 0 {
			lastParameter := req.parameters[len(req.parameters)-1]
			if value, err := lastParameter.checkValue(args[0]); err == nil {
				req.query[lastParameter.title] = append(req.query[lastParameter.title], value)
				return vc.valueOrParameter(req, args[1:]...)
			} else {
				req.err = err
				return nil
			}
		} else {
			req.err = errors.New("no parameter to set value for")
			return nil
		}
	case true:
		maybeParamTitle := strings.TrimPrefix(args[0], "-")
		for _, p := range vc.parameters {
			if p.title == maybeParamTitle {
				req.parameters = append(req.parameters, p)
				req.query[p.title] = nil
				return vc.valueOrParameter(req, args[1:]...)
			}
		}
		req.err = errors.New("parameter " + maybeParamTitle + " not found")
		return nil
	}

	return nil
}

type viaParameter struct {
	title        string
	opts         []viaOption
	values       []string
	defaultValue string
}

func (vp *viaParameter) checkValue(maybeValue string) (string, error) {
	if slices.Contains(vp.opts, OptBoolean) {
		return "", errors.New("value is a flag")
	}

	if maybeValue == "" {
		return "", errors.New("check value not specified")
	}

	if len(vp.values) == 0 {
		return maybeValue, nil
	}

	lcMaybeValue := strings.ToLower(maybeValue)

	for _, v := range vp.values {
		lcv := strings.ToLower(v)
		if strings.HasPrefix(lcv, lcMaybeValue) {
			return v, nil
		}
	}

	return "", errors.New("value " + maybeValue + " is not valid for " + vp.title)
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

func initViaServer() *viaServer {
	vs := new(viaServer)

	vs.AddCommand("backup", echoHandler)
	cleanup := vs.AddCommand("cleanup", echoHandler)

	cleanup.AddParameter("id", OptDefault, OptMultipleValues)
	cleanup.AddParameter("slug", OptMultipleValues)
	cleanup.AddParameter("os", OptMultipleValues, OptEnvVar).SetValues("Windows", "macOS", "Linux")
	cleanup.AddParameter("lang-code", OptMultipleValues, OptEnvVar).SetValues("en", "ru")
	cleanup.AddParameter("download-types", OptMultipleValues, OptEnvVar).SetValues("installer", "dlc", "extra")
	cleanup.AddParameter("no-patches", OptEnvVar, OptBoolean)
	cleanup.AddParameter("downloads-layout", OptEnvVar).SetValues("flat", "sharded").SetDefaultValue("flat")
	cleanup.AddParameter("all", OptBoolean)
	cleanup.AddParameter("test", OptBoolean)

	return vs
}

func main() {

	vs := initViaServer()

	if err := vs.Serve(os.Args[1:]...); err != nil {
		panic(err)
	}

}

func echoHandler(u *url.URL) error {
	fmt.Println(u)
	return nil
}
