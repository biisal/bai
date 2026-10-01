package git

type Noop struct{}

func (Noop) Init() error                          { return nil }
func (Noop) CheckIfDirty() (bool, string, error)  { return false, "", nil }
func (Noop) Add(...string) error                  { return nil }
func (Noop) Commit(string) error                  { return nil }
func (Noop) CheckIfGitInitialized() (bool, error) { return true, nil }
func (Noop) InsertToGitIgnore(...string) error    { return nil }
func (Noop) Dir() string                          { return "" }
