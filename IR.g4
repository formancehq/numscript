grammar IR;

// --- Parser rules ---

program: line* EOF;

line: labelMarker | instruction;

labelMarker: LABEL;

instruction
    : dest '=' instrCall                  # instrWithDest
    | instrCall                           # instrNoDest
    | dest '=' const_                     # constAssign
    | dest '=' left=reg op=(PLUS | MINUS) right=reg       # infixInstr
    | left=reg op=(PLUS_EQ | MINUS_EQ) right=reg # compoundAssignInstr
    ;

dest
    : reg                                 # destReg
    | '_'                                 # destDiscard
    | '[' regList ']'                     # destList
    ;

regList: reg (',' reg)*;

instrCall: instrName '(' args ')';

instrName: IDENTIFIER ('<' typeName '>')?;

typeName: TYPE_KEYWORD;

args: (arg (',' arg)*)?;

arg
    : value                               # positionalArg
    | IDENTIFIER ':' value                # labeledArg
    ;

value
    : reg                                 # valReg
    | LABEL                               # valLabel
    | INT                                 # valInt
    | '[' regList ']'                     # valRegList
    ;

const_
    : STRING                              # constString
    | INT                                 # constInt
    ;

reg: REG;

// --- Lexer rules ---

WS: [ \t]+ -> skip;
NEWLINE: [\r\n]+ -> skip;

// Must come before IDENTIFIER so keywords are not swallowed
TYPE_KEYWORD: 'int' | 'str' | 'portion' | 'monetary';

REG: '$' [a-zA-Z_] [a-zA-Z0-9_]*;
LABEL: '#' [a-zA-Z_] [a-zA-Z0-9_]*;
INT: [0-9]+;
STRING: '"' ('\\"' | ~[\r\n"])* '"';
IDENTIFIER: [a-z] [a-z0-9_]*;

LPAREN: '(';
RPAREN: ')';
LBRACKET: '[';
RBRACKET: ']';
COMMA: ',';
EQ: '=';
PLUS: '+';
MINUS: '-';
PLUS_EQ: '+=';
MINUS_EQ: '-=';
LT: '<';
GT: '>';
UNDERSCORE: '_';
