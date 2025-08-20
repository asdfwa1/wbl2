package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Minishell. Enter 'exit' to quit.")

	setupSignal()

	for {
		fmt.Print("> ")
		input, err := readInput(reader)
		/*if err != nil {                  // раскомментировать если запуск программы будет производиться через go run main.go < test_script.sh
			if err == io.EOF {
				fmt.Println("\nExiting...")
				break
			}
			fmt.Printf("error with read string %v", err)
			continue
		}*/
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "" {
			continue
		}
		if input == "exit" {
			break
		}

		err = executeCommand(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func readInput(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')

	if strings.Contains(line, "\x04") { // CTRL+D
		fmt.Println("Exiting...")
		os.Exit(0)
	}
	return line, err
}

func setupSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)

	go func() {
		for {
			sig := <-signalChan
			if sig == syscall.SIGINT {
				fmt.Println("Interrupted command. Press enter to continue.")
			}
		}
	}()
}

func executeCommand(input string) error {
	parts := strings.Fields(input)

	hasConditional := false
	hasPipeline := false

	for _, part := range parts {
		if part == "&&" || part == "||" {
			hasConditional = true
		}
		if part == "|" {
			hasPipeline = true
		}
	}

	if hasConditional && hasPipeline {
		return fmt.Errorf("mixed conditional and pipeline operators are not supported")
	}

	if hasConditional {
		return executeWithConditional(input)
	}

	if hasPipeline {
		return executeWithPipeline(input)
	}

	return executeBaseCommand(input)
}

func executeBaseCommand(input string) error {
	args := strings.Fields(input)
	if len(args) == 0 {
		return nil
	}

	cmd := args[0]
	switch strings.ToLower(cmd) {
	case "cd":
		return CommandCd(args[1:])
	case "pwd":
		return CommandPwd(args[1:])
	case "echo":
		return CommandEcho(args[1:])
	case "kill":
		return CommandKill(args[1:])
	case "ps":
		return CommandPs(args[1:])
	default:
		return executeExternalCommand(args)
	}
}

func CommandCd(args []string) error {
	if len(args) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		return os.Chdir(home)
	}
	return os.Chdir(args[0])
}

func CommandPwd(args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)
	return nil
}

func CommandEcho(args []string) error {
	fmt.Println(strings.Join(args, " "))
	return nil
}

func CommandKill(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("kill missing PID")
	}

	pid, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("kill invalid PID: %s", args[0])
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = process.Kill()
	if err != nil {
		return killWindowsProcess(pid)
	}

	fmt.Printf("Process %d terminated\n", pid)
	return nil
}

func killWindowsProcess(pid int) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("kill: failed to terminate process %d: %v", pid, err)
	}

	fmt.Printf("Process %d terminated using taskkill\n", pid)
	return nil
}

func CommandPs(args []string) error {
	psArgs := []string{"aux"}
	if len(args) > 0 {
		psArgs = args
	}

	cmd := exec.Command("ps", psArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func executeExternalCommand(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func executeWithPipeline(input string) error {
	commands := strings.Split(input, "|")
	var cmds []*exec.Cmd

	for i, cmdStr := range commands {
		cmdStr = strings.TrimSpace(cmdStr)
		if cmdStr == "" {
			return fmt.Errorf("empty command in pipeline")
		}
		args := strings.Fields(cmdStr)
		cmd := exec.Command(args[0], args[1:]...)

		cmd.Stderr = os.Stderr

		if i > 0 {
			stdin, err := cmds[i-1].StdoutPipe()
			if err != nil {
				return err
			}
			cmd.Stdin = stdin
		} else {
			cmd.Stdin = os.Stdin
		}

		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		cmds[len(cmds)-1].Stdout = os.Stdout
	}

	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return err
		}
	}

	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			if cmd != cmds[len(cmds)-1] {
				continue
			}
			return err
		}
	}

	return nil
}

func executeWithConditional(input string) error {
	parts := strings.Fields(input)
	var commands [][]string
	var operators []string
	var curCommand []string
	var lastSuccess bool

	for _, part := range parts {
		if part == "&&" || part == "||" {
			if len(curCommand) > 0 {
				commands = append(commands, curCommand)
				operators = append(operators, part)
				curCommand = []string{}
			}
		} else {
			curCommand = append(curCommand, part)
		}
	}

	if len(curCommand) > 0 {
		commands = append(commands, curCommand)
	}

	if len(commands) == 0 {
		return nil
	}

	err := executeCommandSlice(commands[0])
	if err == nil {
		lastSuccess = true
	} else {
		lastSuccess = false
	}

	for i := 0; i < len(operators); i++ {
		if i+1 >= len(commands) {
			break
		}

		operator := operators[i]
		if (operator == "&&" && lastSuccess) || (operator == "||" && !lastSuccess) {
			err := executeCommandSlice(commands[i+1])
			if err == nil {
				lastSuccess = true
			} else {
				lastSuccess = false
			}
		} else {
			lastSuccess = false
		}
	}
	return nil
}

func executeCommandSlice(args []string) error {
	if len(args) == 0 {
		return nil
	}
	return executeCommand(strings.Join(args, " "))
}
