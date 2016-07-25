package definition

import (
	"html"
	"log"
	"math"
	"strings"
)

// Map a name to a function which matches the filterFunc definition
var filters = map[string]filterFunc{
	// Trims whitespace
	"trim": func(str string) interface{} { return strings.TrimSpace(str) },

	// Unescapes a HTML string
	"unescape": func(str string) interface{} { return html.UnescapeString(str) },

	// Adds space before capitals
	"respace": func(str string) interface{} {
		i := 1
		for {
			log.Println(str)
			if i >= len(str)-1 {
				break
			}
			// If it's a capital
			// If it's an ampersand
			// If it's a multiplication
			if (str[i] >= 'A' && str[i] <= 'Z') ||
				(str[i] == '&') ||
				(str[i] == 'x' && (str[i+1] >= '0' && str[i+1] <= '9')) {

				// Add a space
				str = str[:i] + " " + str[i:]
				i++
			}
			i++
		}
		return str
	},

	"lowercase": func(str string) interface{} { return strings.ToLower(str) },
	"uppercase": func(str string) interface{} { return strings.ToUpper(str) },

	"pence": func(str string) interface{} {
		pennies := 0
		unit := 0.0
		for i := len(str) - 1; i >= 0; i-- {
			if str[i] >= '0' && str[i] <= '9' {
				pennies += int(str[i]-'0') * int(math.Pow(10, unit))
				unit++
			}
		}
		return pennies
	},
}
