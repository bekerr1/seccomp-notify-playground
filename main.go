package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"syscall"

	libseccomp "github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

const socketPath = "/run/seccomp-agent.socket"

func main() {
	//// Kubernetes client setup
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	log.Fatalf("Cluster config error: %v", err)
	//}
	//clientset, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	log.Fatalf("Clientset error: %v", err)
	//}

	// Unix socket setup
	if err := os.RemoveAll(socketPath); err != nil {
		log.Printf("Warning: Socket removal failed: %v", err)
	}
	listener, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		log.Fatalf("Socket creation error: %v", err)
	}
	addr := &unix.SockaddrUnix{Name: socketPath}
	if err := unix.Bind(listener, addr); err != nil {
		log.Fatalf("Socket bind error: %v", err)
	}
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Fatalf("Socket permission error: %v", err)
	}
	if err := unix.Listen(listener, 5); err != nil {
		log.Fatalf("Socket listen error: %v", err)
	}
	defer unix.Close(listener)

	log.Printf("Listening on %s : new", socketPath)

	for {
		log.Printf("Waiting for connection...")
		fd, _, err := unix.Accept(listener)
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go socketFDHandler(fd)
	}
}

func socketFDHandler(fd int) {
	log.Printf("Accepted connection: %d", fd)
	defer unix.Close(fd)

	notifyFd, stateData, err := recieveSCMRights(fd)
	if err != nil {
		log.Printf("Error receiving SCM rights: %v", err)
		return
	}
	//unix.Close(fd) // Close the original socket fd
	defer func() {
		log.Printf("Closing notifyFd: %d", notifyFd)
		syscall.Close(int(notifyFd)) // Close notifyFd
	}()

	log.Printf("Received notifyFd: %d", notifyFd)
	log.Printf("Received stateData: %v", string(stateData))

	for {
		req, err := libseccomp.NotifReceive(notifyFd)
		if err != nil {
			if err == syscall.ENOENT {
				log.Printf("NotifReceive: syscall canceled")
				return
			}
			log.Printf("NotifReceive error: %v", err)
			return
		}
		if err := libseccomp.NotifIDValid(notifyFd, req.ID); err != nil {
			log.Printf("Invalid notification ID: %v", err)
			continue
		}

		resp := handleSyscall(req)

		if err := libseccomp.NotifRespond(notifyFd, &resp); err != nil {
			log.Printf("NotifRespond error: %v", err)
			return
		}
	}
}

func recieveSCMRights(fd int) (libseccomp.ScmpFd, []byte, error) {
	// Receive SCM_RIGHTS message containing notifyFd and stateData
	oob := make([]byte, unix.CmsgSpace(4))
	stateData := make([]byte, math.MaxInt16)
	n, oobn, _, _, err := unix.Recvmsg(fd, stateData, oob, 0)
	if err != nil {
		log.Printf("Recvmsg error: %v", err)
		return 0, nil, err
	}
	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return 0, nil, fmt.Errorf("ParseSocketControlMessage error: %v", err)
	}
	if len(msgs) == 0 {
		return 0, nil, fmt.Errorf("no SCM_RIGHTS message")
	}
	fds, err := syscall.ParseUnixRights(&msgs[0])
	if err != nil {
		return 0, nil, fmt.Errorf("ParseUnixRights error: %v", err)
	}
	if len(fds) == 0 {
		return 0, nil, fmt.Errorf("no FDs received")
	}
	return libseccomp.ScmpFd(fds[0]), stateData[:n], nil
}

func handleSyscall(req *libseccomp.ScmpNotifReq) libseccomp.ScmpNotifResp {
	syscallName, err := req.Data.Syscall.GetName()
	if err != nil {
		log.Printf("error getting syscall name from req data '%v': %v", req.Data.Syscall, err)
		return libseccomp.ScmpNotifResp{
			ID:    req.ID,
			Error: int32(syscall.ENOTNAM),
		}
	}
	log.Printf("Received syscall: %s [id: %v, args: %v]\n", syscallName, req.ID, req.Data.Args)
	return libseccomp.ScmpNotifResp{
		ID:    req.ID,
		Error: 0,
		Val:   0,
		Flags: libseccomp.NotifRespFlagContinue,
	}
}

//func readStringArg(req *libseccomp.SeccompNotif, argNum uint) (string, error) {
//	if argNum >= uint(len(req.Data.Args)) {
//		return "", fmt.Errorf("invalid argument number: %d", argNum)
//	}
//	buf := make([]byte, 256)
//	n, err := unix.Read(int(req.Data.Args[argNum]), buf)
//	if err != nil {
//		return "", err)
//	}
//	return strings.TrimRight(string(buf[:n]), "\x00"), nil
//}
