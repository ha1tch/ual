package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// --- Conversion Helper ---
func convertValue(srcType string, value interface{}, targetType string) (interface{}, error) {
	if srcType == targetType {
		return value, nil
	}
	switch srcType {
	case "int":
		intVal := value.(int)
		switch targetType {
		case "str":
			return strconv.Itoa(intVal), nil
		case "float":
			return float64(intVal), nil
		}
	case "float":
		floatVal := value.(float64)
		switch targetType {
		case "str":
			return strconv.FormatFloat(floatVal, 'f', -1, 64), nil
		case "int":
			return int(floatVal), nil
		}
	case "str":
		strVal := value.(string)
		switch targetType {
		case "int":
			i, err := strconv.Atoi(strVal)
			if err != nil {
				return nil, err
			}
			return i, nil
		case "float":
			f, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				return nil, err
			}
			return f, nil
		}
	}
	return nil, fmt.Errorf("unsupported conversion from %s to %s", srcType, targetType)
}

// --- Global Variables ---
var globalGManager *GoroutineManager
var globalStrStacks map[string]*StringStack

// --- Spawn Stack (Goroutine Manager) ---

type ManagedGoroutine struct {
	name       string
	pauseChan  chan struct{}
	resumeChan chan struct{}
	stopChan   chan struct{}
	msgChan    chan string
	script     string // holds script instructions
	wg         *sync.WaitGroup
}

func NewManagedGoroutine(name string, wg *sync.WaitGroup) *ManagedGoroutine {
	return &ManagedGoroutine{
		name:       name,
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
		stopChan:   make(chan struct{}),
		msgChan:    make(chan string),
		script:     "",
		wg:         wg,
	}
}

func (mg *ManagedGoroutine) Run() {
	mg.wg.Add(1)
	go func() {
		defer mg.wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		running := true
		paused := false
		for running {
			select {
			case <-mg.stopChan:
				fmt.Printf("[%s] Stopping\n", mg.name)
				running = false
			case msg := <-mg.msgChan:
				// If the spawn receives a multi-line script, store and execute it.
				if mg.name == "spawn" && strings.Contains(msg, "\n") {
					mg.script = msg
					mg.ExecuteScript()
				} else {
					fmt.Printf("[%s] Received message: %s\n", mg.name, msg)
				}
			case <-mg.pauseChan:
				if !paused {
					paused = true
					fmt.Printf("[%s] Paused\n", mg.name)
				}
				<-mg.resumeChan
				paused = false
				fmt.Printf("[%s] Resumed\n", mg.name)
			case <-ticker.C:
				if !paused {
					fmt.Printf("[%s] Working...\n", mg.name)
				}
			}
		}
	}()
}

func (mg *ManagedGoroutine) ExecuteScript() {
	fmt.Printf("[%s] Executing script:\n%s\n", mg.name, mg.script)
	// Split script into lines and execute each line
	lines := strings.Split(mg.script, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		executeSpawnCommand(line)
	}
}

func (mg *ManagedGoroutine) Pause()  { mg.pauseChan <- struct{}{} }
func (mg *ManagedGoroutine) Resume() { mg.resumeChan <- struct{}{} }
func (mg *ManagedGoroutine) Stop()   { close(mg.stopChan) }
func (mg *ManagedGoroutine) SendMessage(msg string) {
	mg.msgChan <- msg
}

type GoroutineManager struct {
	stack []*ManagedGoroutine
	wg    sync.WaitGroup
	lock  sync.Mutex
}

func NewGoroutineManager() *GoroutineManager {
	return &GoroutineManager{
		stack: make([]*ManagedGoroutine, 0),
	}
}

func (gm *GoroutineManager) AddGoroutine(name string) {
	gm.lock.Lock()
	defer gm.lock.Unlock()
	mg := NewManagedGoroutine(name, &gm.wg)
	mg.Run()
	gm.stack = append(gm.stack, mg)
	fmt.Printf("Added goroutine '%s'\n", name)
}

