package parser_test

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/parser"
)

type FeatureStatus string

const (
	StatusSupported   FeatureStatus = "SUPPORTED"
	StatusPartial     FeatureStatus = "PARTIAL"
	StatusUnsupported FeatureStatus = "UNSUPPORTED"
	StatusUnknown     FeatureStatus = "UNKNOWN"
)

type FeatureCategory string

const (
	CategoryBasicSyntax    FeatureCategory = "基础语法"
	CategoryDataTypes      FeatureCategory = "数据类型"
	CategoryNodePattern    FeatureCategory = "节点模式"
	CategoryRelationship   FeatureCategory = "关系模式"
	CategoryPattern        FeatureCategory = "模式匹配"
	CategoryMatchClause    FeatureCategory = "MATCH子句"
	CategoryWhereClause    FeatureCategory = "WHERE子句"
	CategoryReturnClause   FeatureCategory = "RETURN子句"
	CategoryWithClause     FeatureCategory = "WITH子句"
	CategoryCreateClause   FeatureCategory = "CREATE子句"
	CategoryMergeClause    FeatureCategory = "MERGE子句"
	CategorySetClause      FeatureCategory = "SET子句"
	CategoryDeleteClause   FeatureCategory = "DELETE子句"
	CategoryRemoveClause   FeatureCategory = "REMOVE子句"
	CategoryUnwindClause   FeatureCategory = "UNWIND子句"
	CategoryUnionClause    FeatureCategory = "UNION子句"
	CategoryExpressions    FeatureCategory = "表达式"
	CategoryFunctions      FeatureCategory = "函数"
	CategoryAdvanced       FeatureCategory = "高级特性"
)

type TestCase struct {
	Name        string
	Input       string
	Description string
	ExpectedOK  bool
}

type FeatureResult struct {
	Category    FeatureCategory
	Name        string
	Description string
	Status      FeatureStatus
	TestCases   []TestCase
	Passed      int
	Failed      int
	Notes       string
}

type ComplianceReport struct {
	TotalFeatures   int
	SupportedCount  int
	PartialCount    int
	UnsupportedCount int
	UnknownCount    int
	Results         []FeatureResult
}

var complianceReport ComplianceReport

func runTestCases(t *testing.T, category FeatureCategory, name string, testCases []TestCase) FeatureResult {
	result := FeatureResult{
		Category:  category,
		Name:      name,
		TestCases: testCases,
	}

	for _, tc := range testCases {
		p := parser.New(tc.Input)
		_, err := p.Parse()
		
		passed := (err == nil) == tc.ExpectedOK
		if passed {
			result.Passed++
		} else {
			result.Failed++
			if result.Notes != "" {
				result.Notes += "; "
			}
			result.Notes += fmt.Sprintf("%s: %v", tc.Name, err)
		}
	}

	if result.Failed == 0 {
		result.Status = StatusSupported
	} else if result.Passed > 0 {
		result.Status = StatusPartial
	} else {
		result.Status = StatusUnsupported
	}

	complianceReport.Results = append(complianceReport.Results, result)
	return result
}

func TestCompliance_BasicSyntax(t *testing.T) {
	t.Run("关键字大小写不敏感", func(t *testing.T) {
		runTestCases(t, CategoryBasicSyntax, "关键字大小写不敏感", []TestCase{
			{"uppercase", "MATCH (n:Person) RETURN n", "关键字全大写", true},
			{"lowercase", "match (n:Person) return n", "关键字全小写", true},
			{"mixed", "Match (n:Person) Return n", "关键字混合大小写", true},
		})
	})

	t.Run("标识符命名规则", func(t *testing.T) {
		runTestCases(t, CategoryBasicSyntax, "标识符命名规则", []TestCase{
			{"letter_start", "MATCH (n:Person) RETURN n", "字母开头", true},
			{"underscore_start", "MATCH (_n:Person) RETURN _n", "下划线开头", true},
			{"with_numbers", "MATCH (n1:Person) RETURN n1", "包含数字", true},
			{"with_dollar", "MATCH ($n:Person) RETURN $n", "包含$符号", false},
		})
	})

	t.Run("注释语法", func(t *testing.T) {
		runTestCases(t, CategoryBasicSyntax, "注释语法", []TestCase{
			{"single_line", "MATCH // comment\n(n:Person) RETURN n", "单行注释", true},
			{"multi_line", "MATCH /* multi\nline */ (n:Person) RETURN n", "多行注释", true},
			{"end_of_line", "MATCH (n:Person) RETURN n // end comment", "行尾注释", true},
		})
	})
}

