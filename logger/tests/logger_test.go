package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/deis/deis/tests/dockercli"
	"github.com/deis/deis/tests/etcdutils"
	"github.com/deis/deis/tests/utils"
)

func TestLogger(t *testing.T) {
	var err error
	tag, etcdPort := utils.BuildTag(), utils.RandomPort()
	imageName := utils.ImagePrefix() + "logger" + ":" + tag

	//start etcd container
	etcdName := "deis-etcd-" + tag
	cli, stdout, stdoutPipe := dockercli.NewClient()
	dockercli.RunTestEtcd(t, etcdName, etcdPort)
	defer cli.CmdRm("-f", etcdName)

	host, port := utils.HostAddress(), utils.RandomPort()
	fmt.Printf("--- Run %s at %s:%s\n", imageName, host, port)
	name := "deis-logger-" + tag
	defer cli.CmdRm("-f", name)
	go func() {
		_ = cli.CmdRm("-f", name)
		err = dockercli.RunContainer(cli,
			"--name", name,
			"--rm",
			"-p", port+":514/udp",
			imageName,
			"--enable-publish",
			"--log-port="+port,
			"--publish-host="+host,
			"--publish-port="+etcdPort)
	}()
	dockercli.PrintToStdout(t, stdout, stdoutPipe, "deis-logger running")
	if err != nil {
		t.Fatal(err)
	}
	// FIXME: Wait until etcd keys are published
	time.Sleep(5000 * time.Millisecond)
	dockercli.DeisServiceTest(t, name, port, "udp")
	etcdutils.VerifyEtcdValue(t, "/deis/logs/host", host, etcdPort)
	etcdutils.VerifyEtcdValue(t, "/deis/logs/port", port, etcdPort)
}
