package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh"
)

// run ssh in tmux
func sshShellTmux(entry *SSHEntry) error {
	cmd := exec.Command("tmux", "new-window",
		"-n", fmt.Sprintf("%s", entry.displayName),
		fmt.Sprintf("ssh %s", entry.displayName))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// run ssh command
func sshShell2(entry *SSHEntry) error {
	cmd := exec.Command("ssh", entry.displayName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// todo: fix issues when running tui like vim
func sshShell(entry *SSHEntry) {
	privateKey, err := os.ReadFile(entry.keyFile)
	if err != nil {
		log.Printf("failed to read key: %s", err)
		return
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		log.Printf("failed to parse private key: %s", err)
		return
	}

	config := &ssh.ClientConfig{
		User: entry.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%d", entry.serverName, entry.port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Printf("failed to dial: %s", err)
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Printf("failed to create session: %s", err)
		return
	}
	defer session.Close()

	// connect session's stdin/stdout/stderr
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// pty request
	w, h := screen.Size()
	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm"
	}
	if err := session.RequestPty(term, h, w, modes); err != nil {
		log.Printf("failed to request pty: %s", err)
		return
	}

	if err := session.Shell(); err != nil {
		log.Printf("failed to start shell: %s", err)
	}

	if err := session.Wait(); err != nil {
		log.Printf("Session failed: %s", err)
	}
}
