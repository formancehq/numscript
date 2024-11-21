grammar Numscript;

// Tokens
WS: [ \t\r\n]+ -> skip;
NEWLINE: [\r\n]+;
MULTILINE_COMMENT: '/*' (MULTILINE_COMMENT | .)*? '*/' -> skip;
LINE_COMMENT: '//' .*? NEWLINE -> skip;

VARS: 'vars';
MAX: 'max';
SOURCE: 'source';
DESTINATION: 'destination';
SEND: 'send';
FROM: 'from';
UP: 'up';
TO: 'to';
REMAINING: 'remaining';
ALLOWING: 'allowing';
UNBOUNDED: 'unbounded';
OVERDRAFT: 'overdraft';
IF: 'if';
ELSE: 'else';
KEPT: 'kept';
SAVE: 'save';
LPARENS: '(';
RPARENS: ')';
LBRACKET: '[';
RBRACKET: ']';
LBRACE: '{';
RBRACE: '}';
COMMA: ',';
EQ: '=';
STAR: '*';
MINUS: '-';

RATIO_PORTION_LITERAL: [0-9]+ [ ]? '/' [ ]? [0-9]+;
PERCENTAGE_PORTION_LITERAL: [0-9]+ ('.' [0-9]+)? '%';

STRING: '"' ('\\"' | ~[\r\n"])* '"';

IDENTIFIER: [a-z]+ [a-z_]*;
NUMBER: MINUS? [0-9]+;
VARIABLE_NAME: '$' [a-z_]+ [a-z0-9_]*;
ACCOUNT: '@' [a-zA-Z0-9_-]+ (':' [a-zA-Z0-9_-]+)*;
ASSET: [A-Z/0-9]+;

monetaryLit:
	LBRACKET (asset = valueExpr) (amt = valueExpr) RBRACKET;

portion:
	RATIO_PORTION_LITERAL			# ratio
	| PERCENTAGE_PORTION_LITERAL	# percentage;

valueExpr:
	VARIABLE_NAME											# variableExpr
	| ASSET													# assetLiteral
	| STRING												# stringLiteral
	| ACCOUNT												# accountLiteral
	| NUMBER												# numberLiteral
	| monetaryLit											# monetaryLiteral
	| portion												# portionLiteral
	| left = valueExpr op = ('+' | '-') right = valueExpr	# infixExpr;

functionCallArgs: valueExpr ( COMMA valueExpr)*;
functionCall:
	fnName = (OVERDRAFT | IDENTIFIER) LPARENS functionCallArgs? RPARENS;

varOrigin: EQ functionCall;
varDeclaration:
	type_ = IDENTIFIER name = VARIABLE_NAME varOrigin?;
varsDeclaration: VARS LBRACE varDeclaration* RBRACE;

program: varsDeclaration? statement* EOF;

sentAllLit: LBRACKET (asset = valueExpr) STAR RBRACKET;

cap: monetaryLit # litCap | VARIABLE_NAME # varCap;

allotment:
	portion			# portionedAllotment
	| VARIABLE_NAME	# portionVariable
	| REMAINING		# remainingAllotment;

source:
	address = valueExpr ALLOWING UNBOUNDED OVERDRAFT						# srcAccountUnboundedOverdraft
	| address = valueExpr ALLOWING OVERDRAFT UP TO maxOvedraft = valueExpr	#
		srcAccountBoundedOverdraft
	| valueExpr							# srcAccount
	| LBRACE allotmentClauseSrc+ RBRACE	# srcAllotment
	| LBRACE source* RBRACE				# srcInorder
	| MAX cap FROM source				# srcCapped;
allotmentClauseSrc: allotment FROM source;

keptOrDestination:
	TO destination	# destinationTo
	| KEPT			# destinationKept;
destinationInOrderClause: MAX valueExpr keptOrDestination;

destination:
	valueExpr																# destAccount
	| ifBranch = destination IF valueExpr ELSE elseBranch = destination		# destIf
	| LBRACE allotmentClauseDest+ RBRACE									# destAllotment
	| LBRACE destinationInOrderClause* REMAINING keptOrDestination RBRACE	# destInorder;
allotmentClauseDest: allotment keptOrDestination;

sentValue: valueExpr # sentLiteral | sentAllLit # sentAll;

statement:
	SEND sentValue LPARENS SOURCE EQ source DESTINATION EQ destination RPARENS	# sendStatement
	| SAVE sentValue FROM valueExpr												# saveStatement
	| functionCall																# fnCallStatement;