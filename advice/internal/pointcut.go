package internal

import (
	"fmt"
	"github.com/wesovilabs/goa/logger"
	"regexp"
	"strings"
)

const exprPkg = `[a-zA-Z0-9_*\/]+\.`
const exprObj = `[a-zA-Z0-9_*]+\.`
const exprFunc = "[a-zA-Z0-9_*]+"
const exprArgs = `[a-zA-Z0-9_*,.{}()[\]/]+`

var regExp = func() *regexp.Regexp {
	expr := `^`
	expr += fmt.Sprintf(`(?P<pkg>%s)?`, exprPkg)
	expr += fmt.Sprintf(`(?P<obj>%s)?`, exprObj)
	expr += fmt.Sprintf(`(?P<func>%s)`, exprFunc)
	expr += fmt.Sprintf(`(?P<args>\(%s)`, exprArgs)
	expr += `$`
	return regexp.MustCompile(expr)
}()

// NormalizePointcut normalizes the provided expressions
func NormalizePointcut(text string) *regexp.Regexp {
	items := regExp.FindStringSubmatch(text)
	if len(items) != 5 {
		return nil
	}

	regExpStr := "^"
	if items[1] != "" {
		regExpStr += fmt.Sprintf(`%s\.`, processPkg(items[1]))
	}

	if items[2] != "" {
		regExpStr += fmt.Sprintf(`%s\.`, processObj(items[2]))
	}

	regExpStr += processFunc(items[3])
	in, out := processArgsInOut(items[4])
	_, i := processArgs(in)
	regExpStr += fmt.Sprintf(`\(%s\)`, i)
	total, res := processArgs(out)

	if total <= 1 {
		regExpStr += res
	} else {
		regExpStr += fmt.Sprintf(`\(%s\)`, res)
	}

	regExpStr += `$`
	rg, err := regexp.Compile(regExpStr)

	if err != nil {
		logger.Errorf("error processing `%s: %s", text, err.Error())
		return nil
	}

	logger.Infof("[ aspect ] %s => %s", text, regExpStr)

	return rg
}

func processPkg(text string) string {
	out := text[:len(text)-1]
	out = strings.ReplaceAll(out, `*`, `[a-zA-Z0-9_/]*`)
	out = strings.ReplaceAll(out, `/`, `\/`)

	return out
}

func processObj(text string) string {
	out := text[:len(text)-1]
	out = strings.ReplaceAll(out, `*`, `[a-zA-Z0-9_*]*`)

	return out
}

func processFunc(text string) string {
	out := strings.ReplaceAll(text, `*`, `[a-zA-Z0-9_]*`)

	return out
}

func processArgsInOut(text string) (string, string) {
	inStart := 1
	inEnd := 1
	openParen := 1

	for i := 1; i < len(text); i++ {
		switch text[i] {
		case '(':
			openParen++
		case ')':
			openParen--
		}

		if openParen == 0 {
			inEnd = i
			break
		}
	}

	in := text[inStart:inEnd]

	if len(text) == inEnd+1 {
		return in, ""
	}

	if text[inEnd+1] == '(' && text[len(text)-1] == ')' {
		return in, text[inEnd+2 : len(text)-1]
	}

	return text[inStart:inEnd], text[inEnd+1:]
}

func processArgs(text string) (int, string) {
	openParen := 0
	lastIndex := 0
	total := 0
	out := ""

	for index, c := range text {
		switch c {
		case '(':
			openParen++
		case ')':
			openParen--
		case ',':
			if openParen == 0 {
				out += fmt.Sprintf(`%s\,`, processArg(text[lastIndex:index]))
				lastIndex = index + 1
				total++
			}
		}
	}

	out += processArg(text[lastIndex:])
	total++

	return total, out
}

func processArg(text string) string {
	if ok, out := replaceSpecialExprInArg(text); ok {
		return out
	}

	return replaceSpecialCharsInArg(text)
}

func replaceSpecialExprInArg(text string) (bool, string) {
	if strings.HasPrefix(text, "func(") {
		args := strings.TrimPrefix(text, "func")
		in, out := processArgsInOut(args)
		_, i := processArgs(in)
		totalOut, o := processArgs(out)

		if totalOut <= 1 {
			return true, fmt.Sprintf(`func\(%s\)%s`, i, o)
		}

		return true, fmt.Sprintf(`func\(%s\)\(%s\)`, i, o)
	}

	if text == `*` {
		return true, `([a-zA-Z0-9_*.\[\]{}()\/]+|func\(.*\)\(.*\)|func\(.*\))`
	}

	if len(text) > 1 && text[0] == '*' {
		return true, fmt.Sprintf(`\*%s`, processArg(text[1:]))
	}

	if text == `...` {
		return true, `.*`
	}

	return false, text
}

func replaceSpecialCharsInArg(text string) string {
	text = strings.ReplaceAll(text, "[", `\[`)
	text = strings.ReplaceAll(text, "]", `\]`)
	text = strings.ReplaceAll(text, "*", `\*`)
	text = strings.ReplaceAll(text, ".", `\.`)

	return text
}
