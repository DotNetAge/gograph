package api

import (
	"fmt"
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/graph"
)

func graphStoreTestPath(t *testing.T) string {
	return fmt.Sprintf("/tmp/gograph_graphstore_%s_%d.db", t.Name(), os.Getpid())
}

func TestGraphStore_GetNode_ReturnsNil_Issue(t *testing.T) {
	path := graphStoreTestPath(t)
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	gs := NewGraphStore(db)

	nodeID := "test-node-id"
	userID := "user-12345"

	nodes := []*NodeData{
		{
			ID:     nodeID,
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":    userID,
				"name":  "Test User",
				"email": "test@example.com",
			},
		},
	}

	err = gs.UpsertNodes(nodes)
	if err != nil {
		t.Fatalf("UpsertNodes failed: %v", err)
	}

	retrievedNode, err := gs.GetNode(nodeID)
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}

	if retrievedNode == nil {
		t.Error("ISSUE CONFIRMED: GetNode returned nil instead of the created node")
		return
	}

	t.Logf("GetNode returned node with ID: %s", retrievedNode.ID)

	if retrievedNode.ID != nodeID {
		t.Errorf("expected node ID %s, got %s", nodeID, retrievedNode.ID)
	}

	if len(retrievedNode.Labels) != 1 || retrievedNode.Labels[0] != "User" {
		t.Errorf("expected labels [User], got %v", retrievedNode.Labels)
	}

	idProp, exists := retrievedNode.Properties["id"]
	if !exists {
		t.Error("ISSUE CONFIRMED: 'id' property not found in retrieved node")
		return
	}

	if idProp.StringValue() != userID {
		t.Errorf("expected 'id' property value %s, got %s", userID, idProp.StringValue())
	}

	nameProp, exists := retrievedNode.Properties["name"]
	if !exists {
		t.Error("'name' property not found in retrieved node")
	} else if nameProp.StringValue() != "Test User" {
		t.Errorf("expected 'name' property value 'Test User', got %s", nameProp.StringValue())
	}

	emailProp, exists := retrievedNode.Properties["email"]
	if !exists {
		t.Error("'email' property not found in retrieved node")
	} else if emailProp.StringValue() != "test@example.com" {
		t.Errorf("expected 'email' property value 'test@example.com', got %s", emailProp.StringValue())
	}

	t.Log("GetNode test passed: node retrieved successfully with all properties")
}

func TestGraphStore_GetNode_WithMultipleNodes(t *testing.T) {
	path := graphStoreTestPath(t)
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	gs := NewGraphStore(db)

	nodes := []*NodeData{
		{
			ID:     "node-1",
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":   "user-1",
				"name": "Alice",
			},
		},
		{
			ID:     "node-2",
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":   "user-2",
				"name": "Bob",
			},
		},
		{
			ID:     "node-3",
			Labels: []string{"Product"},
			Properties: map[string]interface{}{
				"id":    "product-1",
				"name":  "Widget",
				"price": 99.99,
			},
		},
	}

	err = gs.UpsertNodes(nodes)
	if err != nil {
		t.Fatalf("UpsertNodes failed: %v", err)
	}

	for _, expectedNode := range nodes {
		retrievedNode, err := gs.GetNode(expectedNode.ID)
		if err != nil {
			t.Errorf("GetNode failed for node %s: %v", expectedNode.ID, err)
			continue
		}

		if retrievedNode == nil {
			t.Errorf("ISSUE CONFIRMED: GetNode returned nil for node %s", expectedNode.ID)
			continue
		}

		if retrievedNode.ID != expectedNode.ID {
			t.Errorf("expected node ID %s, got %s", expectedNode.ID, retrievedNode.ID)
		}

		idProp, exists := retrievedNode.Properties["id"]
		if !exists {
			t.Errorf("'id' property not found for node %s", expectedNode.ID)
		} else {
			expectedID := expectedNode.Properties["id"].(string)
			if idProp.StringValue() != expectedID {
				t.Errorf("node %s: expected 'id' property value %s, got %s", expectedNode.ID, expectedID, idProp.StringValue())
			}
		}
	}

	t.Log("GetNode with multiple nodes test passed")
}

func TestGraphStore_GetNode_NonExistent(t *testing.T) {
	path := graphStoreTestPath(t)
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	gs := NewGraphStore(db)

	_, err = gs.GetNode("non-existent-node")
	if err == nil {
		t.Error("expected error for non-existent node, got nil")
	}
	if err != ErrNodeNotFound {
		t.Logf("got expected error for non-existent node: %v", err)
	}
}