func (gm *GoroutineManager) FindGoroutine(name string) *ManagedGoroutine {
	gm.lock.Lock()
	defer gm.lock.Unlock()
	for _, mg := range gm.stack {
		if mg.name == name {
			return mg
		}
	}
	return nil
}

func (gm *GoroutineManager) List() {
	gm.lock.Lock()
	defer gm.lock.Unlock()
	fmt.Println("Spawn Stack (Managed Goroutines):")
	for i, mg := range gm.stack {
		fmt.Printf("%d: %s\n", i, mg.name)
	}
}

func (gm *GoroutineManager) PauseGoroutine(name string) {
	if mg := gm.FindGoroutine(name); mg != nil {
		mg.Pause()
	} else {
		fmt.Printf("No goroutine found with name '%s'\n", name)
	}
}

func (gm *GoroutineManager) ResumeGoroutine(name string) {
	if mg := gm.FindGoroutine(name); mg != nil {
		mg.Resume()
	} else {
		fmt.Printf("No goroutine found with name '%s'\n", name)
	}
}

func (gm *GoroutineManager) StopGoroutine(name string) {
	if mg := gm.FindGoroutine(name); mg != nil {
		mg.Stop()
	} else {
		fmt.Printf("No goroutine found with name '%s'\n", name)
	}
}

func (gm *GoroutineManager) StopAll() {
	gm.lock.Lock()
	defer gm.lock.Unlock()
	for _, mg := range gm.stack {
		mg.Stop()
	}
}

func (gm *GoroutineManager) SendMessageToGoroutine(name, msg string) {
	if mg := gm.FindGoroutine(name); mg != nil {
		mg.SendMessage(msg)
		fmt.Printf("Sent message to '%s'\n", name)
	} else {
		fmt.Printf("No goroutine found with name '%s'\n", name)
	}
}

// --- Forth‑Style Dynamic Stacks ---
// Each stack has a mode ("lifo" default or "fifo") and supports a flip operation.

type IntStack struct {
	data []int
	mode string
}

func NewIntStack() *IntStack {
	return &IntStack{data: []int{}, mode: "lifo"}
}

func (s *IntStack) Push(val int) {
	s.data = append(s.data, val)
}

func (s *IntStack) Pop() (int, bool) {
	if len(s.data) == 0 {
		return 0, false
	}
	if s.mode == "fifo" {
		val := s.data[0]
		s.data = s.data[1:]
		return val, true
	}
	val := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return val, true
}

func (s *IntStack) Dup() bool {
	if len(s.data) == 0 {
		return false
	}
	top := s.data[len(s.data)-1]
	s.Push(top)
	return true
}

func (s *IntStack) Swap() bool {
	if len(s.data) < 2 {
		return false
	}
	s.data[len(s.data)-1], s.data[len(s.data)-2] = s.data[len(s.data)-2], s.data[len(s.data)-1]
	return true
}

func (s *IntStack) Drop() bool {
	_, ok := s.Pop()
	return ok
}

func (s *IntStack) Print() {
	fmt.Printf("IntStack (%s mode): %v\n", s.mode, s.data)
}

func (s *IntStack) Add() bool {
	if len(s.data) < 2 {
		return false
	}
	b, _ := s.Pop()
	a, _ := s.Pop()
	s.Push(a + b)
	return true
}

func (s *IntStack) Sub() bool {
	if len(s.data) < 2 {
		return false
	}
	b, _ := s.Pop()
	a, _ := s.Pop()
	s.Push(a - b)
	return true
}

func (s *IntStack) Mul() bool {
	if len(s.data) < 2 {
		return false
	}
	b, _ := s.Pop()
	a, _ := s.Pop()
	s.Push(a * b)
	return true
}

func (s *IntStack) Div() bool {
	if len(s.data) < 2 {
		return false
	}
	b, _ := s.Pop()
	if b == 0 {
		fmt.Println("Division by zero")
		s.Push(b)
		return false
	}
	a, _ := s.Pop()
	s.Push(a / b)
	return true
}

func (s *IntStack) SetMode(mode string) {
	if mode == "lifo" || mode == "fifo" {
		s.mode = mode
	}
}

