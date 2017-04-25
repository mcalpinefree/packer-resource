package docker

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func RunCmd(name string, arg ...string) int {
	// http://stackoverflow.com/questions/10385551/get-exit-code-go
	cmd := exec.Command(name, arg...)
	cmd.Stdout = cmd.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start: %v")
	}
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
	}
	return 0
}

func fstab_contains_cgroup() bool {
	file, err := os.Open("/etc/fstab")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.HasPrefix(strings.TrimSpace(scanner.Text()), "#") {
			continue
		}
		if strings.Contains(scanner.Text(), "cgroup") {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return false
}

func create_cgroups() {
	previous_dir, _ := os.Getwd()
	os.Chdir("/sys/fs/cgroup")
	file, err := os.Open("/sys/fs/cgroup")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if strings.HasPrefix(fields[0], "#") {
			continue
		}

		if len(fields) > 3 && fields[3] == "1" {
			sys := fields[0]
			os.Mkdir(sys, 0775)
			if RunCmd("mountpoint", "-q", sys) != 0 {
				if RunCmd("mount", "-n", "-t", "cgroup", "-o", sys, "cgroup", sys) != 0 {
					RunCmd("rmdir", sys)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	os.Chdir(previous_dir)
}

// https://gist.github.com/tangfei67/5fdca2ef0ec30b486fe0
// https://github.com/docker/docker/issues/8791
// see also https://github.com/tianon/cgroupfs-mount/blob/master/cgroupfs-mount
func CgroupfsMount() {
	if fstab_contains_cgroup() {
		return
	}

	if _, err := os.Stat("/proc/cgroups"); os.IsNotExist(err) {
		return
	}

	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		return
	}

	// mount /sys/fs/cgroup
	if RunCmd("mountpoint", "-q", "/sys/fs/cgroup") != 0 {
		RunCmd("mount", "-t", "tmpfs", "-o", "uid=0,gid=0,mode=0755", "cgroup", "/sys/fs/cgroup")
	}

	create_cgroups()

}

func StartDocker() *exec.Cmd {
	cmd := exec.Command("dockerd",
	"--host=unix:///var/run/docker.sock",
	"--host=tcp://0.0.0.0:2375",
	"--storage-driver=vfs")
	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start: %v")
	}

	return cmd
}
