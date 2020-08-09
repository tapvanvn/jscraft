package js

import (
	"newcontinent-team.com/jscraft/tokenize"
)

const (
	TokenJSUnknown           = 0
	TokenJSWord              = 1
	TokenJSOperator          = 2
	TokenJSPhraseBreak       = 3
	TokenJSPhraseStrongBreak = 4 //need ; after
	TokenJSScopeBegin        = 5
	TokenJSScopeEnd          = 6
	TokenJSWordBreak         = 7
	TokenJSGlueBegin         = 8
	TokenJSGlueEnd           = 9

	TokenJSBracket        = 100
	TokenJSBlock          = 101
	TokenJSBracketSquare  = 102
	TokenJSUnaryOperator  = 103 // !, ~, ++, --
	TokenJSBinaryOperator = 104 // <>+-*/%, <=, >=, <<, >>, >>>, ||, |, &&, &, ^, **, ==, ===, !=, !==
	TokenJSAssign         = 105 // =
	TokenJSRightArrow     = 106 // =>
	TokenJSLineComment    = 107
	TokenJSBlockComment   = 108
	TokenJSPhrase         = 109
	TokenJSRegex          = 110

	TokenJSFunction       = 200
	TokenJSFunctionLambda = 201
	TokenJSVariable       = 202
	TokenJSString         = 203
	TokenJSFor            = 204
	TokenJSIf             = 205
	TokenJSElseIf         = 206
	TokenJSElse           = 207
	TokenJSSwitch         = 208
	TokenJSWhile          = 209
	TokenJSDo             = 210

	TokenJSCraft      = 300
	TokenJSCraftDebug = 301

	TokenJSPatchStream = 400
)

//TokenName return name from type of token
func TokenName(Type int) string {

	switch Type {

	case TokenJSUnknown:
		return "unknown"

	case TokenJSWord:
		return "word"

	case TokenJSOperator:
		return "operator"

	case TokenJSPhraseBreak:
		return "phrase break"

	case TokenJSBracket:
		return "bracket"

	case TokenJSBlock:
		return "block"

	case TokenJSBracketSquare:
		return "bracket square"

	case TokenJSUnaryOperator:
		return "unary operator"

	case TokenJSBinaryOperator:
		return "binary operator"

	case TokenJSAssign:
		return "assign"

	case TokenJSRightArrow:
		return "right arrow"

	case TokenJSLineComment:
		return "line comment"

	case TokenJSBlockComment:
		return "block comment"

	case TokenJSPhrase:
		return "phrase"

	case TokenJSFunction:
		return "function"

	case TokenJSFunctionLambda:
		return "lambda"

	case TokenJSVariable:
		return "variable"

	case TokenJSString:
		return "string"

	case TokenJSFor:
		return "for"

	case TokenJSIf:
		return "if"

	case TokenJSElseIf:
		return "else if"

	case TokenJSElse:
		return "else"

	case TokenJSSwitch:
		return "switch"

	case TokenJSWhile:
		return "while"

	case TokenJSDo:
		return "do"

	case TokenJSCraft:
		return "craft"

	case TokenJSCraftDebug:
		return "craft debug"

	default:
		return "unknown"
	}
}

//KeyWords keywords of javascript
var KeyWords string = `
,abstract,arguments,await,boolean,
,break,byte,case,catch,
,char,class,const,continue,
,debugger,default,delete,do,
,double,else,enum,eval,
,export,extends,false,final,finally,float,for,function,
,goto,if,implements,import,
,in,instanceof,int,interface,
,let,long,native,new,
,null,package,private,protected,
,public,return,short,static,
,super,switch,synchronized,this,
,throw,throws,transient,true,
,try,typeof,var,void,
,volatile,while,with,yield,`

//Ignores tokens that will be ignore
var Ignores = []int{

	TokenJSLineComment,

	TokenJSBlockComment,
}

