package shell

import (
	"os"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/creack/pty"
)

// openPty opens a pty pair and returns the master and slave.
func openPty() (*os.File, *os.File, error) {
	return pty.Open()
}

// makeRaw puts the given terminal fd into raw mode, matching what readline
// does during a Readline() call.
func makeRaw(f *os.File) error {
	var t syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TCGETS, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return errno
	}
	t.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	t.Iflag &^= syscall.ICRNL
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TCSETS, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return errno
	}
	return nil
}

// TestStdinCopierInterrupt verifies that the stdin copy goroutine can be
// stopped without closing os.Stdin. This is the core of the fix for the
// "second Enter required" bug: the old code closed os.Stdin to unblock the
// copier, but in raw mode a pending read on a terminal fd is not unblocked
// by close(), which deadlocked pauseStdinPipe/closeStdinPipe.
//
// We simulate a terminal by pointing os.Stdin at a pty and putting it in
// raw mode (as readline does), then verify pauseStdinPipe returns promptly.
func TestStdinCopierInterrupt(t *testing.T) {
	// Only meaningful on a real terminal-like fd; skip if we can't set one up.
	ptmx, tty, err := openPty()
	if err != nil {
		t.Skipf("no pty available: %v", err)
	}
	defer ptmx.Close()
	defer tty.Close()

	oldStdin := os.Stdin
	os.Stdin = tty
	defer func() { os.Stdin = oldStdin }()

	// Put the pty slave in raw mode, matching what readline does during a
	// Readline() call. In this mode closing the fd does not unblock a
	// pending read, which is exactly the condition that used to deadlock.
	if err := makeRaw(tty); err != nil {
		t.Fatalf("makeRaw: %v", err)
	}

	s := New()
	s.setupStdinPipe()
	defer s.closeStdinPipe()

	// Give the copier a moment to block in its read on the raw pty.
	time.Sleep(50 * time.Millisecond)

	// pauseStdinPipe must return promptly. Before the fix it blocked
	// forever waiting for the copier, which was stuck in os.Stdin.Read.
	done := make(chan struct{})
	go func() {
		s.pauseStdinPipe()
		close(done)
	}()
	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("pauseStdinPipe deadlocked: copier was not interrupted by closing os.Stdin")
	}
}

// TestStdinCopierForwardsData verifies the copier reads bytes from os.Stdin
// and writes them to the stdin pipe.
func TestStdinCopierForwardsData(t *testing.T) {
	ptmx, tty, err := openPty()
	if err != nil {
		t.Skipf("no pty available: %v", err)
	}
	defer ptmx.Close()
	defer tty.Close()

	oldStdin := os.Stdin
	os.Stdin = tty
	defer func() { os.Stdin = oldStdin }()

	if err := makeRaw(tty); err != nil {
		t.Fatalf("makeRaw: %v", err)
	}

	s := New()
	s.setupStdinPipe()
	defer s.closeStdinPipe()

	// Give the copier a moment to block in select on the pty.
	time.Sleep(50 * time.Millisecond)

	// Write data to the pty master; the copier should forward it to s.stdinR.
	go func() {
		ptmx.Write([]byte("hello"))
	}()

	// Read from the stdin pipe read end with a timeout.
	buf := make([]byte, 16)
	done := make(chan int, 1)
	go func() {
		n, _ := s.stdinR.Read(buf)
		done <- n
	}()
	select {
	case n := <-done:
		if n == 0 {
			t.Fatal("expected data forwarded")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("copier did not forward data")
	}
}

// TestReplRunsCommand drives Repl with a pty-backed stdin, feeding a command
// followed by Ctrl+D (EOF). A hard timeout prevents indefinite hangs.
func TestReplRunsCommand(t *testing.T) {
	ptmx, tty, err := openPty()
	if err != nil {
		t.Skipf("no pty: %v", err)
	}
	defer ptmx.Close()
	defer tty.Close()

	if err := makeRaw(tty); err != nil {
		t.Fatalf("makeRaw: %v", err)
	}

	oldIn := os.Stdin
	os.Stdin = tty
	defer func() { os.Stdin = oldIn }()

	s := newTestShell()

	done := make(chan error, 1)
	go func() {
		done <- s.Repl()
	}()

	// Feed a command, then Ctrl+D to exit the REPL.
	go func() {
		time.Sleep(150 * time.Millisecond)
		ptmx.Write([]byte("echo hi\n"))
		time.Sleep(100 * time.Millisecond)
		ptmx.Write([]byte{0x04})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Repl did not exit")
	}
}
