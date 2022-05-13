package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/runner-x/runner-x/util/iohelpers"
	"github.com/runner-x/runner-x/util/print"
)

const (
	DefaultNproc       = 20
	DefaultFsize       = 1000
	DefaultUid         = 1234
	DefaultGid         = 1234
	ProcessCommandName = "process"
)

// RuntimeAgent struct stores metadata to execute user code with restricted host resources
type RuntimeAgent struct {
	Id       string
	Provider ArgProvider
	Uid      int
	Gid      int

	// workdir represents the directory where all commands should be run
	workdir string

	// rwmutex restricts access to running code with the RuntimeAgent
	rwmutex sync.RWMutex
	state   State
}

func NewTimeoutRuntime(id string, provider ArgProvider) *RuntimeAgent {
	return &RuntimeAgent{Id: id, Provider: provider, Uid: DefaultUid, Gid: DefaultGid}
}

func NewRuntimeAgentWithIds(idStr string, id int, provider ArgProvider, workdir string) *RuntimeAgent {
	return &RuntimeAgent{
		Id:       idStr,
		Provider: provider,
		Uid:      id,
		Gid:      id,
		workdir:  workdir,
		state:    Ready,
		rwmutex:  sync.RWMutex{},
	}
}

func (r *RuntimeAgent) runCmd(props *RunProps) (*RunOutput, error) {
	if props == nil {
		return nil, nil
	}

	// Create a new context and add a timeout to it
	timeout := time.Second * time.Duration(props.Timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd

	// TODO: args provider logic should be abstracted into controller
	numArgs := len(props.RunArgs)
	if numArgs < 1 {
		return nil, nil
	} else {
		cmd = r.Provider.Provide(&ctx, props)
	}

	stdoutPipe, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		panic(stdoutErr)
	}

	stderrPipe, stderrErr := cmd.StderrPipe()
	if stderrErr != nil {
		panic(stderrErr)
	}

	stdoutChannel := iohelpers.GetWriterChannelOutput(stdoutPipe)
	stderrChannel := iohelpers.GetWriterChannelOutput(stderrPipe)

	print.DebugPrintf("\nrunning command with RunProps: %v\n", props)
	print.DebugPrintf("running command from PID: %v\n", os.Getpid())

	err := cmd.Run()

	stdoutAsString := <-stdoutChannel
	stderrAsString := <-stderrChannel

	if ctx.Err() == context.DeadlineExceeded {
		print.DebugPrintf("command error: %v\n", err)
		fmt.Println("command timed out")
		return &RunOutput{
			Stdout: stdoutAsString,
			Stderr: stderrAsString,
		}, err
	}

	if err != nil {
		print.DebugPrintf("(runCommand) error from running command: %v", err)
	}

	return &RunOutput{
		Stdout: stdoutAsString,
		Stderr: stderrAsString,
	}, err
}

func (r *RuntimeAgent) RunCmd(runprops *RunProps) (*RunOutput, error) {
	return r.runCmd(runprops)
}

func (r *RuntimeAgent) IsReady() bool {
	r.rwmutex.RLock() // acquire the lock to read to reading while agent is trying to write to the state

	defer r.rwmutex.RUnlock() // make sure we unlock when we're done
	return r.state == Ready
}

// setState locks rwmutex before changing state
func (r *RuntimeAgent) setState(state State) {
	r.rwmutex.Lock()
	defer r.rwmutex.Unlock()
	r.state = state
}

// SafeRunCmd will acquire lock and set state to NotReady while running the command
// 		This function should ensure that threads can see if RuntimeAgent IsReady() to run a
// 		command with minimal blocking. Since the IsReady() command uses a rwmutex, it
// 		should only require a read lock which should be faster to acquire than a normal
// 		mutex.
func (r *RuntimeAgent) SafeRunCmd(props *RunProps) (*RunOutput, error) {
	r.setState(NotReady)
	defer r.setState(Ready)
	return r.runCmd(props)
}

func (r *RuntimeAgent) RuntimeUid() int {
	return r.Uid
}

func (r *RuntimeAgent) RuntimeGid() int {
	return r.Gid
}

func (r *RuntimeAgent) Workdir() string {
	return r.workdir
}

func (r *RuntimeAgent) writeToWorkdir(props *FileProps) error {

	return nil
}
