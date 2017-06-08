package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	// Check whether or not a file was passed in; panic if no file is listed
	if len(os.Args) < 2 {
		panic("Filename not detected")
	}

	// Open the JSON file; if there is a file error, quit the program
	fileName := os.Args[1]
	jsonFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	tokenArray := getTokens(jsonFile) // Tokenize the JSON file
	printHeader()                     // Print the HTML header
	printTokens(tokenArray)           // Style and print each token
	printFooter()                     // Print the HTML footer
}

// Token carries a kind (which is an ID) and the content of the token
type Token struct {
	content string
	kind    int
}

// Token types are listed here for document readability, this idea taken
// from http://www.cs.sfu.ca/CourseCentral/383/tjd/syntaxAndEBNF.html
const (
	// Parenthesis token types: '{', '}', '[', ']'
	ObjectOpen  = 11
	ObjectClose = 12
	ArrayOpen   = 13
	ArrayClose  = 14

	// Delimiter token types: ':', ','
	DelimiterPair   = 21
	DelimiterMember = 22

	// String token types, either string, escaped string, or the special case
	// StringClose, which is for a single quote at the end of a string: '"'
	StringRegular = 31
	StringEscaped = 32
	StringClose   = 33

	// Number token type
	Number = 41

	// Literal token types: 'true', 'false', 'null'
	LiteralBoolTrue  = 51
	LiteralBoolFalse = 52
	LiteralNull      = 53
)

// getTokens returns an array of tokens from the file that is passed in
func getTokens(jsonFile []byte) []Token {
	tokenArray := make([]Token, 0) // In case the file is of 0 length

	// Iterate over every character in the file
	for i := 0; i < len(jsonFile); {
		currentCharacter := string(jsonFile[i])

		// These are the default token characteristics
		tokenContent := currentCharacter
		tokenKind := 0
		tokenLength := 1
		isToken := true

		// These booleans indicate complex tokens and tokens of variable length
		isString := false
		isStringRegular := false
		isNumber := false

		// Check the previous token to determine whether or not this token is
		// part of a previous string, which would restrict the possible types to
		// only StringRegular or StringEscaped
		if len(tokenArray) > 0 {
			previousToken := tokenArray[len(tokenArray)-1]
			previousKind := previousToken.kind
			previousContent := previousToken.content
			previousCharacter := string(previousContent[len(previousContent)-1])

			switch previousKind {
			case StringRegular:
				// If the last token was a string that ended with a quote, the
				// string is finished
				if previousCharacter != "\"" {
					isString = true
				}

				// If the previous string was the first character in the string,
				// making this token an escape character or the empty string
				if previousCharacter == "\"" && len(previousContent) == 1 {
					isString = true
				}
			case StringEscaped:
				isString = true

				// If this character is the final character in the string,
				// closing a string that ended with an escape character
				if currentCharacter == "\"" {
					tokenKind = StringClose
				}
			}
		}

		// Given that this token is not part of a previous string, we can
		// determine the type of this token by looking at a single character
		if !isString {
			switch currentCharacter {
			case "{":
				tokenKind = ObjectOpen
			case "}":
				tokenKind = ObjectClose
			case "[":
				tokenKind = ArrayOpen
			case "]":
				tokenKind = ArrayClose
			case ":":
				tokenKind = DelimiterPair
			case ",":
				tokenKind = DelimiterMember
			case "\"":
				tokenKind = StringRegular
				isStringRegular = true
			case "-", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				// All valid characters that indicate numbers in JSON. Numbers
				// cannot start with '.', '+', 'e', or 'E'
				tokenKind = Number
				isNumber = true
			case "t":
				// Literals must be in lowercase letters; does not include 'T'
				tokenKind = LiteralBoolTrue
				tokenContent = "true"
				tokenLength = 4
			case "f":
				// Literals must be in lowercase letters; does not include 'F'
				tokenKind = LiteralBoolFalse
				tokenContent = "false"
				tokenLength = 5
			case "n":
				// Literals must be in lowercase letters; does not include 'N'
				tokenKind = LiteralNull
				tokenContent = "null"
				tokenLength = 4
			default:
				// Ignore all whitespace and unreadable or invalid characters
				isToken = false
			}
		}

		// Given that this token is part of a previous string, we determine
		// whether the type of this string is StringRegular or StringEscaped
		if isString {
			switch currentCharacter {
			case "\"":
				if tokenKind != StringClose {
					tokenKind = StringRegular
					isStringRegular = true
				}
			case "\\":
				tokenKind = StringEscaped
				currentCharacter = string(jsonFile[i+1])

				if currentCharacter == "u" {
					// If the current escape character is \u followed by a 4 digit
					// hex string, we add those characters to the string
					for j := i + 1; j < (i + 6); j++ {
						currentCharacter = string(jsonFile[j])
						tokenContent += currentCharacter
						tokenLength++
					}
				} else {
					// If the current escape character is not \u, we add
					// whatever follows the '\' to our string. We benefit from
					// the valid input here.
					tokenContent += currentCharacter
					tokenLength++
				}
			default:
				tokenKind = StringRegular
				isStringRegular = true
			}
		}

		// Given that this token is a StringRegular, we add all digits until we
		// reach an escape character or a quote indicating the end of the string
		if isStringRegular {
			isStringFinished := false
			for j := i + 1; !isStringFinished; j++ {
				currentCharacter = string(jsonFile[j])
				switch currentCharacter {
				case "\"":
					tokenContent += currentCharacter
					tokenLength++
					isStringFinished = true
				case "\\":
					isStringFinished = true
				default:
					tokenContent += currentCharacter
					tokenLength++
				}
			}
		}

		// Given that this token is a number, we add all digits until we reach
		// an invalid character for a number. We take advantage of the fact that
		// our program only receives valid input, which saves a lot of work.
		if isNumber {
			isNumberFinished := false

			// validNextNumCharacter returns true if the next character is a
			// valid numerical character. Because we only receive valid input,
			// we do not need to worry about tracking whether we have received
			// invalid numbers (counting instances of '+', '-', 'e', 'E', and
			// '.'). If we wished to do so, that functionality would be added
			// in this function.
			validNextNumCharacter := func(s string) bool {
				switch s {
				case "-", "+", "e", "E", ".", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
					return true
				}
				return false
			}

			for j := i + 1; !isNumberFinished; j++ {
				currentCharacter = string(jsonFile[j])

				if validNextNumCharacter(currentCharacter) {
					tokenContent += currentCharacter
					tokenLength++
				} else {
					isNumberFinished = true
				}
			}
		}

		// Only save the token if it is a valid token. Whitespace, invalid
		// characters, and unknown characters will be flagged false.
		if isToken {
			newToken := Token{tokenContent, tokenKind}
			tokenArray = append(tokenArray, newToken)
		}

		i += tokenLength
	}

	return tokenArray
}

