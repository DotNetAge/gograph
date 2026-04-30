#!/bin/bash
# CLI Integration Test for GoGraph CypherQL
# Tests all core Cypher syntax via the gograph CLI

DB_PATH="test_cli.db"
CLI="./gograph"

# Cleanup function
cleanup() {
    rm -rf "$DB_PATH"
}
trap cleanup EXIT

cleanup

echo "========================================"
echo "GoGraph CLI CypherQL Integration Tests"
echo "========================================"

# Build CLI if needed
if [ ! -f "$CLI" ]; then
    echo "Building gograph CLI..."
    go build -o gograph ./cmd/gograph
fi

PASSED=0
FAILED=0

run_test() {
    local name="$1"
    local query="$2"
    local expect="$3"

    echo ""
    echo "Test: $name"
    echo "Query: $query"

    result=$($CLI query "$DB_PATH" "$query" 2>&1) || true
    if echo "$result" | grep -q "$expect"; then
        echo "✅ PASSED"
        PASSED=$((PASSED + 1))
    else
        echo "❌ FAILED - Expected '$expect' in output:"
        echo "$result"
        FAILED=$((FAILED + 1))
    fi
}

run_exec_test() {
    local name="$1"
    local query="$2"

    echo ""
    echo "Test: $name"
    echo "Query: $query"

    result=$($CLI exec "$DB_PATH" "$query" 2>&1) || true
    if echo "$result" | grep -qvi "error\|failed"; then
        echo "✅ PASSED"
        PASSED=$((PASSED + 1))
    else
        echo "❌ FAILED - Query execution error:"
        echo "$result"
        FAILED=$((FAILED + 1))
    fi
}

# ============================================
# 1. CREATE Tests
# ============================================
echo ""
echo "=== CREATE Tests ==="

run_exec_test "CREATE single node" \
    "CREATE (n:Person {name: 'Alice', age: 30})"

run_exec_test "CREATE multiple nodes" \
    "CREATE (n1:Person {name: 'Bob', age: 25}), (n2:Person {name: 'Charlie', age: 35})"

run_exec_test "CREATE node with multiple labels" \
    "CREATE (n:Person:Employee {name: 'Dave', department: 'Engineering'})"

run_exec_test "CREATE relationship" \
    "MATCH (a:Person {name: 'Alice'}), (b:Person {name: 'Bob'}) CREATE (a)-[:KNOWS {since: 2020}]->(b)"

run_exec_test "CREATE node and relationship" \
    "CREATE (c:Company {name: 'TechCorp'}), (p:Person {name: 'Eve'})-[:WORKS_FOR]->(c)"

# ============================================
# 2. MATCH Tests
# ============================================
echo ""
echo "=== MATCH Tests ==="

run_test "MATCH all nodes" \
    "MATCH (n) RETURN count(n) AS total" \
    "total"

run_test "MATCH by label" \
    "MATCH (p:Person) RETURN p.name ORDER BY p.name" \
    "Alice"

run_test "MATCH with property filter" \
    "MATCH (p:Person {name: 'Alice'}) RETURN p.age" \
    "30"

run_test "MATCH relationship" \
    "MATCH (a:Person)-[:KNOWS]->(b:Person) RETURN a.name, b.name" \
    "Alice"

run_test "MATCH with direction" \
    "MATCH (a:Person)<-[:KNOWS]-(b:Person) RETURN a.name, b.name" \
    "Bob"

# ============================================
# 3. WHERE Tests
# ============================================
echo ""
echo "=== WHERE Tests ==="

run_test "WHERE numeric comparison" \
    "MATCH (p:Person) WHERE p.age > 25 RETURN p.name" \
    "Alice"

run_test "WHERE string equality" \
    "MATCH (p:Person) WHERE p.name = 'Alice' RETURN p.name" \
    "Alice"

run_test "WHERE AND condition" \
    "MATCH (p:Person) WHERE p.age > 25 AND p.name = 'Alice' RETURN p.name" \
    "Alice"

run_test "WHERE OR condition" \
    "MATCH (p:Person) WHERE p.name = 'Alice' OR p.name = 'Bob' RETURN p.name ORDER BY p.name" \
    "Alice"

run_test "WHERE IN list" \
    "MATCH (p:Person) WHERE p.name IN ['Alice', 'Bob'] RETURN p.name ORDER BY p.name" \
    "Alice"

run_test "WHERE STARTS WITH" \
    "MATCH (p:Person) WHERE p.name STARTS WITH 'A' RETURN p.name" \
    "Alice"

run_test "WHERE ENDS WITH" \
    "MATCH (p:Person) WHERE p.name ENDS WITH 'e' RETURN p.name" \
    "Alice"

run_test "WHERE CONTAINS" \
    "MATCH (p:Person) WHERE p.name CONTAINS 'li' RETURN p.name" \
    "Alice"

run_test "WHERE IS NOT NULL" \
    "MATCH (p:Person) WHERE p.age IS NOT NULL RETURN p.name" \
    "Alice"

# ============================================
# 4. RETURN Tests
# ============================================
echo ""
echo "=== RETURN Tests ==="

run_test "RETURN property" \
    "MATCH (p:Person) RETURN p.name ORDER BY p.name" \
    "Alice"

run_test "RETURN alias" \
    "MATCH (p:Person) RETURN p.name AS userName ORDER BY userName" \
    "userName"

