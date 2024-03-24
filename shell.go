// A Shell also called Read/Evaludate/Print/Loop (REPL) does mainly
// - Read the user input
// - Evaludate that input
// - Print the result of evaluation
// - Loop

package ashell

import (
	"bufio"
	"errors"
	"fmt"
	// "log"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/term"
)

const (
	KeyDefault byte = iota
	KeyCtrlL        = 27
)

const (
	Chdir string = "cd"
	Exit         = "exit"
)

func Run() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, prompt())

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if err = runCommand(input); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func prompt() string {
	hostname, err := os.Hostname()
	prompt := "> "
	if err == nil {
		prompt = fmt.Sprintf("%s)-> ", hostname)
	}
	return prompt
}

func runCommand(input string) error {
	input = strings.Trim(input, "\n")
	prog := strings.Fields(input)

	switch prog[0] {
	// There is no executable called cd on linux systems
	// every shell has to provide it's own implementation
	// A good explanation for why it is like that is given
	// in here https://brennan.io/2015/01/16/write-a-shell-in-c/
	case Chdir:
		if len(prog[1:]) == 0 {
			pth, err := os.UserHomeDir()
			if err != nil {
				return errors.New("path required")
			}
			prog = append(prog, pth)
		}
		return os.Chdir(prog[1])
	// There is no executable called exit on linux systems
	// every shell has to provide it's own implementation
	// A good explanation for why it is like that is given
	// in here https://brennan.io/2015/01/16/write-a-shell-in-c/
	case Exit:
		os.Exit(0)
	}

	cmd := exec.Command(prog[0], prog[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func runHotKey(input byte) error {
	var err error

	hotKey, err := getHotKey()
	if err != nil {
		return err
	}
	var prog string
	switch hotKey {
	case KeyCtrlL:
		prog = "clear"
	default:
		//
	}
	cmd := exec.Command(prog)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func getHotKey() (byte, error) {
	var err error

	terminal, err := term.Open("/dev/tty")
	if err != nil {
		return KeyDefault, fmt.Errorf("cannot open the TTY (reason: %w)", err)
	}
	defer func() {
		terminal.Close()
		terminal.Restore()
	}()

	if err = term.RawMode(terminal); err != nil {
		return 0, fmt.Errorf("cannot change the TTY mode (reason: %w)", err)
	}

	maxByteNumber := 3
	buf := make([]byte, maxByteNumber)
	nOfByteRead, err := terminal.Read(buf)

	if err = term.RawMode(terminal); err != nil {
		return KeyDefault, fmt.Errorf("cannot read from TTY mode (reason: %w)", err)
	}

	// Arrow keys are prefixed with the ANSI escape code which take up the first two bytes.
	// The third byte is the key specific value we are looking for.
	// For example the left arrow key is '<esc>[A' while the right is '<esc>[C'
	// See: https://en.wikipedia.org/wiki/ANSI_escape_code
	keyIdx := 0
	if nOfByteRead == maxByteNumber {
		keyIdx = 2
	}
	return buf[keyIdx], nil
}
