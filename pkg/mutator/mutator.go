package mutator

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strconv"
	"strings"
)

const (
	ProfileGremlinsCompatible = "gremlins-compatible"
	ProfileConservativeFast   = "conservative-fast"
	ProfileConservative       = "conservative"
	ProfileDefault            = "default"
	ProfileAggressive         = "aggressive"
)

type Definition struct {
	Name                 string   `json:"name"`
	Profile              string   `json:"profile"`
	Risk                 string   `json:"risk"`
	EquivalentMutantRisk string   `json:"equivalent_mutant_risk"`
	CompileErrorRisk     string   `json:"compile_error_risk"`
	ASTNodes             []string `json:"ast_nodes"`
	Example              string   `json:"example"`
	Reason               string   `json:"reason"`
}

type Mutant struct {
	ID               string `json:"id"`
	Module           string `json:"module"`
	Package          string `json:"package"`
	File             string `json:"file"`
	Line             int    `json:"line"`
	Function         string `json:"function"`
	Operator         string `json:"operator"`
	Original         string `json:"original"`
	Mutated          string `json:"mutated"`
	StartOffset      int    `json:"start_offset"`
	EndOffset        int    `json:"end_offset"`
	Diff             string `json:"unified_diff"`
	Fingerprint      string `json:"fingerprint"`
	Hint             string `json:"hint"`
	Description      string `json:"description"`
	EquivalentRisk   string `json:"equivalent_risk"`
	Recommendation   string `json:"recommendation"`
	CompileErrorRisk string `json:"compile_error_risk"`
}

func Definitions() []Definition {
	return []Definition{
		{Name: "conditionals-negation", Profile: ProfileGremlinsCompatible, Risk: "low", EquivalentMutantRisk: "medium", CompileErrorRisk: "low", ASTNodes: []string{"ast.BinaryExpr"}, Example: "a == b -> a != b", Reason: "Fast branch behavior signal."},
		{Name: "conditionals-boundary", Profile: ProfileGremlinsCompatible, Risk: "low", EquivalentMutantRisk: "medium", CompileErrorRisk: "low", ASTNodes: []string{"ast.BinaryExpr"}, Example: "a < b -> a <= b", Reason: "Fast boundary-condition signal."},
		{Name: "arithmetic-basic", Profile: ProfileGremlinsCompatible, Risk: "medium", EquivalentMutantRisk: "low", CompileErrorRisk: "medium", ASTNodes: []string{"ast.BinaryExpr"}, Example: "a + b -> a - b", Reason: "Numeric result signal for fast CI."},
		{Name: "conditionals-negation", Profile: ProfileConservativeFast, Risk: "low", EquivalentMutantRisk: "medium", CompileErrorRisk: "low", ASTNodes: []string{"ast.BinaryExpr"}, Example: "a == b -> a != b", Reason: "Fast branch behavior signal."},
		{Name: "conditionals-boundary", Profile: ProfileConservativeFast, Risk: "low", EquivalentMutantRisk: "medium", CompileErrorRisk: "low", ASTNodes: []string{"ast.BinaryExpr"}, Example: "a < b -> a <= b", Reason: "Fast boundary-condition signal."},
		{Name: "arithmetic-basic", Profile: ProfileConservativeFast, Risk: "medium", EquivalentMutantRisk: "low", CompileErrorRisk: "medium", ASTNodes: []string{"ast.BinaryExpr"}, Example: "a + b -> a - b", Reason: "Numeric result signal for fast CI."},
		{Name: "logical", Profile: ProfileConservative, Risk: "low", EquivalentMutantRisk: "medium", CompileErrorRisk: "low", ASTNodes: []string{"ast.BinaryExpr"}, Example: "a && b -> a || b", Reason: "Captures missing boolean combination assertions."},
		{Name: "boolean-literals", Profile: ProfileConservative, Risk: "low", EquivalentMutantRisk: "low", CompileErrorRisk: "low", ASTNodes: []string{"ast.Ident"}, Example: "true -> false", Reason: "Simple branch outcome signal."},
		{Name: "string-empty-literals", Profile: ProfileConservative, Risk: "medium", EquivalentMutantRisk: "medium", CompileErrorRisk: "low", ASTNodes: []string{"ast.BasicLit"}, Example: `"error" -> ""`, Reason: "Controlled string behavior signal with low compile risk."},
		{Name: "nil-checks", Profile: ProfileDefault, Risk: "medium", EquivalentMutantRisk: "high", CompileErrorRisk: "low", ASTNodes: []string{"ast.BinaryExpr"}, Example: "err == nil -> err != nil", Reason: "Important Go error-path signal but high equivalence risk."},
		{Name: "error-returns", Profile: ProfileDefault, Risk: "medium", EquivalentMutantRisk: "high", CompileErrorRisk: "medium", ASTNodes: []string{"ast.IfStmt"}, Example: "err == nil -> err != nil", Reason: "Controlled error-path mutation for nightly/default runs."},
		{Name: "numeric-literals", Profile: ProfileDefault, Risk: "medium", EquivalentMutantRisk: "medium", CompileErrorRisk: "low", ASTNodes: []string{"ast.BasicLit"}, Example: "2 -> 1", Reason: "Controlled numeric literal signal for default runs."},
		{Name: "return-bool-literals", Profile: ProfileDefault, Risk: "medium", EquivalentMutantRisk: "medium", CompileErrorRisk: "low", ASTNodes: []string{"ast.ReturnStmt"}, Example: "return true -> return false", Reason: "Return behavior signal without broad return rewrites."},
		{Name: "literals", Profile: ProfileAggressive, Risk: "high", EquivalentMutantRisk: "high", CompileErrorRisk: "medium", ASTNodes: []string{"ast.BasicLit"}, Example: "1 -> 0", Reason: "Broad campaign-only literal exploration."},
		{Name: "returns", Profile: ProfileAggressive, Risk: "high", EquivalentMutantRisk: "high", CompileErrorRisk: "medium", ASTNodes: []string{"ast.ReturnStmt"}, Example: "return true -> return false", Reason: "Campaign-only return behavior exploration."},
		{Name: "loop-control", Profile: ProfileAggressive, Risk: "high", EquivalentMutantRisk: "high", CompileErrorRisk: "medium", ASTNodes: []string{"ast.ForStmt"}, Example: "i < n -> i <= n", Reason: "Campaign-only loop boundary exploration."},
		{Name: "slice-map-len-boundary", Profile: ProfileAggressive, Risk: "high", EquivalentMutantRisk: "medium", CompileErrorRisk: "low", ASTNodes: []string{"ast.BinaryExpr"}, Example: "len(xs) > 0 -> len(xs) >= 0", Reason: "Targets collection boundary assumptions."},
	}
}

