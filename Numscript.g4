grammar Numscript;

// Tokens
options {
	tokenVocab = 'Lexer';
}

monetaryLit:
	LBRACKET (asset = valueExpr) (amt = valueExpr) RBRACKET;

portion:
	RATIO_PORTION_LITERAL			# ratio
	| PERCENTAGE_PORTION_LITERAL	# percentage;

accountLiteralPart:
	ACCOUNT_TEXT		# accountTextPart
	| VARIABLE_NAME_ACC	# accountVarPart;

valueExpr:
	VARIABLE_NAME_DEFAULT											# variableExpr
	| ASSET															# assetLiteral
	| STRING														# stringLiteral
	| ACCOUNT_START accountLiteralPart (COLON accountLiteralPart)*	# accountLiteral
	| NUMBER														# numberLiteral
	| monetaryLit													# monetaryLiteral
	| portion														# portionLiteral
	| left = valueExpr op = (PLUS | MINUS) right = valueExpr		# infixExpr;

functionCallArgs: valueExpr ( COMMA valueExpr)*;
functionCall:
	fnName = (OVERDRAFT | IDENTIFIER) LPARENS functionCallArgs? RPARENS;

varOrigin: EQ functionCall;
varDeclaration:
	type_ = IDENTIFIER name = VARIABLE_NAME_DEFAULT varOrigin?;
varsDeclaration: VARS LBRACE varDeclaration* RBRACE;

program: varsDeclaration? statement* EOF;

sentAllLit: LBRACKET (asset = valueExpr) STAR RBRACKET;

allotment:
	portion					# portionedAllotment
	| VARIABLE_NAME_DEFAULT	# portionVariable
	| REMAINING				# remainingAllotment;

source:
	address = valueExpr ALLOWING UNBOUNDED OVERDRAFT						# srcAccountUnboundedOverdraft
	| address = valueExpr ALLOWING OVERDRAFT UP TO maxOvedraft = valueExpr	#
		srcAccountBoundedOverdraft
	| valueExpr							# srcAccount
	| LBRACE allotmentClauseSrc+ RBRACE	# srcAllotment
	| LBRACE source* RBRACE				# srcInorder
	| MAX cap = valueExpr FROM source	# srcCapped;
allotmentClauseSrc: allotment FROM source;

keptOrDestination:
	TO destination	# destinationTo
	| KEPT			# destinationKept;
destinationInOrderClause: MAX valueExpr keptOrDestination;

destination:
	valueExpr																# destAccount
	| LBRACE allotmentClauseDest+ RBRACE									# destAllotment
	| LBRACE destinationInOrderClause* REMAINING keptOrDestination RBRACE	# destInorder;
allotmentClauseDest: allotment keptOrDestination;

sentValue: valueExpr # sentLiteral | sentAllLit # sentAll;

statement:
	SEND sentValue LPARENS SOURCE EQ source DESTINATION EQ destination RPARENS	# sendStatement
	| SAVE sentValue FROM valueExpr												# saveStatement
	| functionCall																# fnCallStatement;