// A Shell also called Read/Evaludate/Print/Loop (REPL) does mainly
// - Read the user input
// - Evaludate that input
// - Print the result of evaluation
// - Loop

package ashell

import (
	// "bufio"
	"bufio"
	"errors"
	"fmt"
	"log"
	"slices"

	// "log"
	"os"
	"os/exec"

	"strings"

	"golang.org/x/term"
)

type Key byte

const (
	KeyNULL Key = iota
	KeyEOT      = 0x4  // <C-d>
	KeyFF       = 0xc  // <C-l>
	KeyCR       = 0xd  // <Enter>
	KeyDEL      = 0x7f // <Backspace>
	KeySO       = 0xe  // <C-n>
	KeyDLE      = 0x10 // <C-p>
)

const (
	Chdir string = "cd"
	Exit         = "exit"
)

var CommandHistory []string
var CommandHistoryCursor int

func Run() {
	// we want to have total control of the input character
	// so we will not use this high level Api
	// terminal := term.NewTerminal(os.Stdin, prompt())
	var err error
	fd := int(os.Stdin.Fd())

	mode, err := term.MakeRaw(fd)
	if err != nil {
		log.Printf("cannot change the mode of tty (reason: %v)", err)
		return
	}
	defer term.Restore(fd, mode)

	cmd := make([]byte, 0)

	var cursor int

	reader := bufio.NewReader(os.Stdin)

	printCommand("")

	logFile, err := os.Create("/tmp/goshell.log")
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	for {

		if _, err = term.MakeRaw(fd); err != nil {
			log.Printf("cannot change the mode of tty (reason: %v)", err)
			return
		}

		ch, err := reader.ReadByte()
		if err != nil {
			continue
		}

		switch ch {
		case KeyEOT:
			return
		case KeyCR, KeyFF:
			if ch == KeyFF {
				cmd = []byte("clear")
			}

			// the cmd is empty
			if len(cmd) == 0 {
				fmt.Println()
				printCommand("")
				continue
			}

			// move the cursor complety at left
			fmt.Print("\n\033[1000D")

			if err = term.Restore(fd, mode); err != nil {
				log.Println("cannot restore the default mode of the terminal, your terminal may behave wierdly")
				return
			}

			if err = executeCommand(string(cmd)); err != nil {
				log.Printf("%v", err)
			}

			CommandHistory = append(CommandHistory, string(cmd))
			CommandHistoryCursor = len(CommandHistory)

			// reset the cursor position
			cursor = 0
			cmd = slices.Delete[[]byte](cmd, 0, len(cmd))
		case KeyDEL:
			cmd = slices.Delete[[]byte](cmd, max(0, cursor-1), max(cursor, 0))
			cursor--
			cursor = max(0, cursor)
		case KeySO, KeyDLE:
			historyLen := len(CommandHistory)
			if historyLen == 0 {
				continue
			}
			if ch == KeySO {
				CommandHistoryCursor = min(historyLen-1, CommandHistoryCursor+1)
			} else {
				CommandHistoryCursor = max(0, CommandHistoryCursor-1)
			}
			item := CommandHistory[CommandHistoryCursor]
			printCommand(item)
			cmd = []byte(item)
			cursor = len(cmd)
			continue
		default:
			fmt.Fprintln(logFile, ch)
			if IsPrintable(ch) {
				cmd = slices.Insert[[]byte](cmd, cursor, ch)
				cursor++
			}
		}
		printCommand(string(cmd))
	}
}

func printCommand(cmd string) {
	fmt.Print("\033[2K") // clear the entire line
	fmt.Print("\033[1000D")
	fmt.Print(prompt()) // print the prompt
	fmt.Print(string(cmd))
}

func IsPrintable(ch byte) bool {
	return 32 <= ch && ch <= 126
}

func prompt() string {
	hostname, err := os.Hostname()
	prompt := "> "
	if err == nil {
		prompt = fmt.Sprintf("%s)-> ", hostname)
	}
	return prompt
}

func executeCommand(input string) error {
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
