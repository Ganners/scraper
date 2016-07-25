package definition

import (
	"log"
	"net/url"
	"strings"
)

// Map a name to a function which matches the filterFunc definition
var filters = map[string]filterFunc{
	// Trims whitespace
	"trim": strings.TrimSpace,

	// Unescapes a URL
	"unescape": func(str string) string {
		str, err := url.QueryUnescape(str)
		if err != nil {
			return "UNESCAPE_ERROR"
		}
		return str
	},

	// Adds space before capitals
	"respace": func(str string) string {
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

	"lowercase": strings.ToLower,
	"uppercase": strings.ToUpper,
}
