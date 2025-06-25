package parser

import (
	"fmt"
	"github.com/alecthomas/participle/v2/lexer"
	"testing"
)

var symbolsByRune = lexer.SymbolsByRune(schemaLexer)

func getStringLexer(s string, t *testing.T) lexer.Lexer {
	lex, err := schemaLexer.LexString("testing", s)
	if err != nil {
		t.Fatalf("Cannot create lexer: %v", err)
	}
	return lex
}

func getAllTokens(lex lexer.Lexer) ([]lexer.Token, error) {
	res := make([]lexer.Token, 0)
	for {
		nextToken, err := lex.Next()
		if err != nil {
			return res, err
		}
		if nextToken.Type == lexer.EOF {
			return res, nil
		}
		res = append(res, nextToken)
	}
}

type tokenTest struct {
	name           string
	tokensSequence []string
	success        []string
	fail           []string
}

func (tokenTestCase *tokenTest) assertTokensMatch(tokens []lexer.Token, t *testing.T) {
	if len(tokens) != len(tokenTestCase.tokensSequence) {
		t.Fatalf("Failed to match token sequence: expected length %d, got %d", len(tokenTestCase.tokensSequence), len(tokens))
	}

	for i := range tokens {
		tokenName := symbolsByRune[tokens[i].Type]
		if tokenName != tokenTestCase.tokensSequence[i] {
			t.Fatalf("Failed to match token sequence at position %d: expected %s, got %s", i, tokenTestCase.tokensSequence[i], tokenName)
		}
	}
}

func (tokenTestCase *tokenTest) assertTokensDontMatch(tokens []lexer.Token, t *testing.T) {
	if len(tokens) != len(tokenTestCase.tokensSequence) {
		return
	}

	for i := range tokens {
		tokenName := symbolsByRune[tokens[i].Type]
		if tokenName != tokenTestCase.tokensSequence[i] {
			return
		}
	}

	t.Fatalf("Token sequences match when they should not.")
}

func (tokenTestCase *tokenTest) test(t *testing.T) {
	t.Run(tokenTestCase.name, func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			for _, str := range tokenTestCase.success {
				t.Run(str, func(t *testing.T) {
					lex := getStringLexer(str, t)
					allTokens, err := getAllTokens(lex)
					if err != nil {
						t.Fatalf("Cannot lex string: %v", err)
					}
					tokenTestCase.assertTokensMatch(allTokens, t)
				})
			}
		})
		t.Run("fail", func(t *testing.T) {
			for _, str := range tokenTestCase.fail {
				t.Run(str, func(t *testing.T) {
					lex := getStringLexer(str, t)
					allTokens, err := getAllTokens(lex)
					if err != nil {
						return
					}
					tokenTestCase.assertTokensDontMatch(allTokens, t)
				})
			}
		})
	})
}

func combinePairDirected(a, b tokenTest) tokenTest {
	name := fmt.Sprintf("%s:%s", a.name, b.name)
	sequence := make([]string, 0, len(a.tokensSequence)+len(b.tokensSequence))
	sequence = append(sequence, a.tokensSequence...)
	sequence = append(sequence, b.tokensSequence...)

	successStrings := make([]string, 0, len(a.success)*len(b.success))
	for _, aStr := range a.success {
		for _, bStr := range b.success {
			successStrings = append(successStrings, aStr+bStr)
		}
	}

	// Each possible combination of fail+fail, fail+success, success+fail
	failStrings := make([]string, 0, len(a.fail)*len(b.fail)+len(a.fail)*len(b.success)+len(a.success)*len(b.fail))
	for _, aStr := range a.fail {
		for _, bStr := range b.fail {
			failStrings = append(failStrings, aStr+bStr)
		}
	}
	for _, aStr := range a.fail {
		for _, bStr := range b.success {
			failStrings = append(failStrings, aStr+bStr)
		}
	}
	for _, aStr := range a.success {
		for _, bStr := range b.fail {
			failStrings = append(failStrings, aStr+bStr)
		}
	}

	return tokenTest{
		name:           name,
		tokensSequence: sequence,
		success:        successStrings,
		fail:           failStrings,
	}
}

type pairExclude struct {
	first  string // empty means any token
	second string // empty means any token
}

func (e *pairExclude) matches(a, b tokenTest) bool {
	return (e.first == "" || e.first == a.name) && (e.second == "" || e.second == b.name)
}

func generatePairs(tests []tokenTest, excludes []pairExclude) []tokenTest {
	res := make([]tokenTest, 0)
	for i := range tests {
		for j := i + 1; j < len(tests); j++ {
			a, b := tests[i], tests[j]
			shouldAppendAb := true
			shouldAppendBa := true
			for _, exclude := range excludes {
				if exclude.matches(a, b) {
					shouldAppendAb = false
				}
				if exclude.matches(b, a) {
					shouldAppendBa = false
				}
			}
			if shouldAppendAb {
				res = append(res, combinePairDirected(a, b))
			}
			if shouldAppendBa {
				res = append(res, combinePairDirected(b, a))
			}
		}
	}
	return res
}

