package ansi

import (
	"bytes"

	. "github.com/charmbracelet/x/exp/term/ansi/parser"
	"github.com/rivo/uniseg"
)

// Strip removes ANSI escape codes from a string.
func Strip(s string) string {
	var (
		buf    bytes.Buffer  // buffer for collecting printable characters
		ri     int           // rune index
		rw     int           // rune width
		pstate = GroundState // initial state
	)

	// This implements a subset of the Parser to only collect runes and
	// printable characters.
	for i := 0; i < len(s); i++ {
		var state, action byte
		if pstate != Utf8State {
			state, action = Table.Transition(pstate, s[i])
		}

		// log.Printf("pstate: %s, state: %s, action: %s, code: %c, buf: %q", StateNames[pstate], StateNames[state], ActionNames[action], s[i], buf.String())
		switch {
		case pstate == Utf8State:
			// During this state, collect rw bytes to form a valid rune in the
			// buffer. After getting all the rune bytes into the buffer,
			// transition to GroundState and reset the counters.
			buf.WriteByte(s[i])
			ri++
			if ri < rw {
				continue
			}
			pstate = GroundState
			ri = 0
			rw = 0
		case action == CollectAction:
			// This action happens when we transition to the Utf8State.
			if w := utf8ByteLen(s[i]); w > 1 {
				rw = w
				buf.WriteByte(s[i])
				ri++
			}
		case action == PrintAction || action == ExecuteAction:
			// collects printable ASCII and non-printable characters
			buf.WriteByte(s[i])
		}

		// Transition to the next state.
		// The Utf8State is managed separately above.
		if pstate != Utf8State {
			pstate = state
		}
	}

	return buf.String()
}

// StringWidth returns the width of a string in cells. This is the number of
// cells that the string will occupy when printed in a terminal. ANSI escape
// codes are ignored and wide characters (such as East Asians and emojis) are
// accounted for.
func StringWidth(s string) int {
	if s == "" {
		return 0
	}
	return uniseg.StringWidth(Strip(s))
}
