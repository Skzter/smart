import moo from "moo";

// Moo lexer configuration
const lexer = moo.compile({
    comment: /\/\/.*?$/,
    string: /"(?:\\["\\]|[^\n"\\])*"|'(?:\\['\\]|[^\n'\\])*'|`(?:\\[`\\]|[^`\\])*`/,
    keyword: [
        "import",
        "from",
        "const",
        "let",
        "var",
        "function",
        "async",
        "await",
        "if",
        "else",
        "for",
        "while",
        "return",
        "export",
        "default",
        "class",
        "extends",
        "new",
        "this",
        "super",
        "try",
        "catch",
        "throw",
        "finally",
        "break",
        "continue",
        "switch",
        "case",
        "typeof",
        "instanceof",
        "void",
        "delete",
        "in",
        "of",
        "interface",
        "type",
        "enum",
        "namespace",
        "module",
        "declare",
        "public",
        "private",
        "protected",
        "static",
        "readonly",
        "abstract",
    ],
    functionCall: /[a-zA-Z_$][a-zA-Z0-9_$]*(?=\s*\()/,
    identifier: /[a-zA-Z_$][a-zA-Z0-9_$]*/,
    number: /0x[a-fA-F0-9]+|[0-9]+\.?[0-9]*/,
    operator: /=>|[+\-*/%=<>!&|^~?:]+/,
    punctuation: /[{}[\]();,.]/,
    whitespace: { match: /\s+/, lineBreaks: true },
});

export function tokenize(text: string) {
    lexer.reset(text);
    const tokens = [];
    let token;

    while ((token = lexer.next())) {
        tokens.push(token);
    }

    return tokens;
}
