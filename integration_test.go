package gograph

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/api"
)

func TestFullCRUDOperations(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("CreateNodes", func(t *testing.T) {
		queries := []string{
			`CREATE (a:User {name:"Alice", city:"Beijing", age:30})`,
			`CREATE (b:User {name:"Bob", city:"Shanghai", age:35})`,
			`CREATE (c:User {name:"Charlie", city:"Beijing", age:40})`,
			`CREATE (d:User {name:"David", city:"Guangzhou", age:25})`,
			`CREATE (e:User {name:"Eve", city:"Shenzhen", age:28})`,
			`CREATE (p1:Product {name:"iPhone", price:999.99, brand:"Apple"})`,
			`CREATE (p2:Product {name:"MacBook", price:1999.99, brand:"Apple"})`,
			`CREATE (p3:Product {name:"Galaxy", price:899.99, brand:"Samsung"})`,
			`CREATE (o1:Order {id:"ORD001", total:1099.99})`,
			`CREATE (o2:Order {id:"ORD002", total:2899.98})`,
		}

		for _, q := range queries {
			_, err := db.Exec(ctx, q)
			if err != nil {
				t.Errorf("failed to execute %s: %v", q, err)
			}
		}
	})

	t.Run("ReadWithMatch", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User {name:"Alice"}) RETURN u.name, u.city, u.age`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("expected to find Alice")
		}
	})

	t.Run("UpdateWithSet", func(t *testing.T) {
		_, err := db.Exec(ctx, `MATCH (u:User {name:"Alice"}) SET u.email = "alice@example.com"`)
		if err != nil {
			t.Fatalf("failed to set email: %v", err)
		}

		rows, err := db.Query(ctx, `MATCH (u:User {name:"Alice"}) RETURN u.email`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var email string
			if err := rows.Scan(&email); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if email != "alice@example.com" {
				t.Errorf("expected email 'alice@example.com', got '%s'", email)
			}
		} else {
			t.Error("Alice not found after SET")
		}
	})

	t.Run("DeleteNode", func(t *testing.T) {
		_, err := db.Exec(ctx, `MATCH (u:User {name:"Eve"}) DELETE u`)
		if err != nil {
			t.Fatalf("failed to delete: %v", err)
		}

		rows, err := db.Query(ctx, `MATCH (u:User {name:"Eve"}) RETURN u.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			t.Error("Eve should have been deleted")
		}
	})
}

func TestRelationshipOperations(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	_, _ = db.Exec(ctx, `CREATE (a:User {name:"Alice"})-[:KNOWS]->(b:User {name:"Bob"})`)
	_, _ = db.Exec(ctx, `CREATE (b:User {name:"Bob"})-[:KNOWS]->(c:User {name:"Charlie"})`)
	_, _ = db.Exec(ctx, `CREATE (a:User {name:"Alice"})-[:FOLLOWS]->(c:User {name:"Charlie"})`)

	t.Run("SimpleRelationship", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (a:User)-[r:KNOWS]->(b:User) RETURN a.name, type(r), b.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 2 {
			t.Errorf("expected 2 KNOWS relationships, got %d", count)
		}
	})

	t.Run("RelationshipTraversal", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (a:User)-[*]->(b:User) RETURN a.name, b.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count == 0 {
			t.Error("expected at least one path traversal")
		}
	})

	t.Run("RelationshipWithProperties", func(t *testing.T) {
		_, _ = db.Exec(ctx, `MATCH (a:User {name:"Alice"}), (b:User {name:"Bob"}) CREATE (a)-[r:KNOWS {since:2020}]->(b)`)

		rows, err := db.Query(ctx, `MATCH (a:User)-[r:KNOWS]->(b:User) WHERE r.since > 2019 RETURN a.name, r.since, b.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count == 0 {
			t.Error("expected at least one relationship with since > 2019")
		}
	})
}

func TestWhereConditionsAND(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	_, _ = db.Exec(ctx, `CREATE (a:User {name:"Alice", city:"Beijing", age:30})`)
	_, _ = db.Exec(ctx, `CREATE (b:User {name:"Bob", city:"Shanghai", age:35})`)
	_, _ = db.Exec(ctx, `CREATE (c:User {name:"Charlie", city:"Beijing", age:40})`)
	_, _ = db.Exec(ctx, `CREATE (d:User {name:"David", city:"Guangzhou", age:45})`)

	t.Run("ANDCondition", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) WHERE u.city = "Beijing" AND u.age > 30 RETURN u.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 1 {
			t.Errorf("expected 1 user (Charlie), got %d", count)
		}
	})

	t.Run("ORCondition", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) WHERE u.city = "Beijing" OR u.age > 42 RETURN u.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count < 1 {
			t.Errorf("expected at least 1 user, got %d", count)
		}
	})

	t.Run("INCondition", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) WHERE u.city IN ["Beijing", "Shanghai"] RETURN u.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 3 {
			t.Errorf("expected 3 users, got %d", count)
		}
	})

	t.Run("ComparisonOperators", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) WHERE u.age >= 35 RETURN u.name, u.age`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 3 {
			t.Errorf("expected 3 users with age >= 35, got %d", count)
		}
	})
}