func TestCompliance_DataTypes(t *testing.T) {
	t.Run("标量类型", func(t *testing.T) {
		runTestCases(t, CategoryDataTypes, "标量类型", []TestCase{
			{"integer", "RETURN 12345", "整数", true},
			{"negative_integer", "RETURN -5", "负整数", true},
			{"large_integer", "RETURN 9223372036854775807", "大整数", true},
			{"float", "RETURN 3.14", "浮点数", true},
			{"float_negative", "RETURN -0.5", "负浮点数", true},
			{"scientific", "RETURN 1.5e10", "科学计数法", true},
			{"scientific_negative", "RETURN 1.5e-10", "负指数科学计数法", true},
			{"string_single", "RETURN 'Alice'", "单引号字符串", true},
			{"string_double", `RETURN "Alice"`, "双引号字符串", true},
			{"string_escape", `RETURN "It\'s a test"`, "转义字符串", true},
			{"boolean_true", "RETURN true", "布尔值true", true},
			{"boolean_false", "RETURN false", "布尔值false", true},
			{"null", "RETURN null", "空值", true},
		})
	})

	t.Run("时间类型", func(t *testing.T) {
		runTestCases(t, CategoryDataTypes, "时间类型", []TestCase{
			{"date", "RETURN date()", "日期函数", true},
			{"date_string", `RETURN date("2026-03-28")`, "日期字符串", true},
			{"datetime", "RETURN datetime()", "日期时间函数", true},
			{"datetime_string", `RETURN datetime("2026-03-28T10:30:00")`, "日期时间字符串", true},
			{"time", "RETURN time()", "时间函数", true},
			{"localdatetime", "RETURN localdatetime()", "本地日期时间", true},
			{"localtime", "RETURN localtime()", "本地时间", true},
			{"duration", `RETURN duration({days:1, hours:2})`, "时长", true},
		})
	})

	t.Run("复合类型", func(t *testing.T) {
		runTestCases(t, CategoryDataTypes, "复合类型", []TestCase{
			{"list_empty", "RETURN []", "空列表", true},
			{"list_integers", "RETURN [1, 2, 3]", "整数列表", true},
			{"list_strings", `RETURN ["Alice", "Bob"]`, "字符串列表", true},
			{"list_mixed", "RETURN [1, 'two', true, null]", "混合类型列表", true},
			{"list_nested", "RETURN [[1, 2], [3, 4]]", "嵌套列表", true},
			{"map_empty", "RETURN {}", "空映射", true},
			{"map_simple", "RETURN {name: 'Alice', age: 30}", "简单映射", true},
			{"map_nested", "RETURN {person: {name: 'Alice'}}", "嵌套映射", true},
		})
	})
}

func TestCompliance_NodePattern(t *testing.T) {
	t.Run("节点模式", func(t *testing.T) {
		runTestCases(t, CategoryNodePattern, "节点模式", []TestCase{
			{"empty", "MATCH () RETURN 1", "空节点", true},
			{"variable_only", "MATCH (n) RETURN n", "仅变量", true},
			{"single_label", "MATCH (n:Person) RETURN n", "单个标签", true},
			{"multiple_labels", "MATCH (n:Person:Employee:Manager) RETURN n", "多个标签", true},
			{"properties_only", "MATCH (n {name: 'Alice'}) RETURN n", "仅属性", true},
			{"label_and_properties", "MATCH (n:Person {name: 'Alice'}) RETURN n", "标签和属性", true},
			{"full_node", "MATCH (p:Person:Employee {name: 'Alice', age: 30, active: true}) RETURN p", "完整节点", true},
			{"param_properties", "MATCH (n {name: $name}) RETURN n", "参数化属性", true},
		})
	})
}

func TestCompliance_RelationshipPattern(t *testing.T) {
	t.Run("关系方向", func(t *testing.T) {
		runTestCases(t, CategoryRelationship, "关系方向", []TestCase{
			{"undirected", "MATCH (a)--(b) RETURN a, b", "无向关系", true},
			{"outgoing", "MATCH (a)-->(b) RETURN a, b", "出关系", true},
			{"incoming", "MATCH (a)<--(b) RETURN a, b", "入关系", true},
			{"left_arrow", "MATCH (a)<-(b) RETURN a, b", "左箭头简写", true},
			{"right_arrow", "MATCH (a)->(b) RETURN a, b", "右箭头简写", true},
		})
	})

	t.Run("关系类型", func(t *testing.T) {
		runTestCases(t, CategoryRelationship, "关系类型", []TestCase{
			{"with_type", "MATCH (a)-[:KNOWS]->(b) RETURN a, b", "带类型", true},
			{"with_variable", "MATCH (a)-[r]->(b) RETURN r", "带变量", true},
			{"variable_and_type", "MATCH (a)-[r:KNOWS]->(b) RETURN r", "变量和类型", true},
		})
	})

	t.Run("关系属性", func(t *testing.T) {
		runTestCases(t, CategoryRelationship, "关系属性", []TestCase{
			{"with_properties", "MATCH (a)-[:KNOWS {since: 2020}]->(b) RETURN a, b", "带属性", true},
			{"full_relationship", "MATCH (a)-[r:KNOWS {since: 2020, status: 'close'}]->(b) RETURN r", "完整关系", true},
		})
	})

	t.Run("变长路径", func(t *testing.T) {
		runTestCases(t, CategoryRelationship, "变长路径", []TestCase{
			{"var_length_min", "MATCH (a)-[:KNOWS*1]->(b) RETURN a, b", "最小长度", true},
			{"var_length_range", "MATCH (a)-[:KNOWS*1..5]->(b) RETURN a, b", "范围长度", true},
			{"var_length_unbounded", "MATCH (a)-[:KNOWS*]->(b) RETURN a, b", "无界长度", true},
			{"var_length_with_props", "MATCH (a)-[:KNOWS*1..3 {active: true}]->(b) RETURN a, b", "带属性变长", true},
		})
	})
}