//Patterns Patterns
var Patterns = []tokenize.Pattern{

	//pattern if block
	tokenize.Pattern{
		Type:                 TokenJSIf,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "if", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
		},
	},

	//pattern if phrase
	tokenize.Pattern{
		Type:                 TokenJSIf,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "if", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{ExportType: TokenJSPhrase, IsPhraseUntil: true},
		},
	},

	//phattern else if block
	tokenize.Pattern{
		Type:                 TokenJSElseIf,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "else", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: "if", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
		},
	},

	//pattern else if pharse
	tokenize.Pattern{
		Type:                 TokenJSElseIf,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "else", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: "if", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{ExportType: TokenJSPhrase, IsPhraseUntil: true},
		},
	},
	//pattern else block
	tokenize.Pattern{
		Type:                 TokenJSElse,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "else", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
		},
	},

	//pattern else phrase
	tokenize.Pattern{
		Type:                 TokenJSElse,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "else", IsIgnoreInResult: true},
			tokenize.PatternToken{ExportType: TokenJSPhrase, IsPhraseUntil: true},
		},
	},

	//pattern for
	tokenize.Pattern{
		Type:                 TokenJSFor,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "for", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
		},
	},

	//pattern function with keyword
	tokenize.Pattern{
		Type:                 TokenJSFunction,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "function", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSWord},
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
		},
	},

	//pattern lambda
	tokenize.Pattern{
		Type:                 TokenJSFunctionLambda,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{Type: TokenJSRightArrow, IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
		},
	},

	//pattern switch
	tokenize.Pattern{
		Type:                 TokenJSSwitch,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "switch", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
		},
	},

	//pattern while block
	tokenize.Pattern{
		Type:                 TokenJSWhile,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "while", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
		},
	},
	//pattern while phrase
	tokenize.Pattern{
		Type:                 TokenJSWhile,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "while", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
			tokenize.PatternToken{ExportType: TokenJSPhrase, IsPhraseUntil: true},
		},
	},

	//pattern do block
	tokenize.Pattern{
		Type:                 TokenJSDo,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "do", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
			tokenize.PatternToken{Content: "while", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
		},
	},

	//pattern do phrase
	tokenize.Pattern{
		Type:                 TokenJSDo,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "do", IsIgnoreInResult: true},
			tokenize.PatternToken{ExportType: TokenJSPhrase, IsPhraseUntil: true},
			tokenize.PatternToken{Content: "while", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBracket},
		},
	},

	//pattern jscraft.require
	tokenize.Pattern{
		Type:                 TokenJSCraft,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "jscraft", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: ".", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: "require"},
			tokenize.PatternToken{Type: TokenJSBracket},
		},
	},

	//pattern jscraft.template
	tokenize.Pattern{
		Type:                 TokenJSCraft,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "jscraft", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: ".", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: "template"},
			tokenize.PatternToken{Type: TokenJSBracket, CanNested: true},
		},
	},

	//pattern jscraft.build
	tokenize.Pattern{
		Type:                 TokenJSCraft,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "jscraft", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: ".", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: "build"},
			tokenize.PatternToken{Type: TokenJSBracket, CanNested: true},
		},
	},

	//pattern jscraft.conflict
	tokenize.Pattern{
		Type:                 TokenJSCraft,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "jscraft", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: ".", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: "conflict"},
			tokenize.PatternToken{Type: TokenJSBracket},
		},
	},

	//pattern jscraft.conflict
	tokenize.Pattern{
		Type:                 TokenJSCraft,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "jscraft", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: ".", IsIgnoreInResult: true},
			tokenize.PatternToken{Content: "fetch"},
			tokenize.PatternToken{Type: TokenJSBracket},
		},
	},

	//pattern jscraft.conflict
	tokenize.Pattern{
		Type:                 TokenJSCraftDebug,
		IsRemoveGlobalIgnore: true,
		Struct: []tokenize.PatternToken{
			tokenize.PatternToken{Content: "jscraft_debug", IsIgnoreInResult: true},
			tokenize.PatternToken{Type: TokenJSBlock, CanNested: true},
		},
	},
}
