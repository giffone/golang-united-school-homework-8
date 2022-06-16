package main

import (
	"fmt"
	"os"
	"strings"
)

type args struct {
	r         []rune
	lR, index int
}

func (a *args) fixArgs() bool {
	lArgs := len(os.Args)
	if lArgs == 1 {
		return false
	}
	if lArgs == 2 && os.Args[1] == "" {
		return false
	}
	a.lR = needLenght()
	a.r = make([]rune, a.lR)
	a.index = -1
	for _, arg := range os.Args[1:] {
		for _, r := range arg {
			a.index++
			if r == '\'' || r == '«' || r == '»' || r == '‘' || r == '’' || r == '`' {
				a.r[a.index] = '"'
				continue
			}
			a.r[a.index] = r
		}
		a.index++
		a.r[a.index] = ' '
	}
	return true
}

func (a *args) parse() Arguments {
	arg := Arguments{}
	flags := struct {
		all     map[string]string
		catched string
		match   func(string) (string, bool)
	}{
		all: map[string]string{
			"-id":          arg_id,
			"-item":        arg_item,
			"-operation":   arg_oper,
			"-fileName":    arg_fName,
			"--id":         arg_id,
			"--item":       arg_item,
			"--operation":  arg_oper,
			"--fileName":   arg_fName,
			"-id=":         arg_id,
			"-item=":       arg_item,
			"-operation=":  arg_oper,
			"-fileName=":   arg_fName,
			"--id=":        arg_id,
			"--item=":      arg_item,
			"--operation=": arg_oper,
			"--fileName=":  arg_fName,
		},
	}
	flags.match = func(word string) (string, bool) {
		w := strings.TrimSpace(word)
		if flag, ok := flags.all[w]; ok {
			return flag, true
		}
		return "", false
	}
	j := 0    // start value of flag from
	a.index++ // adding last space for catch last value
	for i := 0; i < len(a.r[:a.index]); i++ {
		symbol := a.r[:a.index][i]
		if symbol == ' ' || symbol == '=' {
			value := string(a.r[:a.index][j:i])
			if flag, ok := flags.match(value); ok {
				flags.catched = flag
				j = i + 1
				continue
			}
			if flags.catched != "" {
				arg[flags.catched] = fmt.Sprintf("%s%s", arg[flags.catched], value)
				j = i
			}
		}
	}
	return arg
}

func needLenght() (n int) {
	for _, arg := range os.Args[1:] {
		n += len(arg)
	}
	return n + 1 // +1 need for last space or will panic
}
