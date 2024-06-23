grammar Numscript;

// Tokens
NEWLINE: [\r\n]+;
MULTILINE_COMMENT: '/*' (MULTILINE_COMMENT | .)*? '*/' -> skip;
LINE_COMMENT: '//' .*? NEWLINE -> skip;
WS: [ \t\r\n]+ -> skip;

SEND: 'send';
LPARENS: '(';
RPARENS: ')';
LBRACKET: '[';
RBRACKET: ']';
EQ: '=';

NUMBER: [0-9]+;
VARIABLE_NAME: '$' [a-z_]+ [a-z0-9_]*;
ACCOUNT: '@' [a-zA-Z0-9_-]+ (':' [a-zA-Z0-9_-]+)*;
ASSET: [A-Z/0-9]+;

program: statement*;

monetaryLit: LBRACKET (asset = ASSET) (amt = NUMBER) RBRACKET;

statement: SEND monetaryLit LPARENS RPARENS;