func TestCompliance_Pattern(t *testing.T) {
	t.Run("路径模式", func(t *testing.T) {
		runTestCases(t, CategoryPattern, "路径模式", []TestCase{
			{"simple_path", "MATCH (a:Person)-[:KNOWS]->(b:Person) RETURN a, b", "简单路径", true},
			{"multi_hop", "MATCH (a)-[:KNOWS]->(b)-[:FRIENDS_WITH]->(c) RETURN a, c", "多跳路径", true},
			{"multiple_patterns", "MATCH (a:Person), (b:Company) RETURN a, b", "多模式", true},
			{"path_variable", "MATCH path = (a)-[:KNOWS]->(b) RETURN path", "路径变量", true},
			{"complex_path", "MATCH (a:Person {name: 'Alice'})-[:WORKS_FOR]->(c:Company)<-[:WORKS_FOR]-(b:Person) RETURN a, b, c", "复杂路径", true},
		})
	})
}

func TestCompliance_MatchClause(t *testing.T) {
	t.Run("MATCH基本语法", func(t *testing.T) {
		runTestCases(t, CategoryMatchClause, "MATCH基本语法", []TestCase{
			{"basic", "MATCH (n:Person) RETURN n", "基本MATCH", true},
			{"optional", "OPTIONAL MATCH (n:Person) RETURN n", "OPTIONAL MATCH", true},
			{"with_where", "MATCH (n:Person) WHERE n.age > 18 RETURN n", "带WHERE", true},
			{"with_return", "MATCH (n:Person) RETURN n.name, n.age", "带RETURN", true},
		})
	})

	t.Run("MATCH后续子句", func(t *testing.T) {
		runTestCases(t, CategoryMatchClause, "MATCH后续子句", []TestCase{
			{"with_delete", "MATCH (n:Person) DELETE n", "带DELETE", true},
			{"with_detach_delete", "MATCH (n:Person) DETACH DELETE n", "带DETACH DELETE", true},
			{"with_set", "MATCH (n:Person) SET n.updated = true", "带SET", true},
			{"with_remove", "MATCH (n:Person) REMOVE n.temp", "带REMOVE", true},
		})
	})
}