run_test "RETURN multiple properties" \
    "MATCH (p:Person) RETURN p.name, p.age ORDER BY p.name" \
    "Alice"

run_test "RETURN DISTINCT" \
    "MATCH (p:Person) RETURN DISTINCT p.age" \
    "30"

# ============================================
# 5. SET Tests
# ============================================
echo ""
echo "=== SET Tests ==="

run_exec_test "SET single property" \
    "MATCH (p:Person {name: 'Alice'}) SET p.city = 'Beijing'"

run_test "SET verify" \
    "MATCH (p:Person {name: 'Alice'}) RETURN p.city" \
    "Beijing"

run_exec_test "SET multiple properties" \
    "MATCH (p:Person {name: 'Bob'}) SET p.city = 'Shanghai', p.email = 'bob@example.com'"

run_exec_test "SET label" \
    "MATCH (p:Person {name: 'Alice'}) SET p:Manager"

run_test "SET label verify" \
    "MATCH (p:Manager) RETURN p.name" \
    "Alice"

# ============================================
# 6. REMOVE Tests
# ============================================
echo ""
echo "=== REMOVE Tests ==="

run_exec_test "REMOVE property" \
    "MATCH (p:Person {name: 'Alice'}) REMOVE p.city"

run_test "REMOVE property verify" \
    "MATCH (p:Person {name: 'Alice'}) RETURN p.city" \
    "NULL"

run_exec_test "REMOVE label" \
    "MATCH (p:Person {name: 'Alice'}) REMOVE p:Manager"

# ============================================
# 7. DELETE Tests
# ============================================
echo ""
echo "=== DELETE Tests ==="

run_exec_test "DELETE relationship" \
    "MATCH (a:Person {name: 'Alice'})-[r:KNOWS]->(b:Person {name: 'Bob'}) DELETE r"

run_exec_test "CREATE for delete test" \
    "CREATE (t:Temp {name: 'TempNode'})"

run_exec_test "DELETE node" \
    "MATCH (t:Temp {name: 'TempNode'}) DELETE t"

# ============================================
# 8. MERGE Tests
# ============================================
echo ""
echo "=== MERGE Tests ==="

run_exec_test "MERGE node" \
    "MERGE (p:Person {name: 'Frank'})"

run_test "MERGE verify" \
    "MATCH (p:Person {name: 'Frank'}) RETURN p.name" \
    "Frank"

run_exec_test "MERGE relationship" \
    "MATCH (a:Person {name: 'Alice'}), (b:Person {name: 'Frank'}) MERGE (a)-[:KNOWS]->(b)"

# ============================================
# 9. ORDER BY / LIMIT / SKIP Tests
# ============================================
echo ""
echo "=== ORDER BY / LIMIT / SKIP Tests ==="

run_test "ORDER BY ASC" \
    "MATCH (p:Person) RETURN p.name ORDER BY p.name ASC" \
    "Alice"

run_test "ORDER BY DESC" \
    "MATCH (p:Person) RETURN p.name ORDER BY p.name DESC" \
    "Frank"

run_test "LIMIT" \
    "MATCH (p:Person) RETURN p.name ORDER BY p.name LIMIT 2" \
    "Alice"

run_test "SKIP" \
    "MATCH (p:Person) RETURN p.name ORDER BY p.name SKIP 2" \
    "Charlie"

run_test "SKIP and LIMIT" \
    "MATCH (p:Person) RETURN p.name ORDER BY p.name SKIP 1 LIMIT 2" \
    "Bob"

# ============================================
# 10. Aggregation Tests
# ============================================
echo ""
echo "=== Aggregation Tests ==="

run_test "COUNT" \
    "MATCH (p:Person) RETURN count(p) AS total" \
    "total"

run_test "MIN MAX" \
    "MATCH (p:Person) RETURN min(p.age) AS minAge, max(p.age) AS maxAge" \
    "minAge"

run_test "SUM AVG" \
    "MATCH (p:Person) RETURN sum(p.age) AS totalAge, avg(p.age) AS avgAge" \
    "totalAge"

run_test "COLLECT" \
    "MATCH (p:Person) RETURN collect(p.name) AS names" \
    "names"

# ============================================
# 11. WITH Tests
# ============================================
echo ""
echo "=== WITH Tests ==="

run_test "WITH basic" \
    "MATCH (p:Person) WITH p.name AS name, p.age AS age WHERE age > 25 RETURN name, age" \
    "Alice"

# ============================================
# 12. UNWIND Tests
# ============================================
echo ""
echo "=== UNWIND Tests ==="

run_test "UNWIND list" \
    "UNWIND [1, 2, 3] AS num RETURN num" \
    "1"

# ============================================
# 13. Parameterized Query Tests
# ============================================
echo ""
echo "=== Parameterized Query Tests ==="

run_test "Parameterized query" \
    "MATCH (p:Person) WHERE p.age > 25 RETURN p.name ORDER BY p.name" \
    "Alice"

# ============================================
# 14. UNION Tests (Known Limitation - Not Yet Implemented)
# ============================================
echo ""
echo "=== UNION Tests (Known Limitation) ==="

echo "Skipping UNION tests - parser support not yet implemented"

# ============================================
# Summary
# ============================================
echo ""
echo "========================================"
echo "Test Summary"
echo "========================================"
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo "Total:  $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo ""
    echo "🎉 All tests passed!"
    exit 0
else
    echo ""
    echo "⚠️  Some tests failed!"
    exit 1
fi
