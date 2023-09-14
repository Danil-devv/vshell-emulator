package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var SupportedCommands = map[string]struct{}{
	"ls":   {},
	"cd":   {},
	"cat":  {},
	"pwd":  {},
	"exit": {},
}

const (
	HELP = iota
	LS
	CD_BACK
	CD_TO
	CAT
	PWD
	EXIT
)

type Options struct {
	path string
}

func ParseFlags(opts *Options) {
	flag.StringVar(&(*opts).path, "systemimage", "",
		"path to file system image. must have .zip or .tar extension")
	flag.Parse()
	return
}

func ValidateFlags(opts *Options) error {
	if opts.path == "" {
		return errors.New("the path to the file system image is not set")
	}

	if _, err := os.Stat(opts.path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(`file "%s" does not exist`, opts.path)
	} else if err != nil {
		return fmt.Errorf("unexpected error: %s", err.Error())
	}

	if s := strings.Split(opts.path, "."); s[len(s)-1] != "tar" && s[len(s)-1] != "zip" {
		return errors.New("file extension must be .tar or .zip")
	}

	return nil
}

func HelpInfo(cmd string) string {
	switch cmd {
	case "ls":
		return fmt.Sprintf("Usage: ls\nList information about the FILEs (the current directory)")
	case "cd":
		return fmt.Sprintf("Usage: cd [PATH TO FILE]...\n" +
			"'..' is processed by removing the immediately previous pathname component back to a slash or" +
			" the beginning of DIR.")
	case "cat":
		return fmt.Sprintf("Usage: cat [FILE]\nPrint FILE to standard output.")
	case "pwd":
		return fmt.Sprintf("Usage: pwd\nPrint the name of the current working directory.")
	default:
		return "unexpected command to help info"
	}
}

func ValidateCommand(cmd []string) (int, error) {
	if len(cmd) == 0 {
		return -1, fmt.Errorf("the command is empty")
	}

	switch cmd[0] {
	case "ls":
		if len(cmd) == 1 {
			return LS, nil
		} else if len(cmd) > 2 {
			return -1, fmt.Errorf("unsupported arg: %s", cmd[1:])
		} else if cmd[1] == "--help" {
			return HELP, nil
		}
		return -1, fmt.Errorf("unsupported arg: %s", cmd[1])
	case "cd":
		if len(cmd) != 2 {
			return -1, fmt.Errorf("unsupported arg")
		}
		if cmd[1] == "--help" {
			return HELP, nil
		} else if cmd[1] == ".." {
			return CD_BACK, nil
		}
		return CD_TO, nil
	case "cat":
		if len(cmd) != 2 {
			return -1, fmt.Errorf("unsupported arg")
		}
		if cmd[1] == "--help" {
			return HELP, nil
		}
		return CAT, nil
	case "pwd":
		if len(cmd) > 2 {
			return -1, fmt.Errorf(`"%s" does not support any args`, cmd[0])
		}
		if len(cmd) == 2 {
			if cmd[1] == "--help" {
				return HELP, nil
			}
			return -1, fmt.Errorf(`"%s" does not support any args instead of '--help'`, cmd[0])
		}
		return PWD, nil
	case "exit":
		return EXIT, nil
	default:
		return -1, fmt.Errorf("unsupported command: %s", cmd[0])
	}
}

func ReadCommand(in *bufio.Reader) ([]string, error) {
	s, err := in.ReadString('\n')
	if err != nil {
		return nil, err
	}

	s = strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	return strings.Split(s, " "), nil
}

func main() {
	defer func() { fmt.Println("The program is finished") }()

	opts := Options{}

	ParseFlags(&opts)
	if err := ValidateFlags(&opts); err != nil {
		log.Fatalf("validation error: %s", err)
	}

	fmt.Println("use command 'exit' to terminate program")
	in := bufio.NewReader(os.Stdin)

	toExit := false
	for !toExit {
		cmd, err := ReadCommand(in)
		if err != nil {
			log.Fatalf("error while reading: %s", err)
		}

		commandType, err := ValidateCommand(cmd)
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch commandType {
		case EXIT:
			toExit = true
		case HELP:
			fmt.Println(HelpInfo(cmd[0]))
		case LS:

		}
	}
}