func TestStringFunctions(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	_, _ = db.Exec(ctx, `CREATE (a:User {name:"Alice", city:"Beijing"})`)
	_, _ = db.Exec(ctx, `CREATE (b:User {name:"Bob", city:"Shanghai"})`)
	_, _ = db.Exec(ctx, `CREATE (c:User {name:"Charlie", city:"NewYork"})`)
	_, _ = db.Exec(ctx, `CREATE (d:User {name:"David", city:"Guangzhou"})`)

	t.Run("STARTSWITH", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) WHERE u.city STARTS WITH "Bei" RETURN u.name, u.city`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("expected to find Alice (Beijing)")
		}
	})

	t.Run("ENDSWITH", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) WHERE u.city ENDS WITH "hou" RETURN u.name, u.city`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("expected to find Bob (Shanghai)")
		}
	})

	t.Run("CONTAINS", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) WHERE u.city CONTAINS "an" RETURN u.name, u.city`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 2 {
			t.Errorf("expected 2 users (Bob:Shanghai, David:Guangzhou), got %d", count)
		}
	})
}

func TestMergeOperations(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("MergeOnCreateSet", func(t *testing.T) {
		_, err := db.Exec(ctx, `MERGE (u:User {name:"Alice"}) ON CREATE SET u.email = "alice@new.com"`)
		if err != nil {
			t.Fatalf("failed to merge: %v", err)
		}

		rows, err := db.Query(ctx, `MATCH (u:User {name:"Alice"}) RETURN u.email`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var email string
			if err := rows.Scan(&email); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if email != "alice@new.com" {
				t.Errorf("expected email 'alice@new.com', got '%s'", email)
			}
		} else {
			t.Error("Alice not found")
		}
	})

	t.Run("MergeOnMatchSet", func(t *testing.T) {
		_, err := db.Exec(ctx, `MERGE (u:User {name:"Alice"}) ON MATCH SET u.lastSeen = "2024-01-01"`)
		if err != nil {
			t.Fatalf("failed to merge: %v", err)
		}

		rows, err := db.Query(ctx, `MATCH (u:User {name:"Alice"}) RETURN u.lastSeen`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var lastSeen string
			if err := rows.Scan(&lastSeen); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if lastSeen != "2024-01-01" {
				t.Errorf("expected lastSeen '2024-01-01', got '%s'", lastSeen)
			}
		} else {
			t.Error("Alice not found")
		}
	})

	t.Run("MergeCreatesIfNotExists", func(t *testing.T) {
		_, err := db.Exec(ctx, `MERGE (u:User {name:"NewUser"})`)
		if err != nil {
			t.Fatalf("failed to merge: %v", err)
		}

		rows, err := db.Query(ctx, `MATCH (u:User {name:"NewUser"}) RETURN u.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("NewUser should have been created")
		}
	})

	t.Run("MergeDoesNotDuplicate", func(t *testing.T) {
		initialCount := 0
		rows, err := db.Query(ctx, `MATCH (u:User {name:"Alice"}) RETURN count(u) as cnt`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		if rows.Next() {
			if err := rows.Scan(&initialCount); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
		}
		rows.Close()

		_, err = db.Exec(ctx, `MERGE (u:User {name:"Alice"})`)
		if err != nil {
			t.Fatalf("failed to merge: %v", err)
		}

		rows, err = db.Query(ctx, `MATCH (u:User {name:"Alice"}) RETURN count(u) as cnt`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var finalCount int
			if err := rows.Scan(&finalCount); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if finalCount != initialCount {
				t.Errorf("expected count to remain %d after MERGE, got %d", initialCount, finalCount)
			}
		}
	})
}