func TestCompliance_WhereClause(t *testing.T) {
	t.Run("比较运算", func(t *testing.T) {
		runTestCases(t, CategoryWhereClause, "比较运算", []TestCase{
			{"equals", "MATCH (n) WHERE n.name = 'Alice' RETURN n", "等于", true},
			{"not_equals", "MATCH (n) WHERE n.name != 'Bob' RETURN n", "不等于", true},
			{"greater", "MATCH (n) WHERE n.age > 18 RETURN n", "大于", true},
			{"greater_equal", "MATCH (n) WHERE n.age >= 18 RETURN n", "大于等于", true},
			{"less", "MATCH (n) WHERE n.age < 65 RETURN n", "小于", true},
			{"less_equal", "MATCH (n) WHERE n.age <= 65 RETURN n", "小于等于", true},
		})
	})

	t.Run("逻辑运算", func(t *testing.T) {
		runTestCases(t, CategoryWhereClause, "逻辑运算", []TestCase{
			{"and", "MATCH (n) WHERE n.age > 18 AND n.active = true RETURN n", "AND", true},
			{"or", "MATCH (n) WHERE n.city = 'Beijing' OR n.city = 'Shanghai' RETURN n", "OR", true},
			{"not", "MATCH (n) WHERE NOT n.deleted RETURN n", "NOT", true},
			{"xor", "MATCH (n) WHERE n.a XOR n.b RETURN n", "XOR", true},
			{"complex", "MATCH (n) WHERE (n.age > 18 OR n.vip = true) AND n.active = true RETURN n", "复杂条件", true},
		})
	})

	t.Run("字符串匹配", func(t *testing.T) {
		runTestCases(t, CategoryWhereClause, "字符串匹配", []TestCase{
			{"contains", "MATCH (n) WHERE n.name CONTAINS 'Li' RETURN n", "CONTAINS", true},
			{"starts_with", "MATCH (n) WHERE n.name STARTS WITH 'A' RETURN n", "STARTS WITH", true},
			{"ends_with", "MATCH (n) WHERE n.name ENDS WITH 'e' RETURN n", "ENDS WITH", true},
			{"matches", "MATCH (n) WHERE n.name =~ 'A.*' RETURN n", "正则匹配", true},
		})
	})

	t.Run("空值判断", func(t *testing.T) {
		runTestCases(t, CategoryWhereClause, "空值判断", []TestCase{
			{"is_null", "MATCH (n) WHERE n.email IS NULL RETURN n", "IS NULL", true},
			{"is_not_null", "MATCH (n) WHERE n.email IS NOT NULL RETURN n", "IS NOT NULL", true},
		})
	})

	t.Run("列表操作", func(t *testing.T) {
		runTestCases(t, CategoryWhereClause, "列表操作", []TestCase{
			{"in_list", "MATCH (n) WHERE n.city IN ['Beijing', 'Shanghai'] RETURN n", "IN列表", true},
			{"exists", "MATCH (n) WHERE EXISTS(n.email) RETURN n", "EXISTS", true},
		})
	})
}

func TestCompliance_ReturnClause(t *testing.T) {
	t.Run("RETURN基本语法", func(t *testing.T) {
		runTestCases(t, CategoryReturnClause, "RETURN基本语法", []TestCase{
			{"single", "MATCH (n) RETURN n.name", "单个返回项", true},
			{"multiple", "MATCH (n) RETURN n.name, n.age, n.city", "多个返回项", true},
			{"alias", "MATCH (n) RETURN n.name AS userName", "别名", true},
			{"distinct", "MATCH (n) RETURN DISTINCT n.city", "DISTINCT", true},
		})
	})

	t.Run("排序分页", func(t *testing.T) {
		runTestCases(t, CategoryReturnClause, "排序分页", []TestCase{
			{"order_asc", "MATCH (n) RETURN n ORDER BY n.name ASC", "升序排序", true},
			{"order_desc", "MATCH (n) RETURN n ORDER BY n.age DESC", "降序排序", true},
			{"order_multiple", "MATCH (n) RETURN n ORDER BY n.age DESC, n.name ASC", "多字段排序", true},
			{"skip", "MATCH (n) RETURN n SKIP 10", "SKIP", true},
			{"limit", "MATCH (n) RETURN n LIMIT 5", "LIMIT", true},
			{"skip_limit", "MATCH (n) RETURN n SKIP 10 LIMIT 5", "SKIP和LIMIT", true},
			{"order_skip_limit", "MATCH (n) RETURN n ORDER BY n.name SKIP 10 LIMIT 5", "完整分页", true},
		})
	})

	t.Run("表达式返回", func(t *testing.T) {
		runTestCases(t, CategoryReturnClause, "表达式返回", []TestCase{
			{"arithmetic", "MATCH (n) RETURN n.age * 2 AS doubleAge", "算术表达式", true},
			{"function", "MATCH (n) RETURN COUNT(n) AS total", "函数调用", true},
			{"case", "MATCH (n) RETURN CASE WHEN n.age < 18 THEN 'minor' ELSE 'adult' END AS status", "CASE表达式", true},
		})
	})
}

func TestCompliance_WithClause(t *testing.T) {
	t.Run("WITH基本语法", func(t *testing.T) {
		runTestCases(t, CategoryWithClause, "WITH基本语法", []TestCase{
			{"basic", "MATCH (n) WITH n RETURN n", "基本WITH", true},
			{"alias", "MATCH (n) WITH n AS node RETURN node", "别名", true},
			{"distinct", "MATCH (n) WITH DISTINCT n.city AS city RETURN city", "DISTINCT", true},
		})
	})

	t.Run("WITH聚合", func(t *testing.T) {
		runTestCases(t, CategoryWithClause, "WITH聚合", []TestCase{
			{"aggregation", "MATCH (n)-[r]->(m) WITH n, COUNT(r) AS relCount RETURN n, relCount", "聚合", true},
			{"collect", "MATCH (n) WITH n, COLLECT(n.name) AS names RETURN names", "COLLECT", true},
		})
	})

	t.Run("WITH过滤排序", func(t *testing.T) {
		runTestCases(t, CategoryWithClause, "WITH过滤排序", []TestCase{
			{"where", "MATCH (n) WITH n WHERE n.age > 18 RETURN n", "WHERE过滤", true},
			{"order_by", "MATCH (n) WITH n ORDER BY n.name RETURN n", "ORDER BY", true},
			{"skip_limit", "MATCH (n) WITH n SKIP 5 LIMIT 10 RETURN n", "SKIP LIMIT", true},
		})
	})
}

