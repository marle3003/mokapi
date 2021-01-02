package pipeline

type mokapiFile struct {
	Pipelines []*pipeline `@@*`
}

type pipeline struct {
	Name string `"pipeline" ( "(" @String ")" )? "{"`
	//Parameters []*Parameter `@parameters*`
	//Options    []*Option `@options*`
	Block *block `"steps" "{" @@ "}" "}"`
}

type parameter struct {
}

type option struct {
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
	Equality *equality `@@`
}

type additive struct {
	Left     *primary `@@`
	Operator string   `( @( "+" | "-" )`
	Right    *primary `@@ )?`
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

type primary struct {
	MemberAccess *memberAccess `( @@`
	Step         *step         ` | @@`
	Literal      *literal      ` | @@ )`
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true" || values[0] == ""
	return nil
}

type literal struct {
	Number *float64 `  @Number`
	String *string  ` | @String`
	Bool   *Boolean ` | @("true" | "false" | "")`
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
	Closure    *closure      ` ( @@`
	Member     *memberAccess ` | @@`
	Identifier string        ` | @Ident`
	Literal    *literal      ` | @@ )`
}

type closure struct {
	Args  []string `"{" (@Ident ( "," @Ident )* "=>")?`
	Block *block   `@@ "}"`
}
