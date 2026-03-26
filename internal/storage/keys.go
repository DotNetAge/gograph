package storage

const (
	KeyPrefixNode        = "node:"
	KeyPrefixRel         = "rel:"
	KeyPrefixLabel       = "label:"
	KeyPrefixProp        = "prop:"
	KeyPrefixAdj         = "adj:"
	KeyPrefixMeta        = "meta:"
	KeyPrefixIndexCount  = "meta:index_count"
)

func NodeKey(nodeID string) []byte {
	return []byte(KeyPrefixNode + nodeID)
}

func RelKey(relID string) []byte {
	return []byte(KeyPrefixRel + relID)
}

func LabelKey(labelName, nodeID string) []byte {
	return []byte(KeyPrefixLabel + labelName + ":" + nodeID)
}

func LabelKeyPrefix(labelName string) []byte {
	return []byte(KeyPrefixLabel + labelName + ":")
}

func PropertyKey(labelName, propName, propValue string) []byte {
	return []byte(KeyPrefixProp + labelName + ":" + propName + ":" + propValue)
}

func PropertyKeyPrefix(labelName, propName string) []byte {
	return []byte(KeyPrefixProp + labelName + ":" + propName + ":")
}

func AdjKey(nodeID, relType string, direction string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":" + relType + ":" + direction)
}

func AdjKeyPrefix(nodeID string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":")
}

func AdjKeyPrefixNodeAndType(nodeID, relType string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":" + relType + ":")
}
