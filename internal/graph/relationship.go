package graph

type Relationship struct {
	ID         string
	StartNodeID string
	EndNodeID   string
	Type       string
	Properties map[string]PropertyValue
}

func NewRelationship(startNodeID, endNodeID, relType string, properties map[string]interface{}) *Relationship {
	props := make(map[string]PropertyValue)
	for k, v := range properties {
		props[k] = ToPropertyValue(v)
	}
	return &Relationship{
		ID:          GenerateID("rel"),
		StartNodeID: startNodeID,
		EndNodeID:   endNodeID,
		Type:        relType,
		Properties:  props,
	}
}

func (r *Relationship) GetProperty(key string) (PropertyValue, bool) {
	v, ok := r.Properties[key]
	return v, ok
}

func (r *Relationship) SetProperty(key string, value PropertyValue) {
	r.Properties[key] = value
}

func (r *Relationship) RemoveProperty(key string) {
	delete(r.Properties, key)
}
