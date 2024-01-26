package hub

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// addChanges git adds new/deleted files
func addChanges() error {
	// Stage new and modified files
	cmd := exec.Command("git", "add", dirChildren)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to git add devs/: %w", err)
	}
	// Stage deletions
	cmd = exec.Command("git", "add", "-u", dirChildren)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to git add -u %s: %w", dirChildren, err)
	}
	return nil
}

// hasPendingChanges checks if there are any uncommitted changes in the local repo.
func hasPendingChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
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

	// (key got messed up being stuffed into env var, so un-mess it)
	keyBytes := []byte(key)
	replaceSpaceWithLF(keyBytes[35 : len(keyBytes)-33])
	keyBytes = append(keyBytes, '\n')

	_, err = tempFile.Write(keyBytes)
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
	changes, err := hasPendingChanges()
	if err != nil {
		return err
	}
	if !changes {
		return errors.New("No changes to save")
	}
	if err := commitChanges("update devices", h.gitAuthor); err != nil {
		return err
	}
	return pushCommit(h.gitRemote, h.gitKey)
}