func TestGraphStore_GetNeighbors_ReturnsEmpty_Issue(t *testing.T) {
	path := graphStoreTestPath(t)
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	gs := NewGraphStore(db)

	nodeAID := "node-a"
	nodeBID := "node-b"

	nodes := []*NodeData{
		{
			ID:     nodeAID,
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":   "user-a",
				"name": "User A",
			},
		},
		{
			ID:     nodeBID,
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":   "user-b",
				"name": "User B",
			},
		},
	}

	err = gs.UpsertNodes(nodes)
	if err != nil {
		t.Fatalf("UpsertNodes failed: %v", err)
	}

	edges := []*EdgeData{
		{
			FromNodeID: nodeAID,
			ToNodeID:   nodeBID,
			Type:       "KNOWS",
			Properties: map[string]interface{}{
				"since": 2020,
			},
		},
	}

	err = gs.UpsertEdges(edges)
	if err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	neighbors, err := gs.GetNeighbors(nodeAID, 1, 10)
	if err != nil {
		t.Fatalf("GetNeighbors failed: %v", err)
	}

	if len(neighbors) == 0 {
		t.Error("ISSUE CONFIRMED: GetNeighbors returned empty result instead of neighbor nodes")
		return
	}

	t.Logf("GetNeighbors returned %d neighbors", len(neighbors))

	foundNodeB := false
	for _, neighbor := range neighbors {
		t.Logf("Neighbor: NodeID=%s, EdgeType=%s", neighbor.Node.ID, neighbor.Edge.Type)

		if neighbor.Node.ID == nodeBID {
			foundNodeB = true

			if neighbor.Edge.Type != "KNOWS" {
				t.Errorf("expected edge type 'KNOWS', got '%s'", neighbor.Edge.Type)
			}

			idProp, exists := neighbor.Node.Properties["id"]
			if !exists {
				t.Error("'id' property not found in neighbor node")
			} else if idProp.StringValue() != "user-b" {
				t.Errorf("expected neighbor 'id' property value 'user-b', got %s", idProp.StringValue())
			}
		}
	}

	if !foundNodeB {
		t.Error("ISSUE CONFIRMED: Node B was not found in neighbors list")
	} else {
		t.Log("GetNeighbors test passed: neighbor node B found with correct edge")
	}
}

func TestGraphStore_GetNeighbors_MultipleEdges(t *testing.T) {
	path := graphStoreTestPath(t)
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	gs := NewGraphStore(db)

	nodes := []*NodeData{
		{
			ID:     "center-node",
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":   "center",
				"name": "Center User",
			},
		},
		{
			ID:     "neighbor-1",
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":   "neighbor1",
				"name": "Neighbor 1",
			},
		},
		{
			ID:     "neighbor-2",
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":   "neighbor2",
				"name": "Neighbor 2",
			},
		},
		{
			ID:     "neighbor-3",
			Labels: []string{"Product"},
			Properties: map[string]interface{}{
				"id":    "product1",
				"name":  "Product 1",
				"price": 50.0,
			},
		},
	}

	err = gs.UpsertNodes(nodes)
	if err != nil {
		t.Fatalf("UpsertNodes failed: %v", err)
	}

	edges := []*EdgeData{
		{
			FromNodeID: "center-node",
			ToNodeID:   "neighbor-1",
			Type:       "FRIEND",
			Properties: map[string]interface{}{"since": 2020},
		},
		{
			FromNodeID: "center-node",
			ToNodeID:   "neighbor-2",
			Type:       "FRIEND",
			Properties: map[string]interface{}{"since": 2021},
		},
		{
			FromNodeID: "center-node",
			ToNodeID:   "neighbor-3",
			Type:       "BOUGHT",
			Properties: map[string]interface{}{"date": "2024-01-15"},
		},
	}

	err = gs.UpsertEdges(edges)
	if err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	neighbors, err := gs.GetNeighbors("center-node", 1, 10)
	if err != nil {
		t.Fatalf("GetNeighbors failed: %v", err)
	}

	if len(neighbors) == 0 {
		t.Error("ISSUE CONFIRMED: GetNeighbors returned empty result")
		return
	}

	if len(neighbors) != 3 {
		t.Errorf("expected 3 neighbors, got %d", len(neighbors))
	}

	expectedNeighbors := map[string]string{
		"neighbor-1": "FRIEND",
		"neighbor-2": "FRIEND",
		"neighbor-3": "BOUGHT",
	}

	foundNeighbors := make(map[string]bool)
	for _, neighbor := range neighbors {
		nodeID := neighbor.Node.ID
		edgeType := neighbor.Edge.Type

		expectedType, exists := expectedNeighbors[nodeID]
		if !exists {
			t.Errorf("unexpected neighbor node: %s", nodeID)
			continue
		}

		if edgeType != expectedType {
			t.Errorf("node %s: expected edge type '%s', got '%s'", nodeID, expectedType, edgeType)
		}

		foundNeighbors[nodeID] = true
	}

	for nodeID := range expectedNeighbors {
		if !foundNeighbors[nodeID] {
			t.Errorf("expected neighbor %s not found in results", nodeID)
		}
	}

	t.Log("GetNeighbors with multiple edges test passed")
}