var singleTokenTests = []tokenTest{
	{
		name:           "Ident",
		tokensSequence: []string{"Ident"},
		success: []string{
			"test",
			"t42est",
			"test42",
			"t",
			"_",
			"_underscore",
			"te-st",
			"test-",
			"test42-",
		},
		fail: []string{
			"42test",
			"-test",
		},
	},
	{
		name:           "String",
		tokensSequence: []string{"String"},
		success: []string{
			`""`,
			`"anything"`,
			`"tes42t"`,
			`"\no backslash handling\"`,
		},
		fail: []string{
			`no quotes`,
			`"single quote`,
		},
	},
	{
		name:           "Slash",
		tokensSequence: []string{"Slash"},
		success: []string{
			`/`,
		},
		fail: []string{
			`//`,
			`\`,
			` /`,
		},
	},
	{
		name:           "Dot",
		tokensSequence: []string{"Dot"},
		success: []string{
			`.`,
		},
		fail: []string{
			`..`,
			`\`,
			` .`,
		},
	},
	{
		name:           "Comma",
		tokensSequence: []string{"Comma"},
		success: []string{
			`,`,
		},
		fail: []string{
			`,,`,
			`\`,
			` ,`,
		},
	},
	{
		name:           "Plus",
		tokensSequence: []string{"Plus"},
		success: []string{
			`+`,
		},
		fail: []string{
			`++`,
			`\`,
			` +`,
		},
	},
	{
		name:           "Star",
		tokensSequence: []string{"Star"},
		success: []string{
			`*`,
		},
		fail: []string{
			`**`,
			`\`,
			` *`,
		},
	},
	{
		name:           "Dollar",
		tokensSequence: []string{"Dollar"},
		success: []string{
			`$`,
		},
		fail: []string{
			`$$`,
			`\`,
			` $`,
		},
	},
	{
		name:           "LParen",
		tokensSequence: []string{"LParen"},
		success: []string{
			`(`,
		},
		fail: []string{
			`((`,
			`\`,
			` (`,
			`)`,
			`[`,
			`]`,
			`{`,
			`}`,
		},
	},
	{
		name:           "RParen",
		tokensSequence: []string{"RParen"},
		success: []string{
			`)`,
		},
		fail: []string{
			`))`,
			`\`,
			` )`,
			`(`,
			`[`,
			`]`,
			`{`,
			`}`,
		},
	},
	{
		name:           "LBracket",
		tokensSequence: []string{"LBracket"},
		success: []string{
			`[`,
		},
		fail: []string{
			`[[`,
			`\`,
			` [`,
			`)`,
			`(`,
			`{`,
			`}`,
		},
	},
	{
		name:           "RBracket",
		tokensSequence: []string{"RBracket"},
		success: []string{
			`]`,
		},
		fail: []string{
			`]]`,
			`\`,
			` ]`,
			`)`,
			`(`,
			`{`,
			`}`,
		},
	},
	{
		name:           "LCBracket",
		tokensSequence: []string{"LCBracket"},
		success: []string{
			`{`,
		},
		fail: []string{
			`{{`,
			`\`,
			` {`,
			`[`,
			`]`,
			`)`,
			`(`,
		},
	},
	{
		name:           "RCBracket",
		tokensSequence: []string{"RCBracket"},
		success: []string{
			`}`,
		},
		fail: []string{
			`}}`,
			`\`,
			` }`,
			`[`,
			`]`,
			`)`,
			`(`,
		},
	},
	{
		name:           "Int",
		tokensSequence: []string{"Int"},
		success: []string{
			`0`,
			`1`,
			`2`,
			`9`,
			`1234567890`,
			`0000123`,
		},
		fail: []string{
			` 0`,
			`\42`,
			`a12`,
			`1a`,
			`-42`,
			`41.`,
			`41.11`,
		},
	},

	{
		name:           "Whitespace",
		tokensSequence: []string{"Whitespace"},
		success: []string{
			" ",
			"\t",
			"\r",
			"\n",
			"  ",
			"\r\n\t    ",
			" \t \r",
		},
		fail: []string{
			" x",
			"\t x",
			" x",
			"              x",
			"\n\n\n\n\nx\n\n\n",
			"   x   ",
		},
	},
}

var pairExcludes = []pairExclude{
	// An ident followed by an int is a single valid ident
	{
		first:  "Ident",
		second: "Int",
	},
	// Fails for edge cases when an invalid Int ends with a valid Ident, e.g. 1a + test == 1atest, which is a valid [Int, Ident] sequence
	{
		first:  "Int",
		second: "Ident",
	},
	// We use a lot of whitespace-prefixed valid tokens to form an invalid token, which leads to correctly parsing the resulting
	// pair as a [Whitespace, Token].
	{
		first:  "Whitespace",
		second: "",
	},
}

func TestLexer(t *testing.T) {
	t.Run("SingleToken", func(t *testing.T) {
		for _, testCase := range singleTokenTests {
			t.Run(testCase.name, testCase.test)
		}
	})
	t.Run("TokenPairs", func(t *testing.T) {
		for _, testCase := range generatePairs(singleTokenTests, pairExcludes) {
			t.Run(testCase.name, testCase.test)
		}
	})
}