func TestCompliance_CreateClause(t *testing.T) {
	t.Run("CREATE基本语法", func(t *testing.T) {
		runTestCases(t, CategoryCreateClause, "CREATE基本语法", []TestCase{
			{"single_node", "CREATE (n:Person {name: 'Alice'})", "单个节点", true},
			{"multiple_nodes", "CREATE (a:Person), (b:Company)", "多个节点", true},
			{"multi_label", "CREATE (n:Person:Employee {name: 'Alice'})", "多标签节点", true},
		})
	})

	t.Run("CREATE关系", func(t *testing.T) {
		runTestCases(t, CategoryCreateClause, "CREATE关系", []TestCase{
			{"with_relationship", "CREATE (a:Person)-[:KNOWS]->(b:Person)", "创建关系", true},
			{"complex", "CREATE (a:Person {name: 'Alice'})-[:WORKS_FOR {since: 2020}]->(c:Company {name: 'TechCorp'})", "复杂创建", true},
		})
	})

	t.Run("CREATE RETURN", func(t *testing.T) {
		runTestCases(t, CategoryCreateClause, "CREATE RETURN", []TestCase{
			{"return_node", "CREATE (n:Person {name: 'Alice'}) RETURN n", "返回创建的节点", true},
			{"return_multiple", "CREATE (a:Person), (b:Company) RETURN a, b", "返回多个", true},
		})
	})
}

func TestCompliance_MergeClause(t *testing.T) {
	t.Run("MERGE基本语法", func(t *testing.T) {
		runTestCases(t, CategoryMergeClause, "MERGE基本语法", []TestCase{
			{"basic", "MERGE (n:Person {name: 'Alice'})", "基本MERGE", true},
			{"relationship", "MERGE (a)-[:KNOWS]->(b)", "MERGE关系", true},
		})
	})

	t.Run("ON CREATE/ON MATCH", func(t *testing.T) {
		runTestCases(t, CategoryMergeClause, "ON CREATE/ON MATCH", []TestCase{
			{"on_create", "MERGE (n:Person {name: 'Alice'}) ON CREATE SET n.createdAt = timestamp()", "ON CREATE", true},
			{"on_match", "MERGE (n:Person {name: 'Alice'}) ON MATCH SET n.updatedAt = timestamp()", "ON MATCH", true},
			{"both", "MERGE (n:Person {name: 'Alice'}) ON CREATE SET n.new = true ON MATCH SET n.existing = true", "两者都有", true},
		})
	})
}

func TestCompliance_SetClause(t *testing.T) {
	t.Run("SET属性", func(t *testing.T) {
		runTestCases(t, CategorySetClause, "SET属性", []TestCase{
			{"single", "MATCH (n) SET n.age = 30", "单个属性", true},
			{"multiple", "MATCH (n) SET n.age = 30, n.city = 'Beijing'", "多个属性", true},
			{"plus_equals", "MATCH (n) SET n.tags += ['new']", "+=操作符", true},
		})
	})

	t.Run("SET标签", func(t *testing.T) {
		runTestCases(t, CategorySetClause, "SET标签", []TestCase{
			{"add_label", "MATCH (n) SET n:Employee", "添加标签", true},
			{"multiple_labels", "MATCH (n) SET n:Employee:Manager", "多个标签", true},
		})
	})
}

func TestCompliance_DeleteClause(t *testing.T) {
	t.Run("DELETE基本语法", func(t *testing.T) {
		runTestCases(t, CategoryDeleteClause, "DELETE基本语法", []TestCase{
			{"node", "MATCH (n) DELETE n", "删除节点", true},
			{"detach", "MATCH (n) DETACH DELETE n", "DETACH DELETE", true},
			{"relationship", "MATCH ()-[r]->() DELETE r", "删除关系", true},
			{"multiple", "MATCH (a)-[r]->(b) DELETE a, r, b", "删除多个", true},
		})
	})
}

func TestCompliance_RemoveClause(t *testing.T) {
	t.Run("REMOVE基本语法", func(t *testing.T) {
		runTestCases(t, CategoryRemoveClause, "REMOVE基本语法", []TestCase{
			{"property", "MATCH (n) REMOVE n.temp", "删除属性", true},
			{"multiple_properties", "MATCH (n) REMOVE n.temp, n.flag", "删除多个属性", true},
			{"label", "MATCH (n) REMOVE n:Employee", "删除标签", true},
		})
	})
}

