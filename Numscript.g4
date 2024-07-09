grammar Numscript;

// Tokens
WS: [ \t\r\n]+ -> skip;
NEWLINE: [\r\n]+;
MULTILINE_COMMENT: '/*' (MULTILINE_COMMENT | .)*? '*/' -> skip;
LINE_COMMENT: '//' .*? NEWLINE -> skip;

SOURCE: 'source';
DESTINATION: 'destination';
SEND: 'send';
LPARENS: '(';
RPARENS: ')';
LBRACKET: '[';
RBRACKET: ']';
LBRACE: '{';
RBRACE: '}';
EQ: '=';

NUMBER: [0-9]+;
VARIABLE_NAME: '$' [a-z_]+ [a-z0-9_]*;
ACCOUNT: '@' [a-zA-Z0-9_-]+ (':' [a-zA-Z0-9_-]+)*;
ASSET: [A-Z/0-9]+;

program: statement*;

monetaryLit: LBRACKET (asset = ASSET) (amt = NUMBER) RBRACKET;

source:
	ACCOUNT					# srcAccount
	| VARIABLE_NAME			# srcVariable
	| LBRACE source* RBRACE	# srcSeq;

destination:
	ACCOUNT							# destAccount
	| VARIABLE_NAME					# destVariable
	| LBRACE destination* RBRACE	# destSeq;

statement:
	SEND monetaryLit LPARENS SOURCE EQ source DESTINATION EQ destination RPARENS;