func TestReturnClause(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	_, _ = db.Exec(ctx, `CREATE (a:User {name:"Alice", score:100})`)
	_, _ = db.Exec(ctx, `CREATE (b:User {name:"Bob", score:85})`)

	t.Run("ReturnProperty", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) RETURN u.name, u.score ORDER BY u.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 2 {
			t.Errorf("expected 2 rows, got %d", count)
		}
	})

	t.Run("ReturnWithAliases", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) RETURN u.name AS userName, u.score AS userScore`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 2 {
			t.Errorf("expected 2 rows, got %d", count)
		}
	})
}

func TestOptionalMatch(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	_, _ = db.Exec(ctx, `CREATE (a:User {name:"Alice"})`)

	t.Run("OptionalMatchWithResults", func(t *testing.T) {
		rows, err := db.Query(ctx, `OPTIONAL MATCH (u:User {name:"Alice"})-[:KNOWS]->(f:User) RETURN u.name, f.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var uName, fName string
			if err := rows.Scan(&uName, &fName); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if uName != "Alice" {
				t.Errorf("expected uName 'Alice', got '%s'", uName)
			}
		}
	})

	t.Run("OptionalMatchNoResults", func(t *testing.T) {
		rows, err := db.Query(ctx, `OPTIONAL MATCH (u:User {name:"NonExistent"})-[:KNOWS]->(f:User) RETURN u.name, f.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var uName, fName string
			if err := rows.Scan(&uName, &fName); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if uName != "NonExistent" {
				t.Errorf("expected uName 'NonExistent', got '%s'", uName)
			}
		}
	})
}

func TestBatchData(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("Create25Products", func(t *testing.T) {
		for i := 1; i <= 25; i++ {
			_, err := db.Exec(ctx, fmt.Sprintf(`CREATE (p:Product {id:%d, name:"Product%d", price:%.2f})`, i, i, float64(i)*10.0))
			if err != nil {
				t.Fatalf("failed to create product %d: %v", i, err)
			}
		}

		rows, err := db.Query(ctx, `MATCH (p:Product {name:"Product25"}) RETURN p.id, p.name, p.price`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("expected to find Product25")
		}
	})

	t.Run("BatchQueryWithWhere", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (p:Product) WHERE p.price > 100 RETURN p.id, p.name, p.price ORDER BY p.price DESC`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count < 10 {
			t.Errorf("expected more than 10 products with price > 100, got %d", count)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("EmptyResult", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:NonExistent) RETURN u.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			t.Error("expected no results for non-existent label")
		}
	})

	t.Run("NodeWithoutProperty", func(t *testing.T) {
		_, err := db.Exec(ctx, `CREATE (u:User {name:"Test"})`)
		if err != nil {
			t.Fatalf("failed to create: %v", err)
		}

		rows, err := db.Query(ctx, `MATCH (u:User {name:"Test"}) RETURN u.email`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var email string
			if err := rows.Scan(&email); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if email != "" {
				t.Error("expected empty email for new user without email property")
			}
		}
	})

	t.Run("MultipleLabels", func(t *testing.T) {
		_, err := db.Exec(ctx, `CREATE (p:Product:Item {name:"TestProduct"})`)
		if err != nil {
			t.Fatalf("failed to create: %v", err)
		}

		rows, err := db.Query(ctx, `MATCH (p:Product) WHERE p.name = "TestProduct" RETURN p.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("expected to find product with label Product")
		}
	})

	t.Run("NumericPropertyComparisons", func(t *testing.T) {
		_, _ = db.Exec(ctx, `CREATE (a:Item {value:10})`)
		_, _ = db.Exec(ctx, `CREATE (b:Item {value:20})`)
		_, _ = db.Exec(ctx, `CREATE (c:Item {value:30})`)

		rows, err := db.Query(ctx, `MATCH (i:Item) WHERE i.value >= 20 RETURN i.value`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 2 {
			t.Errorf("expected 2 items with value >= 20, got %d", count)
		}
	})
}

func TestOrderBy(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	for i := 1; i <= 10; i++ {
		_, _ = db.Exec(ctx, fmt.Sprintf(`CREATE (u:User {name:"User%d", age:%d, score:%d})`, i, 20+i%5, 100-i*5))
	}

	t.Run("OrderByAscending", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) RETURN u.name, u.age ORDER BY u.age ASC`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		var ages []int
		for rows.Next() {
			var name string
			var age int
			if err := rows.Scan(&name, &age); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			ages = append(ages, age)
		}

		for i := 1; i < len(ages); i++ {
			if ages[i] < ages[i-1] {
				t.Errorf("expected ages in ascending order, got %v", ages)
				break
			}
		}
	})

	t.Run("OrderByDescending", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) RETURN u.name, u.score ORDER BY u.score DESC`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		var scores []int
		for rows.Next() {
			var name string
			var score int
			if err := rows.Scan(&name, &score); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			scores = append(scores, score)
		}

		for i := 1; i < len(scores); i++ {
			if scores[i] > scores[i-1] {
				t.Errorf("expected scores in descending order, got %v", scores)
				break
			}
		}
	})

	t.Run("OrderByMultipleColumns", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) RETURN u.age, u.score ORDER BY u.age ASC, u.score DESC`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 10 {
			t.Errorf("expected 10 rows, got %d", count)
		}
	})
}

