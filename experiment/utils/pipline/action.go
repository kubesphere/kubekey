package pipline

import (
	"errors"
	"fmt"
)

type Action interface {
	Execute(vars *Vars) (result *Result, err error)
}

type Command struct {
	Cmd    string
	Result Result
}

type Copy struct {
	Src    string
	Dst    string
	Result Result
}

type Template struct {
	Dst  string
	Data map[string]interface{}
}

type WebServer struct {
	ListenPort    int
	ListenAddress string // 127.0.0.1, 0.0.0.0, *
	Status        string // start, restart, stop
}

type Func struct {
	function func(vars *Vars) (*Result, error)
}

func (f *Func) Execute(vars *Vars) (*Result, error) {
	return f.function(vars)
}

func (a *Command) Execute(vars *Vars) (*Result, error) {
	fmt.Println(a.Cmd)
	return &Result{}, errors.New("123")
}

func (c *Copy) Execute(vars *Vars) (*Result, error) {
	fmt.Println(c.Dst, c.Src)
	return &Result{}, errors.New("123")
}

func (t *Template) Execute(vars *Vars) (*Result, error) {
	fmt.Println(t.Data)
	return &Result{}, errors.New("123")
}

func (w *WebServer) Execute(vars *Vars) (*Result, error) {
	fmt.Println(w.ListenPort)
	return &Result{}, errors.New("123")
}
