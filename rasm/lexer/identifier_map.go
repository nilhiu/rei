package lexer

// IdentToken represents the custom ISA-specific identifier token.
type IdentToken struct {
	tID   TokenID
	subID uint32
}

// NewIdentToken creates a new identifier token for the given [TokenID] with
// the given `subID`.
func NewIdentToken(tokenID TokenID, subID uint32) *IdentToken {
	return &IdentToken{
		tID:   tokenID,
		subID: subID,
	}
}

// IdentifierMap represents an interface that can map an identifier string
// to unique identifier token.
type IdentifierMap interface {
	// Find returns a unique identifier token for the given identifier.
	// If the identifier is unrecognized `false` is returned.
	Find(ident string) (*IdentToken, bool)
}

type simpleIdentifierMap struct {
	idMap map[string]*IdentToken
}

// NewSimpleIdentifierMap returns a new identifier map with the given
// map from string to [IdentToken].
func NewSimpleIdentifierMap(idMap map[string]*IdentToken) IdentifierMap {
	return &simpleIdentifierMap{
		idMap: idMap,
	}
}

func (m *simpleIdentifierMap) Find(ident string) (*IdentToken, bool) {
	i, ok := m.idMap[ident]
	if !ok {
		return nil, false
	}

	return i, true
}
