package repl

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/c-bata/go-prompt"
)

var emptyComplete = func(prompt.Document) []prompt.Suggest { return []prompt.Suggest{} }

func runPrompt(executor func(string), completer func(prompt.Document) []prompt.Suggest,
	firstTime func(), length uint16) {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(prefix),
		prompt.OptionPrefixTextColor(prompt.LightGray),
		prompt.OptionMaxSuggestion(length),
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionAddKeyBind(
			prompt.KeyBind{prompt.ControlC, func(*prompt.Buffer){
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}}),
	)
	firstTime()
	p.Run()
}

// executes prompt waiting for an input with y or n
func yesOrNoQuestion(msg string) string {
	var input string
	for {
		input = prompt.Input(prefix+msg,
			emptyComplete,
			prompt.OptionPrefixTextColor(prompt.LightGray))

		if input == "y" || input == "n" {
			break
		}

		fmt.Println(printPrefix, "invalid command.")
	}

	return input
}

func multipleChoice(names []string) string {
	var input string
	accounts := make(map[string]struct{})
	for _, name := range names {
		accounts[name] = struct{}{}
	}
	if len(accounts) == 0 {
		return ""
	}
	for {
		for ac := range accounts {
			fmt.Println(printPrefix, ac)
		}
		input = prompt.Input(prefix,
			emptyComplete,
			prompt.OptionPrefixTextColor(prompt.LightGray))

		if _, ok := accounts[input]; ok {
			return input
		}

		s := strings.TrimSpace(input)
		if s == "quit" || s == "exit" {
			fmt.Println("Bye!")
			os.Exit(0)
			return ""
		}

		fmt.Println(printPrefix, "invalid command.")

	}
}

// executes prompt waiting an input not blank
func inputNotBlank(msg string) string {
	var input string
	for {
		input = prompt.Input(prefix+msg,
			emptyComplete,
			prompt.OptionPrefixTextColor(prompt.LightGray))

		if strings.TrimSpace(input) != "" {
			break
		}

		fmt.Println(printPrefix, "please enter a value.")
	}

	return input
}
