grammar Numscript;

options {
	tokenVocab = 'Lexer';
}

monetaryLit:
	LBRACKET (asset = valueExpr) (amt = valueExpr) RBRACKET;

accountLiteralPart:
	ACCOUNT_TEXT		# accountTextPart
	| VARIABLE_NAME_ACC	# accountVarPart;

valueExpr:
	VARIABLE_NAME													# variableExpr
	| ASSET															# assetLiteral
	| STRING														# stringLiteral
	| ACCOUNT_START accountLiteralPart (COLON accountLiteralPart)*	# accountLiteral
	| NUMBER														# numberLiteral
	| PERCENTAGE_PORTION_LITERAL									# percentagePortionLiteral
	| monetaryLit													# monetaryLiteral
	| left = valueExpr op = DIV right = valueExpr					# infixExpr
	| left = valueExpr op = (PLUS | MINUS) right = valueExpr		# infixExpr
	| LPARENS valueExpr RPARENS										# parenthesizedExpr
	| functionCall													# application;

functionCallArgs: valueExpr ( COMMA valueExpr)*;
functionCall:
	fnName = (OVERDRAFT | IDENTIFIER) LPARENS functionCallArgs? RPARENS;

varOrigin: EQ valueExpr;
varDeclaration:
	type_ = IDENTIFIER name = VARIABLE_NAME varOrigin?;
varsDeclaration: VARS LBRACE varDeclaration* RBRACE;

program: varsDeclaration? statement* EOF;

sentAllLit: LBRACKET (asset = valueExpr) STAR RBRACKET;

allotment:
	valueExpr	# portionedAllotment
	| REMAINING	# remainingAllotment;

source:
	address = valueExpr ALLOWING UNBOUNDED OVERDRAFT						# srcAccountUnboundedOverdraft
	| address = valueExpr ALLOWING OVERDRAFT UP TO maxOvedraft = valueExpr	#
		srcAccountBoundedOverdraft
	| valueExpr							# srcAccount
	| LBRACE allotmentClauseSrc+ RBRACE	# srcAllotment
	| LBRACE source* RBRACE				# srcInorder
	| ONEOF LBRACE source+ RBRACE		# srcOneof
	| MAX cap = valueExpr FROM source	# srcCapped;
allotmentClauseSrc: allotment FROM source;

keptOrDestination:
	TO destination	# destinationTo
	| KEPT			# destinationKept;
destinationInOrderClause: MAX valueExpr keptOrDestination;

destination:
	valueExpr																	# destAccount
	| LBRACE allotmentClauseDest+ RBRACE										# destAllotment
	| LBRACE destinationInOrderClause* REMAINING keptOrDestination RBRACE		# destInorder
	| ONEOF LBRACE destinationInOrderClause* REMAINING keptOrDestination RBRACE	# destOneof;
allotmentClauseDest: allotment keptOrDestination;

sentValue: valueExpr # sentLiteral | sentAllLit # sentAll;

statement:
	SEND sentValue LPARENS SOURCE EQ source DESTINATION EQ destination RPARENS	# sendStatement
	| SAVE sentValue FROM valueExpr												# saveStatement
	| functionCall																# fnCallStatement;