func TestGraphStore_GetNeighbors_Depth2(t *testing.T) {
	path := graphStoreTestPath(t)
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	gs := NewGraphStore(db)

	nodes := []*NodeData{
		{ID: "node-1", Labels: []string{"User"}, Properties: map[string]interface{}{"id": "1", "name": "Node 1"}},
		{ID: "node-2", Labels: []string{"User"}, Properties: map[string]interface{}{"id": "2", "name": "Node 2"}},
		{ID: "node-3", Labels: []string{"User"}, Properties: map[string]interface{}{"id": "3", "name": "Node 3"}},
	}

	err = gs.UpsertNodes(nodes)
	if err != nil {
		t.Fatalf("UpsertNodes failed: %v", err)
	}

	edges := []*EdgeData{
		{FromNodeID: "node-1", ToNodeID: "node-2", Type: "KNOWS"},
		{FromNodeID: "node-2", ToNodeID: "node-3", Type: "KNOWS"},
	}

	err = gs.UpsertEdges(edges)
	if err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	neighbors, err := gs.GetNeighbors("node-1", 2, 10)
	if err != nil {
		t.Fatalf("GetNeighbors failed: %v", err)
	}

	if len(neighbors) == 0 {
		t.Error("ISSUE CONFIRMED: GetNeighbors returned empty result for depth 2")
		return
	}

	t.Logf("GetNeighbors with depth 2 returned %d neighbors", len(neighbors))

	foundNodes := make(map[string]bool)
	for _, neighbor := range neighbors {
		foundNodes[neighbor.Node.ID] = true
	}

	if !foundNodes["node-2"] {
		t.Error("node-2 not found in neighbors (depth 1)")
	}
	if !foundNodes["node-3"] {
		t.Error("node-3 not found in neighbors (depth 2)")
	}

	t.Log("GetNeighbors with depth 2 test passed")
}

func TestGraphStore_GetNeighbors_Limit(t *testing.T) {
	path := graphStoreTestPath(t)
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	gs := NewGraphStore(db)

	nodes := []*NodeData{
		{ID: "center", Labels: []string{"User"}, Properties: map[string]interface{}{"id": "center"}},
		{ID: "n1", Labels: []string{"User"}, Properties: map[string]interface{}{"id": "n1"}},
		{ID: "n2", Labels: []string{"User"}, Properties: map[string]interface{}{"id": "n2"}},
		{ID: "n3", Labels: []string{"User"}, Properties: map[string]interface{}{"id": "n3"}},
		{ID: "n4", Labels: []string{"User"}, Properties: map[string]interface{}{"id": "n4"}},
		{ID: "n5", Labels: []string{"User"}, Properties: map[string]interface{}{"id": "n5"}},
	}

	err = gs.UpsertNodes(nodes)
	if err != nil {
		t.Fatalf("UpsertNodes failed: %v", err)
	}

	edges := []*EdgeData{
		{FromNodeID: "center", ToNodeID: "n1", Type: "KNOWS"},
		{FromNodeID: "center", ToNodeID: "n2", Type: "KNOWS"},
		{FromNodeID: "center", ToNodeID: "n3", Type: "KNOWS"},
		{FromNodeID: "center", ToNodeID: "n4", Type: "KNOWS"},
		{FromNodeID: "center", ToNodeID: "n5", Type: "KNOWS"},
	}

	err = gs.UpsertEdges(edges)
	if err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	neighbors, err := gs.GetNeighbors("center", 1, 3)
	if err != nil {
		t.Fatalf("GetNeighbors failed: %v", err)
	}

	if len(neighbors) == 0 {
		t.Error("ISSUE CONFIRMED: GetNeighbors returned empty result")
		return
	}

	if len(neighbors) != 3 {
		t.Errorf("expected 3 neighbors (limit), got %d", len(neighbors))
	}

	t.Log("GetNeighbors with limit test passed")
}

