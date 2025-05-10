package main

import (
	"log"
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
		go handle(fd)
	}
}

func handle(fd int) {
	log.Printf("Accepted connection: %d", fd)
	defer unix.Close(fd)
	// Receive SCM_RIGHTS message containing notifyFd
	oob := make([]byte, unix.CmsgSpace(4))
	n, oobn, _, _, err := unix.Recvmsg(fd, nil, oob, 0)
	if err != nil {
		log.Printf("Recvmsg error: %v", err)
		return
	}
	if n > 0 || oobn == 0 {
		log.Printf("No control message received")
		return
	}
	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		log.Printf("ParseSocketControlMessage error: %v", err)
		return
	}
	if len(msgs) == 0 {
		log.Printf("No SCM_RIGHTS message")
		return
	}
	fds, err := syscall.ParseUnixRights(&msgs[0])
	if err != nil {
		log.Printf("ParseUnixRights error: %v", err)
		return
	}
	if len(fds) == 0 {
		log.Printf("No FDs received")
		return
	}
	notifyFd := libseccomp.ScmpFd(fds[0])
	defer syscall.Close(int(notifyFd)) // Close notifyFd
	log.Printf("Received notifyFd from fds [%v]: %d", fds, notifyFd)

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

		syscallName, err := req.Data.Syscall.GetName()
		if err != nil {
			log.Printf("error getting syscall name from req data '%v': %v", req.Data.Syscall, err)
			return
		}
		resp := handleSyscall(syscallName, req)

		if err := libseccomp.NotifRespond(notifyFd, &resp); err != nil {
			log.Printf("NotifRespond error: %v", err)
			return
		}
	}
}

func handleSyscall(syscallName string, req *libseccomp.ScmpNotifReq) libseccomp.ScmpNotifResp {
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
