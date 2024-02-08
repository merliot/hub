package hub

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func commitMsg() (string, error) {
	// Get the current version
	afterJSON, err := execCmd("cat", fileChildren)
	if err != nil {
		return "", err
	}

	// Get the last version
	beforeJSON, err := execCmd("git", "show", "HEAD:"+fileChildren)
	if err != nil {
		return "", err
	}

	var beforeData, afterData map[string]Child
	if err := json.Unmarshal(beforeJSON, &beforeData); err != nil {
		return "", err
	}
	if err := json.Unmarshal(afterJSON, &afterData); err != nil {
		return "", err
	}

	// Compare before and after
	added, removed := diffEntries(beforeData, afterData)

	// Generate commit message
	return generateCommitMessage(added, removed), nil
}

func execCmd(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	//fmt.Println(cmd.String())
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func diffEntries(before, after map[string]Child) (added, removed map[string]Child) {
	added = make(map[string]Child)
	removed = make(map[string]Child)

	// Find added entries
	for id := range after {
		if _, exists := before[id]; !exists {
			added[id] = after[id]
		}
	}

	// Find removed entries
	for id := range before {
		if _, exists := after[id]; !exists {
			removed[id] = before[id]
		}
	}

	return added, removed
}

func generateCommitMessage(added, removed map[string]Child) string {
	var msgs []string
	var na = len(added)
	var nr = len(removed)

	// TODO handle updates in dirChildren/*

	format := func(id string, child Child) string {
		return "[" + id + ", " + child.Model + ", " + child.Name + "]"
	}

	for id, child := range added {
		msgs = append(msgs, "save: device added: "+format(id, child))
	}

	for id, child := range removed {
		msgs = append(msgs, "save: device deleted: "+format(id, child))
	}

	switch {
	case (na == 1 && nr == 0) || (na == 0 && nr == 1):
		return msgs[0]
	default:
		return fmt.Sprintf("save: devices added %d deleted %d\n\n%s",
			na, nr, strings.Join(msgs, "\n"))
	}

	return "something is wrong"
}

// addChanges git adds new/deleted files
func addChanges() error {

	// Stage new and modified files
	cmd := exec.Command("git", "add", fileChildren)
	//fmt.Println(cmd.String())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to git add %s: %w", fileChildren, err)
	}

	// Stage new and modified files
	cmd = exec.Command("git", "add", dirChildren)
	//fmt.Println(cmd.String())
	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to git add devs/: %w", err)
	}

	// Stage deletions
	cmd = exec.Command("git", "add", "-u", dirChildren)
	//fmt.Println(cmd.String())
	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to git add -u %s: %w", dirChildren, err)
	}

	return nil
}

// hasPendingChanges checks if there are any uncommitted changes in the local repo.
func hasPendingChanges() bool {
	diffCmd := exec.Command("git", "diff", "--cached", "--exit-code")
	_, err := diffCmd.CombinedOutput()
	return err != nil
}

// commitChanges commits any pending changes in the local repo.
func commitChanges(commitMessage, author string) error {
	// Setting environment variables for GIT_AUTHOR and GIT_COMMITTER
	os.Setenv("GIT_AUTHOR_NAME", strings.Split(author, " <")[0])
	os.Setenv("GIT_AUTHOR_EMAIL", strings.Trim(strings.Split(author, "<")[1], "> "))
	os.Setenv("GIT_COMMITTER_NAME", strings.Split(author, " <")[0])
	os.Setenv("GIT_COMMITTER_EMAIL", strings.Trim(strings.Split(author, "<")[1], "> "))

	commitCmd := exec.Command("git", "commit", "-am", commitMessage)
	out, err := commitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to commit changes: %s, %w", out, err)
	}
	fmt.Println("Changes committed successfully!")
	return nil
}

func replaceSpaceWithLF(data []byte) {
	for i := 0; i < len(data); i++ {
		if data[i] == ' ' {
			data[i] = '\n'
		}
	}
}

// pushCommit pushes commits in local repo to remote
func pushCommit(remote, key string) error {
	// 1. Change git remote from HTTPS to SSH
	cmd := exec.Command("git", "remote", "set-url", "origin", remote)
	//fmt.Println(cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push commit: %s, %w", out, err)
	}

	// 2. Write key to temp file
	tempFile, err := ioutil.TempFile("", "git-ssh-key")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(key))
	if err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tempFile.Close()

	// 3. Set permissions on the temp file
	if err = os.Chmod(tempFile.Name(), 0400); err != nil {
		return fmt.Errorf("failed to set permissions on temp file: %w", err)
	}

	// 4. Set GIT_SSH_COMMAND environment variable
	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", tempFile.Name())
	os.Setenv("GIT_SSH_COMMAND", sshCmd)

	// 5. Execute git push command
	cmd = exec.Command("git", "push")
	//fmt.Println(cmd.String())
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push commit: %s, %w", out, err)
	}

	fmt.Println("Pushed commit successfully!")
	return nil
}

func (h *Hub) saveChildren() error {
	if h.gitAuthor == "" {
		return errors.New("Can't save: Missing GIT_AUTHOR env var")
	}
	if h.gitRemote == "" {
		return errors.New("Can't save: Missing GIT_REMOTE env var")
	}
	if h.gitKey == "" {
		return errors.New("Can't save: Missing GIT_KEY env var")
	}
	if err := addChanges(); err != nil {
		return err
	}
	changes := hasPendingChanges()
	if !changes {
		return errors.New("No changes to save")
	}
	commitMsg, err := commitMsg()
	if err != nil {
		return err
	}
	println(commitMsg)
	if err := commitChanges(commitMsg, h.gitAuthor); err != nil {
		return err
	}
	return pushCommit(h.gitRemote, h.gitKey)
}
