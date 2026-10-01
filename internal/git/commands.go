package git

func (c *Git) Init() error {
	_, err := c.runGitCommand("init")
	return err
}

func (c *Git) Add(paths ...string) error {
	args := append([]string{"add", "--"}, paths...)
	_, err := c.runGitCommand(args...)
	return err
}

func (c *Git) Commit(message string) error {
	args := []string{"commit", "-m", message}
	_, err := c.runGitCommand(args...)
	return err
}