func TestCompliance_UnwindClause(t *testing.T) {
	t.Run("UNWIND基本语法", func(t *testing.T) {
		runTestCases(t, CategoryUnwindClause, "UNWIND基本语法", []TestCase{
			{"list", "UNWIND [1, 2, 3] AS n RETURN n", "展开列表", true},
			{"operation", "UNWIND [1, 2, 3] AS n RETURN n * 2", "带操作", true},
			{"property", "MATCH (n) UNWIND n.tags AS tag RETURN tag", "展开属性", true},
		})
	})
}

func TestCompliance_UnionClause(t *testing.T) {
	t.Run("UNION基本语法", func(t *testing.T) {
		runTestCases(t, CategoryUnionClause, "UNION基本语法", []TestCase{
			{"union", "MATCH (p:Person) RETURN p.name AS name UNION MATCH (c:Company) RETURN c.name AS name", "UNION", true},
			{"union_all", "MATCH (p:Person {city: 'Beijing'}) RETURN p.name AS name UNION ALL MATCH (p:Person {city: 'Shanghai'}) RETURN p.name AS name", "UNION ALL", true},
		})
	})
}

func TestCompliance_Expressions(t *testing.T) {
	t.Run("算术表达式", func(t *testing.T) {
		runTestCases(t, CategoryExpressions, "算术表达式", []TestCase{
			{"add", "RETURN 1 + 2", "加法", true},
			{"subtract", "RETURN 5 - 3", "减法", true},
			{"multiply", "RETURN 2 * 3", "乘法", true},
			{"divide", "RETURN 6 / 2", "除法", true},
			{"modulo", "RETURN 10 % 3", "取模", true},
			{"power", "RETURN 2 ^ 10", "幂运算", true},
			{"unary_minus", "RETURN -5", "一元负号", true},
			{"parentheses", "RETURN (1 + 2) * 3", "括号", true},
		})
	})

	t.Run("属性访问", func(t *testing.T) {
		runTestCases(t, CategoryExpressions, "属性访问", []TestCase{
			{"simple", "MATCH (n) RETURN n.name", "简单属性", true},
			{"nested", "MATCH (n) RETURN n.address.city", "嵌套属性", true},
		})
	})

	t.Run("列表操作", func(t *testing.T) {
		runTestCases(t, CategoryExpressions, "列表操作", []TestCase{
			{"index", "RETURN [1, 2, 3][0]", "索引访问", true},
			{"slice", "RETURN [1, 2, 3, 4, 5][1..3]", "切片", true},
			{"negative_index", "RETURN [1, 2, 3][-1]", "负索引", true},
		})
	})

	t.Run("CASE表达式", func(t *testing.T) {
		runTestCases(t, CategoryExpressions, "CASE表达式", []TestCase{
			{"simple", "RETURN CASE n.age WHEN 18 THEN 'adult' ELSE 'other' END", "简单CASE", true},
			{"searched", "RETURN CASE WHEN n.age < 18 THEN 'minor' WHEN n.age >= 18 THEN 'adult' END", "搜索CASE", true},
			{"with_else", "RETURN CASE WHEN n.age < 18 THEN 'minor' ELSE 'adult' END", "带ELSE", true},
		})
	})

	t.Run("参数", func(t *testing.T) {
		runTestCases(t, CategoryExpressions, "参数", []TestCase{
			{"param", "MATCH (n) WHERE n.name = $name RETURN n", "命名参数", true},
			{"param_expr", "MATCH (n) WHERE n.age > $minAge RETURN n", "参数表达式", true},
		})
	})
}

