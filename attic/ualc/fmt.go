package fmt

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	
	"ualcompiler/con"
)

// Printf produces a string based on format and extra arguments, then immediately prints it.
// Supported placeholders:
//   %d -> decimal integer
//   %x -> hexadecimal integer
//   %s -> string
func Printf(format string, args ...interface{}) int {
	result := Sprintf(format, args...)
	con.Print(result)
	return 0
}

// Sprintf produces a string based on format and extra arguments, returning the result.
// The same placeholders apply as in Printf.
func Sprintf(format string, args ...interface{}) string {
	// Simple placeholder pattern: %d, %x, %s
	pattern := regexp.MustCompile(`%[dxs]`)
	
	argIndex := 0
	result := pattern.ReplaceAllStringFunc(format, func(match string) string {
		if argIndex >= len(args) {
			return match // Not enough arguments, leave the placeholder
		}
		
		arg := args[argIndex]
		argIndex++
		
		switch match {
		case "%d":
			// Convert to decimal
			if i, ok := arg.(int); ok {
				return strconv.Itoa(i)
			}
			return "<?>"
			
		case "%x":
			// Convert to hexadecimal
			if i, ok := arg.(int); ok {
				return strings.ToLower(strconv.FormatInt(int64(i), 16))
			}
			return "<?>"
			
		case "%s":
			// Convert to string
			return fmt.Sprintf("%v", arg)
			
		default:
			return match
		}
	})
	
	return result
}