type inlineIgnore struct {
	line     int
	operator string
	reason   string
}

func ValidateInlineIgnores(filename string, src []byte, requireReason bool) ([]inlineIgnore, error) {
	var ignores []inlineIgnore
	lines := strings.Split(string(src), "\n")
	for i, line := range lines {
		idx := strings.Index(line, "cervomut:ignore")
		if idx < 0 {
			continue
		}
		if commentIdx := strings.Index(line[:idx], "//"); commentIdx < 0 {
			continue
		}
		rest := strings.TrimSpace(line[idx+len("cervomut:ignore"):])
		fields := strings.Fields(rest)
		operator := "*"
		if len(fields) > 0 && !strings.HasPrefix(fields[0], "reason=") {
			operator = fields[0]
		}
		reason := ""
		if reasonIdx := strings.Index(rest, "reason="); reasonIdx >= 0 {
			raw := strings.TrimSpace(rest[reasonIdx+len("reason="):])
			if parsed, err := strconv.Unquote(raw); err == nil {
				reason = parsed
			} else {
				reason = strings.Trim(raw, `"`)
			}
		}
		if requireReason && reason == "" {
			return nil, fmt.Errorf("%s:%d inline ignore requires reason", filename, i+1)
		}
		ignores = append(ignores, inlineIgnore{line: i + 2, operator: operator, reason: reason})
	}
	return ignores, nil
}

