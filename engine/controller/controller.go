package controller

import (
	"strconv"
	"sync"

	"github.com/runner-x/runner-x/engine/runtime"
)

type Controller interface {
	SubmitRequest(runprops *runtime.RunProps) (*runtime.RunOutput, error)
	runRequest(runprops *runtime.RunProps, agent *runtime.RuntimeAgent)
}

type AsyncController struct {
	agents map[int]*agentData
}

// State represents whether the worker is ready for another job or not
type State string

const (

	// Ready means ready to run another job and no other jobs are currently running
	Ready = State("ready")

	// NotReady means not ready to run another job since agent is running request or cleaning up from a finished job
	NotReady = State("notready")
)

type agentData struct {
	state   State
	rwmutex sync.RWMutex
	agent   runtime.RuntimeAgent
}

func NewAsyncControllerWithMap(a map[int]*agentData) *AsyncController {
	return &AsyncController{agents: a}
}

func NewAsyncController(size int) *AsyncController {
	m := make(map[int]*agentData)

	for i := int(0); i < size; i++ {
		key := int(i + 1)
		m[key] = &agentData{
			state:   Ready,
			rwmutex: sync.RWMutex{},
			agent: runtime.RuntimeAgent{
				Id:       "agent_" + strconv.FormatInt(int64(key), 10),
				Provider: &runtime.ProcessorArgsProvider{},
				Uid:      key,
				Gid:      key,
			},
		}
	}
	return &AsyncController{agents: m}
}

func (ac *AsyncController) SubmitRequest(runprops *runtime.RunProps) (*runtime.RunOutput, error) {
	return nil, nil
}