func TestCompliance_Functions(t *testing.T) {
	t.Run("聚合函数", func(t *testing.T) {
		runTestCases(t, CategoryFunctions, "聚合函数", []TestCase{
			{"count", "MATCH (n) RETURN COUNT(n)", "COUNT", true},
			{"count_star", "MATCH (n) RETURN COUNT(*)", "COUNT(*)", true},
			{"count_distinct", "MATCH (n) RETURN COUNT(DISTINCT n.city)", "COUNT DISTINCT", true},
			{"sum", "MATCH (n) RETURN SUM(n.age)", "SUM", true},
			{"avg", "MATCH (n) RETURN AVG(n.age)", "AVG", true},
			{"min", "MATCH (n) RETURN MIN(n.age)", "MIN", true},
			{"max", "MATCH (n) RETURN MAX(n.age)", "MAX", true},
			{"collect", "MATCH (n) RETURN COLLECT(n.name)", "COLLECT", true},
		})
	})

	t.Run("数学函数", func(t *testing.T) {
		runTestCases(t, CategoryFunctions, "数学函数", []TestCase{
			{"abs", "RETURN ABS(-5)", "ABS", true},
			{"ceil", "RETURN CEIL(3.14)", "CEIL", true},
			{"floor", "RETURN FLOOR(3.99)", "FLOOR", true},
			{"round", "RETURN ROUND(3.5)", "ROUND", true},
			{"rand", "RETURN RAND()", "RAND", true},
			{"sign", "RETURN SIGN(-10)", "SIGN", true},
		})
	})

	t.Run("字符串函数", func(t *testing.T) {
		runTestCases(t, CategoryFunctions, "字符串函数", []TestCase{
			{"length", "RETURN LENGTH('Alice')", "LENGTH", true},
			{"toupper", "RETURN TOUPPER('hello')", "TOUPPER", true},
			{"tolower", "RETURN TOLOWER('HELLO')", "TOLOWER", true},
			{"replace", "RETURN REPLACE('hello', 'l', 'L')", "REPLACE", true},
			{"substring", "RETURN SUBSTRING('hello', 1, 3)", "SUBSTRING", true},
			{"trim", "RETURN TRIM('  hello  ')", "TRIM", true},
			{"ltrim", "RETURN LTRIM('  hello')", "LTRIM", true},
			{"rtrim", "RETURN RTRIM('hello  ')", "RTRIM", true},
			{"left", "RETURN LEFT('hello', 3)", "LEFT", true},
			{"right", "RETURN RIGHT('hello', 3)", "RIGHT", true},
			{"split", "RETURN SPLIT('a,b,c', ',')", "SPLIT", true},
		})
	})

	t.Run("列表函数", func(t *testing.T) {
		runTestCases(t, CategoryFunctions, "列表函数", []TestCase{
			{"head", "RETURN HEAD([1, 2, 3])", "HEAD", true},
			{"last", "RETURN LAST([1, 2, 3])", "LAST", true},
			{"tail", "RETURN TAIL([1, 2, 3])", "TAIL", true},
			{"size", "RETURN SIZE([1, 2, 3])", "SIZE", true},
			{"range", "RETURN RANGE(1, 10)", "RANGE", true},
			{"reverse", "RETURN REVERSE([1, 2, 3])", "REVERSE", true},
		})
	})

	t.Run("图函数", func(t *testing.T) {
		runTestCases(t, CategoryFunctions, "图函数", []TestCase{
			{"id", "MATCH (n) RETURN ID(n)", "ID", true},
			{"labels", "MATCH (n) RETURN LABELS(n)", "LABELS", true},
			{"type", "MATCH ()-[r]->() RETURN TYPE(r)", "TYPE", true},
			{"properties", "MATCH (n) RETURN PROPERTIES(n)", "PROPERTIES", true},
			{"nodes", "MATCH p = ()-[]->() RETURN nodes(p)", "nodes", true},
			{"relationships", "MATCH p = ()-[]->() RETURN relationships(p)", "relationships", true},
			{"length_path", "MATCH p = ()-[]->() RETURN length(p)", "length", true},
		})
	})

	t.Run("条件函数", func(t *testing.T) {
		runTestCases(t, CategoryFunctions, "条件函数", []TestCase{
			{"coalesce", "RETURN COALESCE(NULL, 'default')", "COALESCE", true},
			{"nullif", "RETURN NULLIF('value', 'value')", "NULLIF", true},
		})
	})

	t.Run("时间函数", func(t *testing.T) {
		runTestCases(t, CategoryFunctions, "时间函数", []TestCase{
			{"timestamp", "RETURN TIMESTAMP()", "TIMESTAMP", true},
			{"date_func", "RETURN DATE()", "DATE", true},
			{"datetime_func", "RETURN DATETIME()", "DATETIME", true},
		})
	})
}