func Generate(pkg, filename string, src []byte, profile string) ([]Mutant, error) {
	if profile == "" {
		profile = ProfileConservativeFast
	}
	ignores, err := ValidateInlineIgnores(filename, src, true)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var mutants []Mutant
	var fn string
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			prev := fn
			fn = n.Name.Name
			ast.Inspect(n.Body, func(child ast.Node) bool {
				collectNode(&mutants, fset, pkg, filename, src, fn, child, profile, ignores)
				return true
			})
			fn = prev
			return false
		default:
			collectNode(&mutants, fset, pkg, filename, src, fn, node, profile, ignores)
		}
		return true
	})
	return mutants, nil
}

func collectNode(mutants *[]Mutant, fset *token.FileSet, pkg, filename string, src []byte, fn string, node ast.Node, profile string, ignores []inlineIgnore) {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		addBinaryMutants(mutants, fset, pkg, filename, src, fn, n, profile, ignores)
	case *ast.Ident:
		if n.Name == "true" {
			addMutation(mutants, fset, pkg, filename, src, fn, n, "boolean-literals", "true", "false", profile, ignores)
		}
		if n.Name == "false" {
			addMutation(mutants, fset, pkg, filename, src, fn, n, "boolean-literals", "false", "true", profile, ignores)
		}
	case *ast.BasicLit:
		if n.Kind == token.INT && n.Value != "0" {
			addMutation(mutants, fset, pkg, filename, src, fn, n, "numeric-literals", n.Value, "0", profile, ignores)
			addMutation(mutants, fset, pkg, filename, src, fn, n, "literals", n.Value, "0", profile, ignores)
		}
		if n.Kind == token.INT && n.Value == "0" {
			addMutation(mutants, fset, pkg, filename, src, fn, n, "numeric-literals", "0", "1", profile, ignores)
			addMutation(mutants, fset, pkg, filename, src, fn, n, "literals", "0", "1", profile, ignores)
		}
		if n.Kind == token.STRING && n.Value != `""` {
			addMutation(mutants, fset, pkg, filename, src, fn, n, "string-empty-literals", n.Value, `""`, profile, ignores)
			addMutation(mutants, fset, pkg, filename, src, fn, n, "literals", n.Value, `""`, profile, ignores)
		}
	case *ast.ReturnStmt:
		for _, result := range n.Results {
			ident, ok := result.(*ast.Ident)
			if !ok {
				continue
			}
			if ident.Name == "true" {
				addMutation(mutants, fset, pkg, filename, src, fn, ident, "return-bool-literals", "true", "false", profile, ignores)
				addMutation(mutants, fset, pkg, filename, src, fn, ident, "returns", "true", "false", profile, ignores)
			}
			if ident.Name == "false" {
				addMutation(mutants, fset, pkg, filename, src, fn, ident, "return-bool-literals", "false", "true", profile, ignores)
				addMutation(mutants, fset, pkg, filename, src, fn, ident, "returns", "false", "true", profile, ignores)
			}
		}
	case *ast.ForStmt:
		if expr, ok := n.Cond.(*ast.BinaryExpr); ok {
			addLoopControlMutant(mutants, fset, pkg, filename, src, fn, expr, profile, ignores)
		}
	}
}

