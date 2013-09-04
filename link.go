package main

import (
	"github.com/daisy-consortium/pipeline-clientlib-go"
)

//Convinience for testing
type PipelineApi interface {
	Alive() (alive pipeline.Alive, err error)
	Scripts() (scripts pipeline.Scripts, err error)
	Script(id string) (script pipeline.Script, err error)
}

//Maintains some information about the pipeline client
type PipelineLink struct {
	pipeline       PipelineApi //Allows access to the pipeline fwk
	Version        string      //Framework version
	Authentication bool        //Framework authentication
	Mode           string      //Framework mode
}

func NewLink() (pLink *PipelineLink, err error) {
	pLink = &PipelineLink{
		pipeline: *pipeline.NewPipeline("http://localhost:8181/ws/"),
	}
	//assure that the pipeline is up
	err = bringUp(pLink)
	if err != nil {
		return nil, err
	}
	return
}

//checks if the pipeline is up
//otherwise it brings it up and fills the
//link object
func bringUp(pLink *PipelineLink) error {
	alive, err := pLink.pipeline.Alive()
	if err != nil {
		return err
	}
	pLink.Version = alive.Version
	pLink.Mode = alive.Mode
	pLink.Authentication = alive.Authentication
	return nil
}

//ScriptList returns the list of scripts available in the framework
func (p PipelineLink) Scripts() (scripts []pipeline.Script, err error) {
	scriptsStruct, err := p.pipeline.Scripts()
	if err != nil {
		return
	}
	scripts = make([]pipeline.Script, len(scriptsStruct.Scripts))
	//fill the script list with the complete definition
	for idx, script := range scriptsStruct.Scripts {
		scripts[idx], err = p.pipeline.Script(script.Id)
		if err != nil {
			return nil, err
		}
	}
	return scripts, err
}

func (p PipelineLink) Execute(job JobRequest) error {
	return nil
}

func jobRequestToPipeline(req JobRequest) (pReq pipeline.JobRequest, err error) {
	pReq = pipeline.JobRequest{
		Script: pipeline.Script{Id: req.Script},
	}
	for name, values := range req.Inputs {
		input := pipeline.Input{Name: name}
		for _, value := range values {
			input.Items = append(input.Items, pipeline.Item{Value: value.String()})
		}
		pReq.Inputs = append(pReq.Inputs, input)
	}
	for name, values := range req.Options {
		option := pipeline.Option{Name: name}
		if len(values) > 1 {
			for _, value := range values {
				option.Items = append(option.Items, pipeline.Item{Value: value})
			}
		} else {
			option.Value = values[0]
		}
		pReq.Options = append(pReq.Options, option)

	}
	return
}