func (s *IntStack) Flip() {
	for i, j := 0, len(s.data)-1; i < j; i, j = i+1, j-1 {
		s.data[i], s.data[j] = s.data[j], s.data[i]
	}
}

type StringStack struct {
	data []string
	mode string
}

func NewStringStack() *StringStack {
	return &StringStack{data: []string{}, mode: "lifo"}
}

func (s *StringStack) Push(val string) {
	s.data = append(s.data, val)
}

func (s *StringStack) Pop() (string, bool) {
	if len(s.data) == 0 {
		return "", false
	}
	if s.mode == "fifo" {
		val := s.data[0]
		s.data = s.data[1:]
		return val, true
	}
	val := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return val, true
}

func (s *StringStack) Dup() bool {
	if len(s.data) == 0 {
		return false
	}
	top := s.data[len(s.data)-1]
	s.Push(top)
	return true
}

func (s *StringStack) Swap() bool {
	if len(s.data) < 2 {
		return false
	}
	s.data[len(s.data)-1], s.data[len(s.data)-2] = s.data[len(s.data)-2], s.data[len(s.data)-1]
	return true
}

func (s *StringStack) Drop() bool {
	_, ok := s.Pop()
	return ok
}

func (s *StringStack) Print() {
	fmt.Printf("StringStack (%s mode): %v\n", s.mode, s.data)
}

func (s *StringStack) Add() bool {
	if len(s.data) < 2 {
		return false
	}
	b, _ := s.Pop()
	a, _ := s.Pop()
	s.Push(a + b)
	return true
}

// sub <trimChar>: remove trailing occurrences of the given character.
func (s *StringStack) Sub(trimChar string) bool {
	if len(s.data) == 0 {
		return false
	}
	top, _ := s.Pop()
	s.Push(strings.TrimRight(top, trimChar))
	return true
}

// mul <n>: replicates the string n times.
func (s *StringStack) Mul(n int) bool {
	if len(s.data) == 0 {
		return false
	}
	str, _ := s.Pop()
	s.Push(strings.Repeat(str, n))
	return true
}

// div <delim>: splits the string by the delimiter and joins with a space.
func (s *StringStack) Div(delim string) bool {
	if len(s.data) == 0 {
		return false
	}
	str, _ := s.Pop()
	parts := strings.Split(str, delim)
	s.Push(strings.Join(parts, " "))
	return true
}

func (s *StringStack) SetMode(mode string) {
	if mode == "lifo" || mode == "fifo" {
		s.mode = mode
	}
}

func (s *StringStack) Flip() {
	for i, j := 0, len(s.data)-1; i < j; i, j = i+1, j-1 {
		s.data[i], s.data[j] = s.data[j], s.data[i]
	}
}

type FloatStack struct {
	data []float64
	mode string
}

func NewFloatStack() *FloatStack {
	return &FloatStack{data: []float64{}, mode: "lifo"}
}

func (s *FloatStack) Push(val float64) {
	s.data = append(s.data, val)
}

func (s *FloatStack) Pop() (float64, bool) {
	if len(s.data) == 0 {
		return 0, false
	}
	if s.mode == "fifo" {
		val := s.data[0]
		s.data = s.data[1:]
		return val, true
	}
	val := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return val, true
}

func (s *FloatStack) Dup() bool {
	if len(s.data) == 0 {
		return false
	}
	top := s.data[len(s.data)-1]
	s.Push(top)
	return true
}

func (s *FloatStack) Swap() bool {
	if len(s.data) < 2 {
		return false
	}
	s.data[len(s.data)-1], s.data[len(s.data)-2] = s.data[len(s.data)-2], s.data[len(s.data)-1]
	return true
}

func (s *FloatStack) Drop() bool {
	_, ok := s.Pop()
	return ok
}

func (s *FloatStack) Print() {
	fmt.Printf("FloatStack (%s mode): %v\n", s.mode, s.data)
}

