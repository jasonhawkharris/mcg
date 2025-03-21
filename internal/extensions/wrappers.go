package extensions

import (
	"fmt"
	"reflect"
)

// ExtensionWrapper wraps any extension-like value to implement the Extension interface
type ExtensionWrapper struct {
	name        string
	description string
	value       interface{}
}

// Name returns the extension name
func (w *ExtensionWrapper) Name() string {
	return w.name
}

// Description returns the extension description
func (w *ExtensionWrapper) Description() string {
	return w.description
}

// Commands returns the extension commands
func (w *ExtensionWrapper) Commands() []Command {
	// We call Commands() on the wrapped value and convert the result
	result, err := callExtensionMethod(w.value, "Commands")
	if err != nil {
		return nil
	}
	
	// Convert the result to a slice of Commands
	cmdSlice := convertToInterfaceSlice(result)
	
	// Create a wrapper command for each command
	commands := make([]Command, 0, len(cmdSlice))
	for _, cmdValue := range cmdSlice {
		cmdName, err := callExtensionMethod(cmdValue, "Name")
		if err != nil {
			continue
		}
		
		cmdNameStr, ok := cmdName.(string)
		if !ok {
			continue
		}
		
		cmdDesc, err := callExtensionMethod(cmdValue, "Description")
		if err != nil {
			continue
		}
		
		cmdDescStr, ok := cmdDesc.(string)
		if !ok {
			continue
		}
		
		// Create a wrapper command
		commands = append(commands, &CommandWrapper{
			name:        cmdNameStr,
			description: cmdDescStr,
			value:       cmdValue,
		})
	}
	
	return commands
}

// CommandWrapper wraps any command-like value to implement the Command interface
type CommandWrapper struct {
	name        string
	description string
	value       interface{}
}

// Name returns the command name
func (w *CommandWrapper) Name() string {
	return w.name
}

// Description returns the command description
func (w *CommandWrapper) Description() string {
	return w.description
}

// Execute runs the command
func (w *CommandWrapper) Execute(args []string) (string, error) {
	// Call Execute on the wrapped value
	result, err := callExtensionMethodWithArgs(w.value, "Execute", []interface{}{args})
	if err != nil {
		return "", err
	}
	
	// Convert the result to a string
	resultStr, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("command's Execute() doesn't return a string")
	}
	
	return resultStr, nil
}

// callExtensionMethod calls a method on an extension or command using reflection
func callExtensionMethod(value interface{}, methodName string) (interface{}, error) {
	return callExtensionMethodWithArgs(value, methodName, nil)
}

// callExtensionMethodWithArgs calls a method on an extension or command with arguments
func callExtensionMethodWithArgs(value interface{}, methodName string, args []interface{}) (interface{}, error) {
	// Get the value's reflection value
	valueVal := reflect.ValueOf(value)
	
	// Get the method
	method := valueVal.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}
	
	// Prepare arguments
	var argVals []reflect.Value
	if args != nil {
		argVals = make([]reflect.Value, len(args))
		for i, arg := range args {
			argVals[i] = reflect.ValueOf(arg)
		}
	}
	
	// Call the method
	resultVals := method.Call(argVals)
	
	// Check for error
	if len(resultVals) >= 2 {
		errVal := resultVals[1]
		if !errVal.IsNil() {
			return nil, errVal.Interface().(error)
		}
	}
	
	// Return the result
	if len(resultVals) >= 1 {
		return resultVals[0].Interface(), nil
	}
	
	return nil, nil
}

// convertToInterfaceSlice converts any slice to a slice of interfaces
func convertToInterfaceSlice(value interface{}) []interface{} {
	// Get the value's reflection value
	valueVal := reflect.ValueOf(value)
	
	// If it's not a slice, return nil
	if valueVal.Kind() != reflect.Slice {
		return nil
	}
	
	// Convert the slice to a slice of interfaces
	length := valueVal.Len()
	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		result[i] = valueVal.Index(i).Interface()
	}
	
	return result
}