func addBinaryMutants(mutants *[]Mutant, fset *token.FileSet, pkg, filename string, src []byte, fn string, expr *ast.BinaryExpr, profile string, ignores []inlineIgnore) {
	type candidate struct {
		operator    string
		replacement string
	}
	var candidates []candidate
	switch expr.Op {
	case token.EQL:
		candidates = append(candidates, candidate{operator: "conditionals-negation", replacement: "!="})
	case token.NEQ:
		candidates = append(candidates, candidate{operator: "conditionals-negation", replacement: "=="})
	case token.LSS:
		candidates = append(candidates,
			candidate{operator: "conditionals-boundary", replacement: "<="},
			candidate{operator: "conditionals-negation", replacement: ">="},
		)
	case token.LEQ:
		candidates = append(candidates,
			candidate{operator: "conditionals-boundary", replacement: "<"},
			candidate{operator: "conditionals-negation", replacement: ">"},
		)
	case token.GTR:
		candidates = append(candidates,
			candidate{operator: "conditionals-boundary", replacement: ">="},
			candidate{operator: "conditionals-negation", replacement: "<="},
		)
	case token.GEQ:
		candidates = append(candidates,
			candidate{operator: "conditionals-boundary", replacement: ">"},
			candidate{operator: "conditionals-negation", replacement: "<"},
		)
	case token.LAND:
		candidates = append(candidates, candidate{operator: "logical", replacement: "||"})
	case token.LOR:
		candidates = append(candidates, candidate{operator: "logical", replacement: "&&"})
	case token.ADD:
		candidates = append(candidates, candidate{operator: "arithmetic-basic", replacement: "-"})
	case token.SUB:
		candidates = append(candidates, candidate{operator: "arithmetic-basic", replacement: "+"})
	case token.MUL:
		candidates = append(candidates, candidate{operator: "arithmetic-basic", replacement: "/"})
	case token.QUO:
		candidates = append(candidates, candidate{operator: "arithmetic-basic", replacement: "*"})
	}
	if len(candidates) == 0 {
		return
	}
	if isNilCheck(expr) {
		candidates = []candidate{{operator: "nil-checks", replacement: candidates[0].replacement}}
	}
	if isLenComparison(expr) {
		candidates = append(candidates, candidate{operator: "slice-map-len-boundary", replacement: boundaryReplacement(expr.Op)})
	}
	for _, mutation := range candidates {
		if mutation.replacement == "" {
			continue
		}
		addMutation(mutants, fset, pkg, filename, src, fn, expr, mutation.operator, expr.Op.String(), mutation.replacement, profile, ignores)
	}
}

func addLoopControlMutant(mutants *[]Mutant, fset *token.FileSet, pkg, filename string, src []byte, fn string, expr *ast.BinaryExpr, profile string, ignores []inlineIgnore) {
	replacement := boundaryReplacement(expr.Op)
	if replacement == "" {
		return
	}
	addMutation(mutants, fset, pkg, filename, src, fn, expr, "loop-control", expr.Op.String(), replacement, profile, ignores)
}

func boundaryReplacement(op token.Token) string {
	switch op {
	case token.LSS:
		return "<="
	case token.LEQ:
		return "<"
	case token.GTR:
		return ">="
	case token.GEQ:
		return ">"
	default:
		return ""
	}
}

func isLenComparison(expr *ast.BinaryExpr) bool {
	return isLenCall(expr.X) || isLenCall(expr.Y)
}

func isLenCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "len"
}

func isNilCheck(expr *ast.BinaryExpr) bool {
	if expr.Op != token.EQL && expr.Op != token.NEQ {
		return false
	}
	return isNil(expr.X) || isNil(expr.Y)
}

