package graph

// GitChecker determines whether a file has changed since a given commit.
// Implementations wrap the actual git CLI (Humble Object pattern).
type GitChecker interface {
	// HasFileChangedSince returns true if the file at filePath has been
	// modified in any commit after the given commit hash.
	HasFileChangedSince(commit, filePath string) (bool, error)
}