func (s *FloatStack) Add() bool {
	if len(s.data) < 2 {
		return false
	}
	b, _ := s.Pop()
	a, _ := s.Pop()
	s.Push(a + b)
	return true
}

func (s *FloatStack) Sub() bool {
	if len(s.data) < 2 {
		return false
	}
	b, _ := s.Pop()
	a, _ := s.Pop()
	s.Push(a - b)
	return true
}

func (s *FloatStack) Mul() bool {
	if len(s.data) < 2 {
		return false
	}
	b, _ := s.Pop()
	a, _ := s.Pop()
	s.Push(a * b)
	return true
}

func (s *FloatStack) Div() bool {
	if len(s.data) < 2 {
		return false
	}
	b, _ := s.Pop()
	if b == 0 {
		fmt.Println("Division by zero")
		s.Push(b)
		return false
	}
	a, _ := s.Pop()
	s.Push(a / b)
	return true
}

func (s *FloatStack) SetMode(mode string) {
	if mode == "lifo" || mode == "fifo" {
		s.mode = mode
	}
}

func (s *FloatStack) Flip() {
	for i, j := 0, len(s.data)-1; i < j; i, j = i+1, j-1 {
		s.data[i], s.data[j] = s.data[j], s.data[i]
	}
}

// --- Stack Selector ---
type StackSelector struct {
	name string
	typ  string // "int", "str", "float", or "spawn"
}

// --- Spawn Script Executor ---
// This function interprets a spawn command (used when executing a script).
func executeSpawnCommand(cmd string) {
	tokens := strings.Fields(cmd)
	if len(tokens) == 0 {
		return
	}
	op := strings.ToLower(tokens[0])
	switch op {
	case "list":
		globalGManager.List()
	case "add":
		if len(tokens) < 2 {
			fmt.Println("add requires a goroutine name")
		} else {
			globalGManager.AddGoroutine(tokens[1])
		}
	case "pause":
		if len(tokens) < 2 {
			fmt.Println("pause requires a goroutine name")
		} else {
			globalGManager.PauseGoroutine(tokens[1])
		}
	case "resume":
		if len(tokens) < 2 {
			fmt.Println("resume requires a goroutine name")
		} else {
			globalGManager.ResumeGoroutine(tokens[1])
		}
	case "stop":
		if len(tokens) < 2 {
			fmt.Println("stop requires a goroutine name")
		} else {
			globalGManager.StopGoroutine(tokens[1])
		}
	default:
		fmt.Println("Unknown spawn command:", op)
	}
}

