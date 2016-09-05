package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	docker "github.com/fsouza/go-dockerclient"
)

func Fatal(m string) {
	fmt.Printf(m)
	os.Exit(0)
}

func Cleanup(client *docker.Client, ID string) {
	err := client.KillContainer(docker.KillContainerOptions{
		ID: ID,
	})

	if err != nil {
		fmt.Println("Error in kill container:", err.Error())
	}

	err = client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            ID,
		RemoveVolumes: true,
		Force:         true,
	})

	if err != nil {
		fmt.Println("Error in remove container:", err.Error())
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <docker-image>\n", os.Args[0])
		os.Exit(0)
	}
	image := os.Args[1]
	var cmd []string = os.Args[2:]

	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	client, err := docker.NewClient("unix://var/run/docker.sock")
	if err != nil {
		Fatal(err.Error())
	}

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        image,
			Cmd:          cmd,
			Tty:          true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
		},
	})

	if err != nil {
		Fatal(err.Error())
	}

	err = client.StartContainer(container.ID, nil)
	if err != nil {
		Fatal(err.Error())
	}

	// start the signal monitor
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT)

	go func() {
		<-sig
		client.KillContainer(docker.KillContainerOptions{
			ID: container.ID,
		})
	}()

	err = client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    container.ID,
		InputStream:  os.Stdin,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		RawTerminal:  true,
		// Success:      notify,
	})

	client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
		Force:         true,
	})
}