func TestGraphStore_Integration_FullWorkflow(t *testing.T) {
	path := graphStoreTestPath(t)
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	gs := NewGraphStore(db)

	t.Log("Step 1: Create nodes with UpsertNodes")
	nodes := []*NodeData{
		{
			ID:     "user-1",
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":    "user-1",
				"name":  "Alice",
				"email": "alice@example.com",
				"age":   30,
			},
		},
		{
			ID:     "user-2",
			Labels: []string{"User"},
			Properties: map[string]interface{}{
				"id":    "user-2",
				"name":  "Bob",
				"email": "bob@example.com",
				"age":   25,
			},
		},
	}

	err = gs.UpsertNodes(nodes)
	if err != nil {
		t.Fatalf("UpsertNodes failed: %v", err)
	}
	t.Log("Nodes created successfully")

	t.Log("Step 2: Verify nodes with GetNode")
	for _, expectedNode := range nodes {
		node, err := gs.GetNode(expectedNode.ID)
		if err != nil {
			t.Errorf("GetNode failed for %s: %v", expectedNode.ID, err)
			continue
		}
		if node == nil {
			t.Errorf("ISSUE CONFIRMED: GetNode returned nil for %s", expectedNode.ID)
			continue
		}
		t.Logf("Retrieved node: ID=%s, Labels=%v", node.ID, node.Labels)
	}

	t.Log("Step 3: Create edges with UpsertEdges")
	edges := []*EdgeData{
		{
			FromNodeID: "user-1",
			ToNodeID:   "user-2",
			Type:       "FRIEND",
			Properties: map[string]interface{}{
				"since":     2020,
				"close":     true,
				"frequency": "daily",
			},
		},
	}

	err = gs.UpsertEdges(edges)
	if err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}
	t.Log("Edges created successfully")

	t.Log("Step 4: Verify neighbors with GetNeighbors")
	neighbors, err := gs.GetNeighbors("user-1", 1, 10)
	if err != nil {
		t.Fatalf("GetNeighbors failed: %v", err)
	}

	if len(neighbors) == 0 {
		t.Error("ISSUE CONFIRMED: GetNeighbors returned empty result")
		return
	}

	t.Logf("Found %d neighbors", len(neighbors))

	for _, neighbor := range neighbors {
		t.Logf("Neighbor: NodeID=%s, EdgeType=%s", neighbor.Node.ID, neighbor.Edge.Type)

		idProp := neighbor.Node.Properties["id"]
		nameProp := neighbor.Node.Properties["name"]
		t.Logf("  Node properties: id=%v, name=%v", idProp.StringValue(), nameProp.StringValue())

		sinceProp := neighbor.Edge.Properties["since"]
		closeProp := neighbor.Edge.Properties["close"]
		t.Logf("  Edge properties: since=%v, close=%v", sinceProp.IntValue(), closeProp.BoolValue())
	}

	t.Log("Integration test completed successfully")
}

func TestGraphStore_PropertyTypes(t *testing.T) {
	path := graphStoreTestPath(t)
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	gs := NewGraphStore(db)

	nodes := []*NodeData{
		{
			ID:     "test-node",
			Labels: []string{"Test"},
			Properties: map[string]interface{}{
				"stringProp": "hello",
				"intProp":    42,
				"int64Prop":  int64(999),
				"floatProp":  3.14,
				"boolProp":   true,
			},
		},
	}

	err = gs.UpsertNodes(nodes)
	if err != nil {
		t.Fatalf("UpsertNodes failed: %v", err)
	}

	node, err := gs.GetNode("test-node")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}

	if node == nil {
		t.Fatal("ISSUE CONFIRMED: GetNode returned nil")
	}

	tests := []struct {
		propName     string
		expectedType graph.PropertyType
		checkFunc    func(graph.PropertyValue) bool
	}{
		{
			propName:     "stringProp",
			expectedType: graph.PropertyTypeString,
			checkFunc:    func(p graph.PropertyValue) bool { return p.StringValue() == "hello" },
		},
		{
			propName:     "intProp",
			expectedType: graph.PropertyTypeInt,
			checkFunc:    func(p graph.PropertyValue) bool { return p.IntValue() == 42 },
		},
		{
			propName:     "int64Prop",
			expectedType: graph.PropertyTypeInt,
			checkFunc:    func(p graph.PropertyValue) bool { return p.IntValue() == 999 },
		},
		{
			propName:     "floatProp",
			expectedType: graph.PropertyTypeFloat,
			checkFunc:    func(p graph.PropertyValue) bool { return p.FloatValue() > 3.13 && p.FloatValue() < 3.15 },
		},
		{
			propName:     "boolProp",
			expectedType: graph.PropertyTypeBool,
			checkFunc:    func(p graph.PropertyValue) bool { return p.BoolValue() == true },
		},
	}

	for _, tc := range tests {
		prop, exists := node.Properties[tc.propName]
		if !exists {
			t.Errorf("property %s not found", tc.propName)
			continue
		}

		if prop.Type() != tc.expectedType {
			t.Errorf("property %s: expected type %v, got %v", tc.propName, tc.expectedType, prop.Type())
		}

		if !tc.checkFunc(prop) {
			t.Errorf("property %s: value check failed", tc.propName)
		}
	}

	t.Log("Property types test passed")
}
