package gracefulServer

import (
	"net/http"
	"net"
	"time"
	"os"
	"syscall"
	"os/signal"
	"context"
	"os/exec"
)

const (
	envSignKey = "YUSHUAILIU-GRACEFUL"
	defaultTimeout = 10 * time.Second
)

type Graceful struct {
	server *http.Server
	listener net.Listener
	timeout time.Duration
	err error
	hooks map[string][]func()
}

func (graceful *Graceful)ListenAndServer(add string, handler http.Handler) error {
	graceful.server = &http.Server{Addr:add, Handler:handler}
	return graceful.run()
}

func (graceful *Graceful) getTimeout() time.Duration {
	if graceful.timeout == 0 {
		graceful.timeout = defaultTimeout
	}
	return graceful.timeout
}

func (graceful *Graceful)SetTimeout(timeout time.Duration) *Graceful {
	graceful.timeout = timeout
	return graceful
}

func (graceful *Graceful)run() (err error)  {

	if _, ok := syscall.Getenv(envSignKey); ok {
		fp := os.NewFile(3, "")
		if graceful.listener, err = net.FileListener(fp); err != nil {
			return
		}
	} else {
		if graceful.listener, err = net.Listen("tcp", graceful.server.Addr); err != nil {
			return
		}
	}

	startErrorChan := make(chan error)

	go func() {
		if err := graceful.server.Serve(graceful.listener); err != nil {
			startErrorChan <- err
		}
	}()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan)

	for {
		select {
		case startErrorChan := <-signalChan:
			switch startErrorChan {
			case syscall.SIGINT, syscall.SIGTERM:
				graceful.callHooks("beforeStop")
				signal.Stop(signalChan)
				graceful.Stop()
				graceful.callHooks("beforeStop")
				return graceful.err
			case syscall.SIGUSR2:
				graceful.callHooks("beforeReload")
				graceful.reload().Stop()
				graceful.callHooks("afterReload")
				return graceful.err
			}
		}
	}
}

func (graceful *Graceful)Stop() *Graceful {
	ctx, callback := context.WithTimeout(context.Background(), graceful.getTimeout())
	defer callback()

	if err := graceful.server.Shutdown(ctx); err != nil {
		graceful.err = err
	}

	return graceful
}

func (graceful *Graceful) reload() *Graceful {
	fp, err := graceful.listener.(*net.TCPListener).File()
	if err != nil {
		graceful.err = err
		return graceful
	}

	defer fp.Close()

	args := make([]string, 0)

	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	}

	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), envSignKey + "=true")
	cmd.ExtraFiles = []*os.File{fp}

	graceful.err = cmd.Start()
	return graceful
}

func (graceful *Graceful)RunServer(server *http.Server) error {
	graceful.server = server
	return graceful.run()
}

func (graceful *Graceful)AddBeforeStopHook(hooks... func()) error {
	return graceful.addHook("beforeStop", hooks)
}

func (graceful *Graceful)AddAfterStopHook(hooks... func()) error {
	return graceful.addHook("afterStop", hooks)
}

func (graceful *Graceful)AddBeforeReloadHook(hooks... func()) error {
	return graceful.addHook("beforeReload", hooks)
}

func (graceful *Graceful)AddAfterReloadHook(hooks... func()) error {
	return graceful.addHook("afterReload", hooks)
}

func (graceful *Graceful) addHook(hookMap string, hooks []func()) (err error) {
	if _, ok := graceful.hooks[hookMap]; !ok {
		graceful.hooks[hookMap] = make([]func(), 0)
	}

	for _, hook := range hooks{
		graceful.hooks[hookMap] = append(graceful.hooks[hookMap], hook)
	}
	return err
}

func (graceful *Graceful)callHooks(name string) {
	if _, ok := graceful.hooks[name]; !ok {
		return
	}

	for _, hook := range graceful.hooks[name] {
		hook()
	}
}

func RunServer(server *http.Server) (*Graceful, error) {
	graceful := NewGraceful()
	err := graceful.RunServer(server)
	return graceful, err
}

func ListenAndServer(add string, handler http.Handler) (*Graceful, error) {
	graceful := NewGraceful()
	err := graceful.ListenAndServer(add, handler)
	return graceful, err
}

func NewGraceful() *Graceful {
	in := &Graceful{
		hooks: map[string][]func(){},
	}
	return in
}