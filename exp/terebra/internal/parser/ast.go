package parser

import "fmt"

type Node interface {
	node()
}

type ConnectType int

const (
	ConnectPipe ConnectType = iota
	ConnectAuger
)

func (ct ConnectType) String() string {
	switch ct {
	case ConnectPipe:
		return "|"
	case ConnectAuger:
		return "|>"
	default:
		return fmt.Sprintf("ConnectType(%d)", ct)
	}
}

type Pipeline struct {
	Commands []*Command
	Connects []ConnectType // Connects[i] connects Commands[i] to Commands[i+1]
}

func (p *Pipeline) node() {}

type Command struct {
	Name       string
	Args       []string
	Redirects  []*Redirect
	Background bool
}

func (c *Command) node() {}

type RedirectType int

const (
	RedirectStdout RedirectType = iota
	RedirectStderr
	RedirectAppend
	RedirectStderrAppend
	RedirectStdin
	RedirectStderrToStdout
)

func (rt RedirectType) String() string {
	switch rt {
	case RedirectStdout:
		return ">"
	case RedirectStderr:
		return "2>"
	case RedirectAppend:
		return ">>"
	case RedirectStderrAppend:
		return "2>>"
	case RedirectStdin:
		return "<"
	case RedirectStderrToStdout:
		return "2>&1"
	default:
		return fmt.Sprintf("RedirectType(%d)", rt)
	}
}

type Redirect struct {
	Type RedirectType
	File string
}

func (r *Redirect) String() string {
	return fmt.Sprintf("%s %s", r.Type, r.File)
}
