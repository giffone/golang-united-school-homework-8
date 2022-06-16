package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Arguments map[string]string

const (
	arg_id    = "id"
	arg_item  = "item"
	arg_oper  = "operation"
	arg_fName = "fileName"
)

func Perform(args Arguments, writer io.Writer) error {
	if len(args) == 0 {
		return nil
	}
	err := args.validArgs()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(args[arg_fName], os.O_RDWR|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	bRd, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	defer file.Close()
	if args[arg_oper] == "list" {
		writer.Write(bRd)
		return nil
	}
	// for operations exept "list"
	// make structure for "json-arg" from args and "json-file" from file
	tj := newTwoJSON()
	args.fixJSON()
	err = tj.unmarshalMe([2][]byte{[]byte(args[arg_item]), bRd}) // {"json-arg", "json-file"}
	if err != nil {
		return err
	}
	switch args[arg_oper] {
	case "add":
		if id, ok := tj.isExist(tj.ids()); ok {
			e := fmt.Sprintf("Item with id %s already exists", id)
			writer.Write([]byte(e))
			return nil
		}
		if err != nil {
			writer.Write([]byte(err.Error()))
			return nil
		}
		file.Write([]byte(args[arg_item]))
		writer.Write([]byte(args[arg_item]))
	case "findById":
		if j, err := tj.find(args[arg_id]); err == nil && j != nil {
			writer.Write(j)
		}
	case "remove":
		if _, ok := tj.isExist([]string{args[arg_id]}); !ok {
			e := fmt.Sprintf("Item with id %s not found", args[arg_id])
			writer.Write([]byte(e))
			return nil
		}
		if j, err := tj.remove(args[arg_id]); err == nil && j != nil {
			os.Remove(args[arg_fName])
			f, err := os.OpenFile(args[arg_fName], os.O_RDWR|os.O_APPEND|os.O_CREATE, 0o644)
			if err != nil {
				return err
			}
			f.Write(j)
		}
	}
	return nil
}

func (a Arguments) validArgs() error {
	op := map[string]string{
		"add":      arg_item,
		"list":     "",
		"findById": arg_id,
		"remove":   arg_id,
	}
	// check flags by priority
	for _, flag := range []string{arg_oper, arg_fName} {
		if v, ok := a[flag]; ok {
			err := a.valid(flag, v, op)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("-%s flag is missing", flag)
		}
	}
	return nil
}

func (a Arguments) valid(flag, value string, op map[string]string) error {
	value = strings.TrimSpace(value)
	if len(value) == 0 {
		return fmt.Errorf("-%s flag has to be specified", flag)
	}
	if value[0] == '"' {
		value = value[1:]
	}
	if len(value) == 0 {
		return fmt.Errorf("-%s flag has to be specified", flag)
	}
	lV := len(value) - 1
	if value[lV] == '"' {
		value = value[:lV]
		a[flag] = value
	}
	switch flag {
	case arg_oper:
		if slave, ok := op[value]; !ok {
			if value == "" {
				return fmt.Errorf("-%s flag has to be specified", flag)
			}
			return fmt.Errorf("Operation %s not allowed!", value)
		} else {
			if slave == "" {
				return nil
			}
			if v, ok := a[slave]; !ok || v == "" {
				return fmt.Errorf("-%s flag has to be specified", slave)
			}
		}
	case arg_fName:
		if !strings.HasSuffix(value, ".json") {
			return errors.New("wrong type")
		}
	}
	return nil
}

func (a Arguments) fixJSON() {
	if len(a[arg_item]) > 0 && a[arg_item][0] != '[' {
		a[arg_item] = fmt.Sprintf("[%s]", a[arg_item])
	}
}

type twoJSON struct {
	m map[string]*[]sliceJSON // two key map "json-arg" and "json-file"
}

func newTwoJSON() *twoJSON {
	return &twoJSON{
		m: make(map[string]*[]sliceJSON),
	}
}

type sliceJSON struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func (tj *twoJSON) unmarshalMe(files [2][]byte) error {
	for i, k := range []string{"json-arg", "json-file"} {
		err := tj.unmarshal(k, files[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (tj *twoJSON) unmarshal(key string, file []byte) error {
	if len(file) == 0 {
		return nil
	}
	sj := []sliceJSON{}
	err := json.Unmarshal(file, &sj)
	if err != nil {
		return err
	}
	tj.m[key] = &sj
	return nil
}

func (tj twoJSON) ids() []string {
	var i []string
	for _, h := range *(tj.m["json-arg"]) { // 0 - "json-arg"
		i = append(i, h.Id)
	}
	return i
}

func (tj twoJSON) isExist(ids []string) (string, bool) {
	if _, ok := tj.m["json-file"]; !ok { // file not exist
		return "", false
	}
	for _, i := range ids {
		for _, g := range *(tj.m["json-file"]) { // 1 - "json-file"
			if i == g.Id {
				return g.Id, true
			}
		}
	}
	return "", false
}

func (tj twoJSON) find(id string) ([]byte, error) {
	if len(tj.m) == 0 {
		return nil, nil
	}
	// 1 - "json-file"
	for _, g := range *(tj.m["json-file"]) {
		if g.Id == id {
			return json.Marshal(g)
		}
	}
	return nil, nil
}

func (tj twoJSON) remove(id string) ([]byte, error) {
	if len(tj.m) == 0 {
		return nil, nil
	}
	s := []*sliceJSON{}
	// 1 - "json-file"
	for _, g := range *(tj.m["json-file"]) {
		if g.Id != id {
			s = append(s, &g)
			// buf, err := json.Marshal(g)
			// if err != nil {
			// 	return nil, err
			// }
			// b.WriteString(string(buf))
		}
	}
	return json.Marshal(s)
}

func parseArgs() Arguments {
	a := new(args)
	if a.fixArgs() {
		return a.parse()
	}
	return nil
}

func main() {
	// os.Args = append(os.Args, `-operation «add» -item ‘{«id»: "1", «email»: «email@test.com», «age»: 23}’ --fileName «users.json»`)
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
