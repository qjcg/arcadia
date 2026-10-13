# Job-control spike: fg/bg over a mvdan/sh interp engine

Findings for issue #12, the go/no-go gate ADR 0004 set before the
engine swap. The prototype is `spike/jobctl`; it runs real processes
and real signals over a persistent `interp.Runner` and is kept as
executable evidence.

## Question

Can terebra's job control — `&`, `jobs`, `fg`, `bg` — run over
mvdan/sh's `interp.Runner`, given that `interp` has no job tables and
no SIGTSTP handling of its own?

## Method

The spike drives one external command (`sleep 30`) through a REPL loop
where every line is parsed with `syntax` and evaluated by a persistent
`interp.Runner`. A custom exec handler spawns commands in their own
process groups and waits with `WUNTRACED`. The scenario proves:

- a foreground command blocks the shell;
- SIGTSTP to the job's process group stops it and returns control to
  the shell (the handler returns instead of blocking);
- `jobs` lists the stopped job;
- `fg` resumes and blocks the shell again, and a second stop returns
  control again;
- `bg` resumes without blocking;
- `sleep 30 &` never blocks the shell and still registers a job.

## Findings

What `interp` covers (nothing to build):

- Parsing and evaluation of `&`, pipelines, redirections, exit
  statuses, and the rest of bash semantics.
- Running each external command through `ExecHandlerFunc`; the
  handler receives the interpreter's stdio, environment, and directory
  via `HandlerCtx`.
- Backgrounding: `cmd &` returns from `Runner.Run` immediately while
  the handler runs detached, so the shell stays responsive and the
  handler can still record the job.

What terebra keeps owning (unchanged from the current `job.go`):

- The job table and `%` job specs.
- Process-group management (`Setpgid`, one group per job).
- Signal plumbing: `SIGTSTP`/`SIGCONT` delivery and `SIGTTOU`
  ignoring around terminal-group switches.
- Waiting with `WUNTRACED`/`WCONTINUED`: `os/exec.Cmd.Wait` cannot
  report stops, so the handler waits with `syscall.Wait4` directly and
  returns an exit status (128+signal) on stop, which is exactly the
  hook that hands control back to the REPL on Ctrl+Z.
- Terminal control (`tcsetpgrp`) when a job is foregrounded.

## Open risks (not exercised by the spike)

1. **Pipeline jobs.** `interp` runs pipeline stages concurrently and
   the handler sees each stage alone. Giving one pipeline a single
   process group (bash's job model) needs cross-stage coordination in
   the handler layer. Mitigation: group stages by a shared key derived
   from the statement position; worst case, v1 tracks per-stage jobs.
2. **Controlling terminal.** The spike proves stop/continue and job
   bookkeeping without a controlling terminal; `tcsetpgrp`/`SIGTTOU`
   porting is standard shell plumbing the current `job.go` already
   performs and must be carried over unchanged.
3. **Builtin dispatch.** `interp` builtins bypass the exec handler
   [sh-issue-705], so `jobs`/`fg`/`bg` (and terebra's other builtins)
   must be dispatched in the REPL layer before the runner, as the
   spike does, or wrapped via `CallHandlerFunc`.

## Decision

**Go.** The seam holds: job control lives entirely in the exec-handler
and REPL layers, `interp` needs no fork or patch, and the one real
unknown (pipeline job grouping) has a workable mitigation. The engine
swap in issues #16 and #17 is unblocked from the job-control side;
issue #15 (builtin behavior diff) remains.

## Confirmation

- `go test ./spike/jobctl/` passes repeatedly (verified with
  `-count=5`) with real processes and signals.
- The scenario covers every acceptance criterion in issue #12.

## Link references

[sh-issue-705]: https://github.com/mvdan/sh/issues/705