// --- Main CLI ---
// At startup we create two default stacks: dstack and rstack (as int stacks).
// The spawn stack is always available as "spawn".
func main() {
	gManager := NewGoroutineManager()
	globalGManager = gManager

	// Dynamic stacks.
	intStacks := make(map[string]*IntStack)
	strStacks := make(map[string]*StringStack)
	floatStacks := make(map[string]*FloatStack)
	globalStrStacks = make(map[string]*StringStack)

	// Create default Forth stacks.
	intStacks["dstack"] = NewIntStack() // data stack
	intStacks["rstack"] = NewIntStack() // return stack

	var currentSelector *StackSelector

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("iual 0.0.1")
	fmt.Println("An exceedingly trivial interactive ual 0.0.1 interpreter")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  Spawn Stack Commands (active only when @spawn is selected):")
	fmt.Println("    list, add <name>, pause <name>, resume <name>, stop <name>, bring, run")
	fmt.Println("  Create new stack: new <stack name> <int|str|float>")
	fmt.Println("  Stack selector: @<stack name>  (e.g., @dstack, @rstack, or @spawn)")
	fmt.Println("  Compound commands (selector followed by colon):")
	fmt.Println("       @dstack: push:1 pop mul")
	fmt.Println("       @dstack: div(10,2)")
	fmt.Println("       @spawn: bring(str,@sstack) run")
	fmt.Println("    Tokens may be simple (e.g., push), with colon (push:1), or function-like (div(10,2)).")
	fmt.Println("  For int stacks: available ops: push, pop, dup, swap, drop, print, add, sub, mul, div, lifo, fifo, flip, bring <srcType>,<srcStack>")
	fmt.Println("  For string stacks: available ops: push, pop, dup, swap, drop, print, add, sub <char>, mul <n>, div <delim>, lifo, fifo, flip, bring <srcType>,<srcStack>")
	fmt.Println("  For float stacks: similar to int stacks.")
	fmt.Println("  Explicit stack ops: int|str|float <op> <stack name> [value]")
	fmt.Println("  Send from stack: send <int|str|float> <stack name> <goroutine>")
	fmt.Println("  quit")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Compound command: starts with "@" and contains ":".
		if strings.HasPrefix(input, "@") && strings.Contains(input, ":") {
			parts := strings.SplitN(input, ":", 2)
			selectorPart := parts[0] // e.g., "@dstack"
			compoundOps := strings.TrimSpace(parts[1])
			selName := selectorPart[1:]
			var selType string
			if selName == "spawn" {
				selType = "spawn"
			} else if _, ok := intStacks[selName]; ok {
				selType = "int"
			} else if _, ok := strStacks[selName]; ok {
				selType = "str"
			} else if _, ok := floatStacks[selName]; ok {
				selType = "float"
			} else {
				fmt.Printf("No stack with name '%s' found\n", selName)
				continue
			}
			currentSelector = &StackSelector{name: selName, typ: selType}
			fmt.Printf("Stack selector set to '%s' of type %s\n", selName, selType)
			tokens := strings.Fields(compoundOps)
			for _, token := range tokens {
				// Function-like syntax: op(arg1,arg2,...)
				if strings.Contains(token, "(") && strings.HasSuffix(token, ")") {
					opParts := strings.SplitN(token, "(", 2)
					funcOp := strings.ToLower(opParts[0])
					argList := strings.TrimSuffix(opParts[1], ")")
					argTokens := strings.Split(argList, ",")
					switch currentSelector.typ {
					case "int":
						stack := intStacks[currentSelector.name]
						for _, a := range argTokens {
							a = strings.TrimSpace(a)
							val, err := strconv.Atoi(a)
							if err != nil {
								fmt.Println("Invalid int argument:", a)
								continue
							}
							stack.Push(val)
						}
					case "float":
						stack := floatStacks[currentSelector.name]
						for _, a := range argTokens {
							a = strings.TrimSpace(a)
							val, err := strconv.ParseFloat(a, 64)
							if err != nil {
								fmt.Println("Invalid float argument:", a)
								continue
							}
							stack.Push(val)
						}
					case "str":
						stack := strStacks[currentSelector.name]
						for _, a := range argTokens {
							a = strings.TrimSpace(a)
							a = strings.Trim(a, "\"")
							stack.Push(a)
						}
					}
					token = strings.ToLower(opParts[0])
				}

				// Check for colon syntax: op:arg
				var opName, opArg string
				if strings.Contains(token, ":") {
					subParts := strings.SplitN(token, ":", 2)
					opName = strings.ToLower(subParts[0])
					opArg = subParts[1]
				} else {
					opName = strings.ToLower(token)
				}

				switch currentSelector.typ {
				case "int":
					stack := intStacks[currentSelector.name]
					switch opName {
					case "push":
						if opArg == "" {
							fmt.Println("push requires an argument")
							continue
						}
						val, err := strconv.Atoi(opArg)
						if err != nil {
							fmt.Println("Invalid int:", opArg)
							continue
						}
						stack.Push(val)
					case "pop":
						if val, ok := stack.Pop(); ok {
							fmt.Println("Popped:", val)
						} else {
							fmt.Println("Stack is empty")
						}
					case "dup":
						if !stack.Dup() {
							fmt.Println("Cannot duplicate: stack is empty")
						}
					case "swap":
						if !stack.Swap() {
							fmt.Println("Cannot swap: less than 2 elements")
						}
					case "drop":
						if !stack.Drop() {
							fmt.Println("Cannot drop: stack is empty")
						}
					case "print":
						stack.Print()
					case "add":
						if !stack.Add() {
							fmt.Println("Not enough elements for addition")
						}
					case "sub":
						if !stack.Sub() {
							fmt.Println("Not enough elements for subtraction")
						}
					case "mul":
						if !stack.Mul() {
							fmt.Println("Not enough elements for multiplication")
						}
					case "div":
						if !stack.Div() {
							fmt.Println("Not enough elements for division or division by zero")
						}
					case "lifo":
						stack.SetMode("lifo")
						fmt.Println("Set mode to lifo")
					case "fifo":
						stack.SetMode("fifo")
						fmt.Println("Set mode to fifo")
					case "flip":
						stack.Flip()
						fmt.Println("Stack flipped")
					case "bring":
						if opArg == "" {
							fmt.Println("bring requires argument in form <srcType>,<srcStack>")
							continue
						}
						params := strings.Split(opArg, ",")
						if len(params) < 2 {
							fmt.Println("bring requires two parameters: srcType and srcStack")
							continue
						}
						srcType := strings.ToLower(strings.TrimSpace(params[0]))
						srcStackName := strings.TrimSpace(params[1])
						if strings.HasPrefix(srcStackName, "@") {
							srcStackName = srcStackName[1:]
						}
						targetType := currentSelector.typ
						var value interface{}
						var ok bool
						switch srcType {
						case "int":
							srcStack, exists := intStacks[srcStackName]
							if !exists {
								fmt.Printf("No int stack named '%s'\n", srcStackName)
								continue
							}
							value, ok = srcStack.Pop()
							if !ok {
								fmt.Println("Source int stack is empty")
								continue
							}
						case "str":
							srcStack, exists := strStacks[srcStackName]
							if !exists {
								fmt.Printf("No string stack named '%s'\n", srcStackName)
								continue
							}
							value, ok = srcStack.Pop()
							if !ok {
								fmt.Println("Source string stack is empty")
								continue
							}
						case "float":
							srcStack, exists := floatStacks[srcStackName]
							if !exists {
								fmt.Printf("No float stack named '%s'\n", srcStackName)
								continue
							}
							value, ok = srcStack.Pop()
							if !ok {
								fmt.Println("Source float stack is empty")
								continue
							}
						default:
							fmt.Println("Unknown source type. Use int, str, or float.")
							continue
						}
						converted, err := convertValue(srcType, value, targetType)
						if err != nil {
							fmt.Println("Conversion error:", err)
							continue
						}
						switch targetType {
						case "int":
							stack.Push(converted.(int))
						}
						fmt.Printf("Brought value from %s stack '%s' to selected %s stack '%s'\n", srcType, srcStackName, targetType, currentSelector.name)
					default:
						fmt.Println("Unknown operation:", opName)
					}
				case "str":
					stack := strStacks[currentSelector.name]
					switch opName {
					case "push":
						if opArg == "" {
							fmt.Println("push requires an argument")
							continue
						}
						// Remove surrounding quotes if present.
						val := strings.Trim(opArg, "\"")
						stack.Push(val)
					case "pop":
						if val, ok := stack.Pop(); ok {
							fmt.Println("Popped:", val)
						} else {
							fmt.Println("Stack is empty")
						}
					case "dup":
						if !stack.Dup() {
							fmt.Println("Cannot duplicate: stack is empty")
						}
					case "swap":
						if !stack.Swap() {
							fmt.Println("Cannot swap: less than 2 elements")
						}
					case "drop":
						if !stack.Drop() {
							fmt.Println("Cannot drop: stack is empty")
						}
					case "print":
						stack.Print()
					case "add":
						if !stack.Add() {
							fmt.Println("Not enough elements for concatenation")
						}
					case "sub":
						if opArg == "" {
							fmt.Println("sub requires an argument (character to trim)")
							continue
						}
						if !stack.Sub(opArg) {
							fmt.Println("Sub operation failed")
						}
					case "mul":
						if opArg == "" {
							fmt.Println("mul requires an argument")
							continue
						}
						n, err := strconv.Atoi(opArg)
						if err != nil {
							fmt.Println("Invalid multiplier:", opArg)
							continue
						}
						if !stack.Mul(n) {
							fmt.Println("Mul operation failed")
						}
					case "div":
						if opArg == "" {
							fmt.Println("div requires an argument (delimiter)")
							continue
						}
						if !stack.Div(opArg) {
							fmt.Println("Div operation failed")
						}
					case "lifo":
						stack.SetMode("lifo")
						fmt.Println("Set mode to lifo")
					case "fifo":
						stack.SetMode("fifo")
						fmt.Println("Set mode to fifo")
					case "flip":
						stack.Flip()
						fmt.Println("Stack flipped")
					case "bring":
						if opArg == "" {
							fmt.Println("bring requires argument in form <srcType>,<srcStack>")
							continue
						}
						params := strings.Split(opArg, ",")
						if len(params) < 2 {
							fmt.Println("bring requires two parameters: srcType and srcStack")
							continue
						}
						srcType := strings.ToLower(strings.TrimSpace(params[0]))
						srcStackName := strings.TrimSpace(params[1])
						if strings.HasPrefix(srcStackName, "@") {
							srcStackName = srcStackName[1:]
						}
						targetType := currentSelector.typ
						var value interface{}
						var ok bool
						switch srcType {
						case "int":
							srcStack, exists := intStacks[srcStackName]
							if !exists {
								fmt.Printf("No int stack named '%s'\n", srcStackName)
								continue
							}
							value, ok = srcStack.Pop()
							if !ok {
								fmt.Println("Source int stack is empty")
								continue
							}
						case "str":
							srcStack, exists := strStacks[srcStackName]
							if !exists {
								fmt.Printf("No string stack named '%s'\n", srcStackName)
								continue
							}
							value, ok = srcStack.Pop()
							if !ok {
								fmt.Println("Source string stack is empty")
								continue
							}
						case "float":
							srcStack, exists := floatStacks[srcStackName]
							if !exists {
								fmt.Printf("No float stack named '%s'\n", srcStackName)
								continue
							}
							value, ok = srcStack.Pop()
							if !ok {
								fmt.Println("Source float stack is empty")
								continue
							}
						default:
							fmt.Println("Unknown source type. Use int, str, or float.")
							continue
						}
						converted, err := convertValue(srcType, value, targetType)
						if err != nil {
							fmt.Println("Conversion error:", err)
							continue
						}
						switch targetType {
						case "str":
							stack.Push(converted.(string))
						}
						fmt.Printf("Brought value from %s stack '%s' to selected %s stack '%s'\n", srcType, srcStackName, targetType, currentSelector.name)
					default:
						fmt.Println("Unknown operation:", opName)
					}
				case "spawn":
					switch opName {
					case "list":
						gManager.List()
					case "add":
						if opArg == "" {
							fmt.Println("add requires a goroutine name")
							continue
						}
						gManager.AddGoroutine(opArg)
					case "pause":
						if opArg == "" {
							fmt.Println("pause requires a goroutine name")
							continue
						}
						gManager.PauseGoroutine(opArg)
					case "resume":
						if opArg == "" {
							fmt.Println("resume requires a goroutine name")
							continue
						}
						gManager.ResumeGoroutine(opArg)
					case "stop":
						if opArg == "" {
							fmt.Println("stop requires a goroutine name")
							continue
						}
						gManager.StopGoroutine(opArg)
					case "bring":
						if opArg == "" {
							fmt.Println("bring requires argument in form <srcType>,<srcStack>")
							continue
						}
						params := strings.Split(opArg, ",")
						if len(params) < 2 {
							fmt.Println("bring requires two parameters: srcType and srcStack")
							continue
						}
						srcType := strings.ToLower(strings.TrimSpace(params[0]))
						srcStackName := strings.TrimSpace(params[1])
						if strings.HasPrefix(srcStackName, "@") {
							srcStackName = srcStackName[1:]
						}
						if srcType != "str" {
							fmt.Println("For spawn bring, only string scripts are supported.")
							continue
						}
						srcStack, exists := strStacks[srcStackName]
						if !exists {
							fmt.Printf("No string stack named '%s'\n", srcStackName)
							continue
						}
						var scriptLines []string
						for {
							instr, ok := srcStack.Pop()
							if !ok {
								break
							}
							scriptLines = append(scriptLines, instr)
						}
						// Reverse to preserve original order.
						for i, j := 0, len(scriptLines)-1; i < j; i, j = i+1, j-1 {
							scriptLines[i], scriptLines[j] = scriptLines[j], scriptLines[i]
						}
						fullScript := strings.Join(scriptLines, "\n")
						// Instead of executing immediately, store the script in the spawn.
						gManager.FindGoroutine(currentSelector.name).script = fullScript
						fmt.Printf("Script stored in spawn '%s'. Use run to execute.\n", currentSelector.name)
					case "run":
						gManager.FindGoroutine(currentSelector.name).ExecuteScript()
					default:
						fmt.Println("Unknown spawn operation:", opName)
					}
				default:
					fmt.Println("Compound commands for type", currentSelector.typ, "not implemented in this example.")
				}
			}
			continue
		}

		// Global commands and explicit operations.
		tokens := strings.Fields(input)
		if len(tokens) == 0 {
			continue
		}
		switch strings.ToLower(tokens[0]) {
		case "new":
			if len(tokens) < 3 {
				fmt.Println("Usage: new <stack name> <int|str|float>")
				continue
			}
			stackName := tokens[1]
			stackType := strings.ToLower(tokens[2])
			switch stackType {
			case "int":
				if _, exists := intStacks[stackName]; exists {
					fmt.Printf("Int stack '%s' already exists\n", stackName)
				} else {
					intStacks[stackName] = NewIntStack()
					fmt.Printf("Created new int stack '%s'\n", stackName)
				}
			case "str":
				if _, exists := strStacks[stackName]; exists {
					fmt.Printf("String stack '%s' already exists\n", stackName)
				} else {
					strStacks[stackName] = NewStringStack()
					globalStrStacks[stackName] = strStacks[stackName]
					fmt.Printf("Created new string stack '%s'\n", stackName)
				}
			case "float":
				if _, exists := floatStacks[stackName]; exists {
					fmt.Printf("Float stack '%s' already exists\n", stackName)
				} else {
					floatStacks[stackName] = NewFloatStack()
					fmt.Printf("Created new float stack '%s'\n", stackName)
				}
			default:
				fmt.Println("Unknown stack type. Use int, str, or float.")
			}
		case "spawn":
			if len(tokens) < 2 {
				fmt.Println("Usage: spawn <goroutine name>")
				continue
			}
			gManager.AddGoroutine(tokens[1])
		case "pause":
			if len(tokens) < 2 {
				fmt.Println("Usage: pause <goroutine name>")
				continue
			}
			gManager.PauseGoroutine(tokens[1])
		case "resume":
			if len(tokens) < 2 {
				fmt.Println("Usage: resume <goroutine name>")
				continue
			}
			gManager.ResumeGoroutine(tokens[1])
		case "stop":
			if len(tokens) < 2 {
				fmt.Println("Usage: stop <goroutine name>")
				continue
			}
			gManager.StopGoroutine(tokens[1])
		case "list":
			gManager.List()
		case "send":
			if len(tokens) < 4 {
				fmt.Println("Usage: send <int|str|float> <stack name> <goroutine>")
				continue
			}
			stackType := strings.ToLower(tokens[1])
			stackName := tokens[2]
			target := tokens[3]
			switch stackType {
			case "int":
				stack, exists := intStacks[stackName]
				if !exists {
					fmt.Printf("No int stack named '%s'\n", stackName)
					continue
				}
				val, ok := stack.Pop()
				if !ok {
					fmt.Println("Int stack is empty")
					continue
				}
				gManager.SendMessageToGoroutine(target, strconv.Itoa(val))
			default:
				fmt.Println("send: only int send implemented in this example.")
			}
		case "quit":
			fmt.Println("Stopping all spawn goroutines and exiting...")
			gManager.StopAll()
			gManager.wg.Wait()
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}
