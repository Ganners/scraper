package definition

// A problem with traditional scrapers is that they parse HTML by converting it
// into DOM objects which are then queried. This is very powerful, but is
// difficult to configure and is a step removed from the problem.
//
// The idea here is that we will have definition files, these files will be the
// reverse of templates and will be treated as repeating objects, and will have
// some filters which can perform some common operations.
//
// To keep it dynamic, the output will be a slice of map[string]string
//
// Definition:
//  - White space will be skipped
//  - Comments will be skipped
//  - The HTML is just text here, invalid HTML will work fine because of this.
//    It will have to match exactly. Where it can be skipped, the syntax `{{_}}`
//    can be used.
//  - Variables are defined by the definition `{{variableName}}`, optionally
//    they can be overloaded with filters, e.g. `{{variableName|filter1|filter2}}`
//  - The filters that will be available are:
//    - Escape (perform url.QueryEscape)
//    - Trim (will perform strings.TrimSpace)