// printTokens iterates the array of tokens properly and prints them to standard
// output. It calls extra functions to help with HTML styling, but tracks
// indentation at this level.
func printTokens(tokenArray []Token) {
	indentationLevel := 0 // How many '\t' should be prepended
	isToIndent := false   // Is this token to be indented
	for _, token := range tokenArray {
		fmt.Print(styleHTML(token, &indentationLevel, &isToIndent))
	}
}

// styleHTML calls other functions to help with HTML styling and combines their
// outputs into a single string
func styleHTML(token Token, indentationLevel *int, isToIndent *bool) string {
	colorPre, colorPost := addColor(token)
	whiteSpacePre, whiteSpacePost := addWhiteSpace(token, indentationLevel, isToIndent)
	escapedString := escapeString(token)
	return whiteSpacePre + colorPre + escapedString + colorPost + whiteSpacePost
}

// addColor outputs the <span> tags necessary to color each token. Most of the
// base colors are taken from: https://github.com/reedes/vim-colors-pencil,
// though the use of the colors is different
func addColor(token Token) (string, string) {
	var colorPre, colorPost, color string
	printInColor := true

	switch token.kind {
	case ObjectOpen, ObjectClose:
		color = "#D75F5F"
	case ArrayOpen, ArrayClose:
		color = "#10A778"
	case DelimiterPair:
		color = "#005F87"
	case DelimiterMember:
		color = "#CCCCCC"
	case StringRegular, StringClose:
		color = "#424242"
	case StringEscaped:
		color = "#C30771"
	case Number: // Number
		color = "#6855DE"
	case LiteralBoolTrue, LiteralBoolFalse, LiteralNull: // Literals
		color = "#20A5BA"
	default:
		printInColor = false
	}

	if printInColor {
		colorPre = "<span style=\"color:" + color + "\">"
		colorPost = "</span>"
	}

	return colorPre, colorPost
}

// addWhiteSpace adds white space before and after the token to ensure
// consistent styling. Spacing direction is generally based on the spacing style
// used at https://jsonformatter.curiousconcept.com
func addWhiteSpace(token Token, indentationLevel *int, isToIndent *bool) (string, string) {
	var whiteSpacePre, whiteSpacePost, indentString string

	for i := 1; i < *indentationLevel; i++ {
		indentString += "\t"
	}

	if *isToIndent {
		whiteSpacePre = indentString + "\t"
	}

	*isToIndent = false

	switch token.kind {
	case ObjectOpen, ArrayOpen:
		whiteSpacePost = "\n"
		*indentationLevel++
		*isToIndent = true
	case ObjectClose, ArrayClose:
		whiteSpacePre = "\n" + indentString
		*indentationLevel--
	case DelimiterPair:
		whiteSpacePre = " "
		whiteSpacePost = " "
	case DelimiterMember:
		whiteSpacePost = "\n"
		*isToIndent = true
	}

	return whiteSpacePre, whiteSpacePost
}

// escapeString replaces all characters that cannot be displayed properly in
// HTML with HTML symbols (using their entity number)
func escapeString(token Token) string {
	var escapedString string
	newString := token.content

	for _, character := range newString {
		stringToAdd := string(character)
		switch stringToAdd {
		case "<":
			stringToAdd = "&lt;"
		case ">":
			stringToAdd = "&gt;"
		case "&":
			stringToAdd = "&amp;"
		case "\"":
			stringToAdd = "&quot;"
		case "'":
			stringToAdd = "&apos;"
		}
		escapedString += stringToAdd
	}

	return escapedString
}

// printHeader prints a standard HTML header, sets the background color, and
// sets up the text styling
func printHeader() {
	fmt.Println("<!doctype html>")
	fmt.Println("<html>")
	fmt.Println("\t" + "<head>")
	fmt.Println("\t\t" + "<title>Assignment 2 - Colorized JSON</title>")
	fmt.Println("\t" + "</head>")
	fmt.Println("\t" + "<body style=\"background-color:#F1F1F1\">")
	fmt.Println("\t\t" + "<span style=\"font-family:monospace; tab-size:4; white-space:pre\">")
}

// printFooter prints a standard HTML footer
func printFooter() {
	fmt.Print("\n")
	fmt.Println("\t\t" + "</span>")
	fmt.Println("\t" + "</body>")
	fmt.Println("</html>")
}
