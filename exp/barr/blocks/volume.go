package blocks

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Volume represents the volume of the default pulseaudio sink.
type Volume struct {
	Level string
	Mute  bool
}

func (v *Volume) String() string {
	ctx := context.Background()
	output, err := exec.CommandContext(ctx, "pactl", "list", "sinks").Output()
	if err != nil {
		return "pactl error"
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.Contains(line, "Mute"):
			v.Mute = strings.Contains(strings.Fields(line)[1], "yes")
		case strings.Contains(line, "front-left:"):
			v.Level = strings.Fields(line)[4]
		}
	}

	if v.Mute {
		return fmt.Sprintf("🔇%s", v.Level)
	} else {
		return fmt.Sprintf("🔊%s", v.Level)
	}
}
