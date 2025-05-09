lexer grammar Lexer;
WS: [ \t\r\n]+ -> skip;
NEWLINE: [\r\n]+;
MULTILINE_COMMENT: '/*' (MULTILINE_COMMENT | .)*? '*/' -> skip;
LINE_COMMENT: '//' .*? NEWLINE -> channel(HIDDEN);

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
ONEOF: 'oneof';
THROUGH: 'through';
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
DIV: '/';
RESTRICT: '\\';

PERCENTAGE_PORTION_LITERAL: [0-9]+ ('.' [0-9]+)? '%';

STRING: '"' ('\\"' | ~[\r\n"])* '"';

IDENTIFIER: [a-z]+ [a-z_]*;
NUMBER: MINUS? [0-9]+ ('_' [0-9]+)*;
ASSET: [A-Z][A-Z0-9]* ('/' [0-9]+)?;

ACCOUNT_START: '@' -> pushMode(ACCOUNT_MODE);
COLON: ':' -> pushMode(ACCOUNT_MODE);
fragment VARIABLE_NAME_FRAGMENT: '$' [a-z_]+ [a-z0-9_]*;

mode ACCOUNT_MODE;
ACCOUNT_TEXT: [a-zA-Z0-9_-]+ -> popMode;
VARIABLE_NAME_ACC: VARIABLE_NAME_FRAGMENT -> popMode;

mode DEFAULT_MODE;
VARIABLE_NAME: VARIABLE_NAME_FRAGMENT;