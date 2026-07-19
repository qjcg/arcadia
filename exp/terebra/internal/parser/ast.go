package parser

import "fmt"

type Node interface {
	node()
}

type ChainingOp int

const (
	ChainingThen ChainingOp = iota // ;
	ChainingAnd                    // &&
	ChainingOr                     // ||
)

func (co ChainingOp) String() string {
	switch co {
	case ChainingThen:
		return ";"
	case ChainingAnd:
		return "&&"
	case ChainingOr:
		return "||"
	default:
		return fmt.Sprintf("ChainingOp(%d)", co)
	}
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

// Script is a sequence of pipelines separated by chaining operators (&&, ||, ;).
type Script struct {
	Pipelines []*Pipeline
	Ops       []ChainingOp // Ops[i] connects Pipelines[i] to Pipelines[i+1]
}

func (s *Script) node() {}

type Pipeline struct {
	Commands []*Command
	Connects []ConnectType // Connects[i] connects Commands[i] to Commands[i+1]
	Encoder  string        // "json", "yaml", or "cue" for |>encoder pipe endpoint
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
	RedirectHeredoc
	RedirectHeredocDash // <<- (strip leading tabs)
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
	case RedirectHeredoc:
		return "<<"
	case RedirectHeredocDash:
		return "<<-"
	default:
		return fmt.Sprintf("RedirectType(%d)", rt)
	}
}

type Redirect struct {
	Type    RedirectType
	File    string
	Content string // For heredocs, the content between lines
	Quoted  bool   // For heredocs, whether delimiter was quoted
}

func (r *Redirect) String() string {
	if r.Type == RedirectHeredoc || r.Type == RedirectHeredocDash {
		return fmt.Sprintf("%s %s", r.Type, r.File)
	}
	return fmt.Sprintf("%s %s", r.Type, r.File)
}