func isNil(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func addMutation(mutants *[]Mutant, fset *token.FileSet, pkg, filename string, src []byte, fn string, node ast.Node, operator, original, mutated, profile string, ignores []inlineIgnore) {
	if !operatorEnabled(operator, profile) {
		return
	}
	pos := fset.Position(node.Pos())
	if ignored(pos.Line, operator, ignores) {
		return
	}
	start := fset.Position(node.Pos()).Offset
	end := fset.Position(node.End()).Offset
	if start < 0 || end > len(src) || start >= end {
		return
	}
	mutatedSrc := append([]byte{}, src...)
	segment := string(src[start:end])
	next, err := replaceFirst(segment, original, mutated)
	if err != nil {
		return
	}
	mutatedSrc = append(mutatedSrc[:start], append([]byte(next), mutatedSrc[end:]...)...)
	diff := unifiedDiff(filename, string(src), string(mutatedSrc))
	fp := fingerprint(filename, strconv.Itoa(pos.Line), strconv.Itoa(start), strconv.Itoa(end), operator, original, mutated, diff)
	id := fmt.Sprintf("%s:%d:%s:%s", filename, pos.Line, operator, fp[:12])
	*mutants = append(*mutants, Mutant{
		ID:               id,
		Package:          pkg,
		File:             filename,
		Line:             pos.Line,
		Function:         fn,
		Operator:         operator,
		Original:         original,
		Mutated:          mutated,
		StartOffset:      start,
		EndOffset:        end,
		Diff:             diff,
		Fingerprint:      fp,
		Hint:             hint(operator),
		Description:      description(fn, operator, original, mutated),
		EquivalentRisk:   equivalentRisk(operator),
		Recommendation:   recommendation(operator),
		CompileErrorRisk: compileErrorRisk(operator),
	})
}

func operatorEnabled(operator, profile string) bool {
	switch operator {
	case "conditionals-negation", "conditionals-boundary", "arithmetic-basic":
		return true
	case "logical", "boolean-literals", "string-empty-literals":
		return profile == ProfileConservative || profile == ProfileDefault || profile == ProfileAggressive
	case "nil-checks", "error-returns", "numeric-literals", "return-bool-literals":
		return profile == ProfileDefault || profile == ProfileAggressive
	case "literals", "returns", "loop-control", "slice-map-len-boundary":
		return profile == ProfileAggressive
	default:
		return false
	}
}

func equivalentRisk(operator string) string {
	switch operator {
	case "arithmetic-basic", "boolean-literals":
		return "low"
	case "conditionals-negation", "conditionals-boundary", "logical", "string-empty-literals", "numeric-literals", "return-bool-literals", "slice-map-len-boundary":
		return "medium"
	case "nil-checks", "error-returns", "literals", "returns", "loop-control":
		return "high"
	default:
		return "unknown"
	}
}

func recommendation(operator string) string {
	switch operator {
	case "arithmetic-basic", "conditionals-negation", "conditionals-boundary":
		return "fast-ci"
	case "logical", "boolean-literals", "string-empty-literals":
		return "conservative"
	case "nil-checks", "error-returns", "numeric-literals", "return-bool-literals":
		return "default"
	case "literals", "returns", "loop-control":
		return "aggressive"
	default:
		return "review"
	}
}

func compileErrorRisk(operator string) string {
	switch operator {
	case "conditionals-negation", "conditionals-boundary", "logical", "boolean-literals", "nil-checks", "string-empty-literals", "numeric-literals", "return-bool-literals", "slice-map-len-boundary":
		return "low"
	case "arithmetic-basic", "error-returns", "literals", "returns", "loop-control":
		return "medium"
	default:
		return "unknown"
	}
}

func ignored(line int, operator string, ignores []inlineIgnore) bool {
	for _, ignore := range ignores {
		if ignore.line == line && operatorMatchesIgnore(operator, ignore.operator) {
			return true
		}
	}
	return false
}

func operatorMatchesIgnore(operator, ignored string) bool {
	if ignored == "*" || ignored == operator {
		return true
	}
	if ignored == "conditionals" && strings.HasPrefix(operator, "conditionals-") {
		return true
	}
	return false
}

func replaceFirst(s, old, new string) (string, error) {
	idx := strings.Index(s, old)
	if idx < 0 {
		return "", errors.New("original token not found")
	}
	return s[:idx] + new + s[idx+len(old):], nil
}

func unifiedDiff(filename, original, mutated string) string {
	return fmt.Sprintf("--- %s\n+++ %s\n@@\n-%s\n+%s\n", filename, filename, strings.TrimRight(original, "\n"), strings.TrimRight(mutated, "\n"))
}

func fingerprint(parts ...string) string {
	hash := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(hash[:])
}

func hint(operator string) string {
	switch operator {
	case "conditionals-negation", "conditionals-boundary", "nil-checks":
		return "Add assertions for the opposite branch or boundary condition."
	case "logical":
		return "Add a test where only one side of the boolean expression changes the outcome."
	case "boolean-literals", "return-bool-literals":
		return "Assert both boolean outcomes instead of only executing the path."
	case "arithmetic-basic":
		return "Add assertions for the computed numeric result and edge cases."
	case "string-empty-literals":
		return "Add an assertion for non-empty text or exact message behavior."
	case "numeric-literals":
		return "Add an assertion for numeric boundaries and configured constants."
	case "slice-map-len-boundary":
		return "Add tests for empty and single-element collection boundaries."
	case "loop-control":
		return "Add tests for loop boundary counts and off-by-one behavior."
	default:
		return "Add an assertion that observes the changed behavior."
	}
}

func description(fn, operator, original, mutated string) string {
	where := "expression"
	if fn != "" {
		where = "function " + fn
	}
	return fmt.Sprintf("%s mutation in %s: changed %s to %s.", operator, where, original, mutated)
}

func FormatNode(fset *token.FileSet, node ast.Node) string {
	var b strings.Builder
	_ = printer.Fprint(&b, fset, node)
	return b.String()
}