func TestCompliance_Advanced(t *testing.T) {
	t.Run("列表推导", func(t *testing.T) {
		runTestCases(t, CategoryAdvanced, "列表推导", []TestCase{
			{"list_comprehension", "RETURN [x IN [1,2,3] WHERE x > 1 | x * 2]", "列表推导", true},
			{"pattern_comprehension", "MATCH (p:Person) RETURN [(p)-[:KNOWS]->(f) | f.name] AS friends", "模式推导", true},
		})
	})

	t.Run("索引约束", func(t *testing.T) {
		runTestCases(t, CategoryAdvanced, "索引约束", []TestCase{
			{"create_index", "CREATE INDEX FOR (p:Person) ON (p.name)", "创建索引", false},
			{"create_constraint", "CREATE CONSTRAINT ON (p:Person) ASSERT p.email IS UNIQUE", "创建约束", false},
			{"drop_index", "DROP INDEX FOR (p:Person) ON (p.name)", "删除索引", false},
		})
	})

	t.Run("复杂查询", func(t *testing.T) {
		runTestCases(t, CategoryAdvanced, "复杂查询", []TestCase{
			{"multi_clause", "MATCH (n:Person) WHERE n.age > 18 WITH n ORDER BY n.name SKIP 10 LIMIT 5 RETURN n.name, n.age", "多子句查询", true},
			{"aggregation_grouping", "MATCH (n:Person)-[:WORKS_FOR]->(c:Company) WITH c, COUNT(n) AS employees RETURN c.name, employees", "聚合分组", true},
			{"complex_filter", "MATCH (n:Person) WHERE (n.age > 18 AND n.active = true) OR n.vip = true RETURN n", "复杂过滤", true},
			{"path_analysis", "MATCH path = (a)-[:KNOWS*1..3]->(b) WHERE a.name = 'Alice' RETURN path", "路径分析", true},
		})
	})
}

func TestMain(m *testing.M) {
	complianceReport = ComplianceReport{}
	exitCode := m.Run()
	
	generateReport()
	os.Exit(exitCode)
}

func generateReport() {
	sort.Slice(complianceReport.Results, func(i, j int) bool {
		if complianceReport.Results[i].Category != complianceReport.Results[j].Category {
			return complianceReport.Results[i].Category < complianceReport.Results[j].Category
		}
		return complianceReport.Results[i].Name < complianceReport.Results[j].Name
	})

	for _, r := range complianceReport.Results {
		complianceReport.TotalFeatures++
		switch r.Status {
		case StatusSupported:
			complianceReport.SupportedCount++
		case StatusPartial:
			complianceReport.PartialCount++
		case StatusUnsupported:
			complianceReport.UnsupportedCount++
		case StatusUnknown:
			complianceReport.UnknownCount++
		}
	}

	var sb strings.Builder
	sb.WriteString("# OpenCypher 合规性测试报告\n\n")
	sb.WriteString(fmt.Sprintf("生成时间: %s\n\n", "2026-03-29"))
	sb.WriteString("## 总体统计\n\n")
	sb.WriteString("| 状态 | 数量 | 百分比 |\n")
	sb.WriteString("|------|------|--------|\n")
	total := complianceReport.TotalFeatures
	if total > 0 {
		sb.WriteString(fmt.Sprintf("| ✅ 支持 | %d | %.1f%% |\n", complianceReport.SupportedCount, float64(complianceReport.SupportedCount)/float64(total)*100))
		sb.WriteString(fmt.Sprintf("| ⚠️ 部分支持 | %d | %.1f%% |\n", complianceReport.PartialCount, float64(complianceReport.PartialCount)/float64(total)*100))
		sb.WriteString(fmt.Sprintf("| ❌ 不支持 | %d | %.1f%% |\n", complianceReport.UnsupportedCount, float64(complianceReport.UnsupportedCount)/float64(total)*100))
	}
	sb.WriteString(fmt.Sprintf("| **总计** | **%d** | **100%%** |\n\n", total))

	sb.WriteString("## 详细结果\n\n")

	currentCategory := FeatureCategory("")
	for _, r := range complianceReport.Results {
		if r.Category != currentCategory {
			currentCategory = r.Category
			sb.WriteString(fmt.Sprintf("### %s\n\n", currentCategory))
		}

		statusIcon := "❓"
		switch r.Status {
		case StatusSupported:
			statusIcon = "✅"
		case StatusPartial:
			statusIcon = "⚠️"
		case StatusUnsupported:
			statusIcon = "❌"
		}

		sb.WriteString(fmt.Sprintf("- %s **%s** (%d/%d 通过)", statusIcon, r.Name, r.Passed, r.Passed+r.Failed))
		if r.Notes != "" {
			sb.WriteString(fmt.Sprintf(" - *%s*", r.Notes))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n## 不支持的功能\n\n")
	for _, r := range complianceReport.Results {
		if r.Status == StatusUnsupported {
			sb.WriteString(fmt.Sprintf("- **%s - %s**: %s\n", r.Category, r.Name, r.Notes))
		}
	}

	sb.WriteString("\n## 部分支持的功能\n\n")
	for _, r := range complianceReport.Results {
		if r.Status == StatusPartial {
			sb.WriteString(fmt.Sprintf("- **%s - %s**: %s\n", r.Category, r.Name, r.Notes))
		}
	}

	reportPath := "/Users/ray/workspaces/ai-ecosystem/gograph/spec/cypher_parser/COMPLIANCE_REPORT.md"
	os.WriteFile(reportPath, []byte(sb.String()), 0644)
	fmt.Printf("\n合规性报告已生成: %s\n", reportPath)
}
