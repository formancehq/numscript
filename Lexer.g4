lexer grammar Lexer;
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
PLUS: '+';
MINUS: '-';

RATIO_PORTION_LITERAL: [0-9]+ [ ]? '/' [ ]? [0-9]+;
PERCENTAGE_PORTION_LITERAL: [0-9]+ ('.' [0-9]+)? '%';

STRING: '"' ('\\"' | ~[\r\n"])* '"';

IDENTIFIER: [a-z]+ [a-z_]*;
NUMBER: MINUS? [0-9]+ ('_' [0-9]+)*;
// VARIABLE_NAME: '$' [a-z_]+ [a-z0-9_]*;

ASSET: [A-Z/0-9]+;

ACCOUNT_START: '@' -> pushMode(ACCOUNT_MODE);
COLON: ':' -> pushMode(ACCOUNT_MODE);
// fragment ACCOUNT_FRAGMENT_PART: [a-zA-Z0-9_-]+ | VARIABLE_NAME; ACCOUNT: '@' [a-zA-Z0-9_-]+ (':'
// ACCOUNT_FRAGMENT_PART)*;

fragment VARIABLE_NAME_FRAMGMENT: '$' [a-z_]+ [a-z0-9_]*;

mode ACCOUNT_MODE;

ACCOUNT_TEXT: [a-zA-Z0-9_-]+ -> popMode;
VARIABLE_NAME_ACC: VARIABLE_NAME_FRAMGMENT -> popMode;

mode DEFAULT_MODE;
VARIABLE_NAME_DEFAULT: VARIABLE_NAME_FRAMGMENT;