func TestLimitSkip(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	for i := 1; i <= 20; i++ {
		_, _ = db.Exec(ctx, fmt.Sprintf(`CREATE (n:Num {value:%d})`, i))
	}

	t.Run("Limit3", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (n:Num) RETURN n.value ORDER BY n.value LIMIT 3`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 3 {
			t.Errorf("expected 3 rows, got %d", count)
		}
	})

	t.Run("Skip5", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (n:Num) RETURN n.value ORDER BY n.value SKIP 5`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 15 {
			t.Errorf("expected 15 rows after SKIP 5, got %d", count)
		}
	})

	t.Run("Skip5Limit3", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (n:Num) RETURN n.value ORDER BY n.value SKIP 5 LIMIT 3`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		var values []int
		for rows.Next() {
			var v int
			if err := rows.Scan(&v); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			values = append(values, v)
		}

		if len(values) != 3 {
			t.Errorf("expected 3 rows, got %d", len(values))
		}
		if len(values) >= 1 && values[0] != 6 {
			t.Errorf("expected first value to be 6 (after SKIP 5), got %d", values[0])
		}
	})
}

func TestUnwind(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("UnwindList", func(t *testing.T) {
		rows, err := db.Query(ctx, `UNWIND [1, 2, 3, 4, 5] AS n RETURN n`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var n int
			if err := rows.Scan(&n); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			count++
		}
		if count != 5 {
			t.Errorf("expected 5 rows from UNWIND, got %d", count)
		}
	})

	t.Run("UnwindWithMatch", func(t *testing.T) {
		_, _ = db.Exec(ctx, `CREATE (u:User {name:"Alice"})`)
		rows, err := db.Query(ctx, `UNWIND ["Alice", "Bob"] AS name MATCH (u:User {name:name}) RETURN u.name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 1 {
			t.Errorf("expected 1 match, got %d", count)
		}
	})
}

func TestAggregateFunctions(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	for i := 1; i <= 10; i++ {
		_, _ = db.Exec(ctx, fmt.Sprintf(`CREATE (o:Order {id:%d, amount:%d.00})`, i, i*10))
	}

	t.Run("CountAll", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (o:Order) RETURN count(o)`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if count != 10 {
				t.Errorf("expected count 10, got %d", count)
			}
		}
	})

	t.Run("SumAmount", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (o:Order) RETURN sum(o.amount)`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var total float64
			if err := rows.Scan(&total); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			expected := 550.0
			if total != expected {
				t.Errorf("expected sum 550.0, got %.2f", total)
			}
		}
	})

	t.Run("AvgAmount", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (o:Order) RETURN avg(o.amount)`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var avg float64
			if err := rows.Scan(&avg); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			expected := 55.0
			if avg != expected {
				t.Errorf("expected avg 55.0, got %.2f", avg)
			}
		}
	})

	t.Run("MinMaxAmount", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (o:Order) RETURN min(o.amount), max(o.amount)`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var minAmt, maxAmt float64
			if err := rows.Scan(&minAmt, &maxAmt); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if minAmt != 10.0 {
				t.Errorf("expected min 10.0, got %.2f", minAmt)
			}
			if maxAmt != 100.0 {
				t.Errorf("expected max 100.0, got %.2f", maxAmt)
			}
		}
	})

	t.Run("CountWithGroupBy", func(t *testing.T) {
		_, _ = db.Exec(ctx, `CREATE (a:Category {name:"Electronics"})`)
		_, _ = db.Exec(ctx, `CREATE (b:Category {name:"Electronics"})`)
		_, _ = db.Exec(ctx, `CREATE (c:Category {name:"Clothing"})`)

		rows, err := db.Query(ctx, `MATCH (c:Category) RETURN c.name, count(*)`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		counts := make(map[string]int)
		for rows.Next() {
			var name string
			var cnt int
			if err := rows.Scan(&name, &cnt); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			counts[name] = cnt
		}

		if counts["Electronics"] != 2 {
			t.Errorf("expected Electronics count 2, got %d", counts["Electronics"])
		}
		if counts["Clothing"] != 1 {
			t.Errorf("expected Clothing count 1, got %d", counts["Clothing"])
		}
	})
}

func TestWithClause(t *testing.T) {
	path := testPath(t)
	defer os.Remove(path)

	db, err := api.Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	for i := 1; i <= 10; i++ {
		_, _ = db.Exec(ctx, fmt.Sprintf(`CREATE (u:User {name:"User" + "%d", age:%d})`, i, 15+i%10))
	}

	t.Run("WithOrderByLimit", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) WITH u ORDER BY u.age DESC RETURN u.name, u.age SKIP 2 LIMIT 3`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}
		if count != 3 {
			t.Errorf("expected 3 rows, got %d", count)
		}
	})

	t.Run("WithAggregation", func(t *testing.T) {
		rows, err := db.Query(ctx, `MATCH (u:User) WITH count(u) AS total RETURN total`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			var total int
			if err := rows.Scan(&total); err != nil {
				t.Fatalf("failed to scan: %v", err)
			}
			if total != 10 {
				t.Errorf("expected total 10, got %d", total)
			}
		}
	})
}
