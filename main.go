package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Arguments map[string]string

const (
	id        = "id"
	item      = "item"
	operation = "operation"
	fName     = "fileName"
)

func Perform(args Arguments, writer io.Writer) error {
	if len(args) == 0 {
		return nil
	}
	err := args.validArgs()
	if err != nil {
		return err
	}
	return nil
}

func (a Arguments) validArgs() error {
	op := map[string]string{
		"add":      fName,
		"list":     fName,
		"findById": fName,
		"remove":   fName,
	}
	for flag, value := range a {
		v := value
		if v[0] == '"' {
			v = v[1:]
		}
		lV := len(v) - 1
		if v[lV] == '"' {
			v = v[:lV]
			a[flag] = v
		}
		switch flag {
		case operation:
			if slave, ok := op[v]; ok {
				if _, ok := a[slave]; !ok {
					return errors.New("need filename")
				}
			} else {
				return fmt.Errorf("%s flag has to be specified", flag)
			}
		case fName:
			if !strings.HasSuffix(v, ".json") {
				return errors.New("wrong type")
			}
		}
	}
	return nil
}

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
			"-id":          id,
			"-item":        item,
			"-operation":   operation,
			"-fileName":    fName,
			"--id":         id,
			"--item":       item,
			"--operation":  operation,
			"--fileName":   fName,
			"-id=":         id,
			"-item=":       item,
			"-operation=":  operation,
			"-fileName=":   fName,
			"--id=":        id,
			"--item=":      item,
			"--operation=": operation,
			"--fileName=":  fName,
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

func makeJSONFile() {
}

func needLenght() (n int) {
	for _, arg := range os.Args[1:] {
		n += len(arg)
	}
	return n + 1 // +1 need for last space or will panic
}

func parseArgs() Arguments {
	a := new(args)
	if a.fixArgs() {
		return a.parse()
	}
	return nil
}

func main() {
	os.Args = append(os.Args, `-operation «add» -item ‘{«id»: "1", «email»: «email@test.com», «age»: 23}’ --fileName «users.json»`)
	// os.Args = append(os.Args, "")
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
