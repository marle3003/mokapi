package pipeline

type mokapiFile struct {
	Pipelines []*pipeline `@@*`
}

type pipeline struct {
	Name string `"pipeline" ( "(" @String? ")" )? "{"`
	//Parameters []*Parameter `@parameters*`
	//Options    []*Option `@options*`
	Stages []*stage `"stages" "{" @@* "}" "}"`
}

type stage struct {
	Name  string      `"stage" "(" @String ")" "{"`
	When  *expression `("when" "{" @@ "}")?`
	Steps *block      `"steps" "{" @@ "}" "}"`
}

func (s *stage) DisplayName() string {
	return s.Name[1 : len(s.Name)-1]
}

type parameter struct {
}

type option struct {
}

type when struct {
	Expression *expression `@@?`
}

type block struct {
	Statements []*statement `( @@ EOL* )*`
}

type statement struct {
	Assignment *assignment `( @@`
	Expression *expression ` | @@ )`
}

type assignment struct {
	Variable   *variable   `@@`
	Operator   string      `@( ":=" | "=" )`
	Expression *expression `@@`
}

type variable struct {
	Member     string `( @Member`
	Identifier string ` | @Ident )`
}

type expression struct {
	OrCondition *orCondition `( @@ | "(" @@ ")" )`
}

type orCondition struct {
	Left  *andCondition `@@`
	Right *orCondition  `( "||" @@)?`
}

type andCondition struct {
	Left  *equality     `@@`
	Right *andCondition `( "&&" @@)?`
}

type equality struct {
	Left     *relational `@@`
	Operator string      `( @( "==" | "!=" )`
	Right    *relational `@@ )?`
}

type relational struct {
	Left     *additive `@@`
	Operator string    `( @( "<" | ">" | "<=" | ">=" )`
	Right    *additive `@@ )?`
}

type additive struct {
	Left     *unary    `@@`
	Operator string    `( @( "+" | "-" )`
	Right    *additive `@@ )?`
}

type unary struct {
	Operator string   `@('!')?`
	Primary  *primary `@@`
}

type primary struct {
	Closure      *closure      `( @@`
	MemberAccess *memberAccess ` | @@`
	Step         *step         ` | @@`
	Literal      *literal      ` | @@ `
	Expression   *expression   ` | "(" @@ ")" )`
}

type Boolean bool

func NewBoolean(b bool) *Boolean {
	boolean := Boolean(b)
	return &boolean
}

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true" || values[0] == ""
	return nil
}

type literal struct {
	Number *float64 `  @Number`
	String *string  ` | @String`
	Bool   *Boolean ` | @("true" | "false")`
}

type step struct {
	Name string      `@Ident`
	Args []*argument `( @@ ( "," @@ )* )?`
}

type memberAccess struct {
	Name string      `@Member`
	Args []*argument `( @@ ( "," @@ )* )?`
}

type argument struct {
	Name  string         `( @Ident ":" )?`
	Value *argumentValue `@@`
}

type argumentValue struct {
	Expression *expression `@@`
}

type closure struct {
	Args  []string `"{" (@Ident ( "," @Ident )* "=>")?`
	Block *block   `@@ "}"`
}
