package runner

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"github.com/weiliang-ms/easyctl/pkg/util/command"
	"github.com/weiliang-ms/easyctl/pkg/util/constant"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"sort"
	"time"
)

func sftpConnect(user, password, host string, port string) (sftpClient *sftp.Client, err error) { //参数: 远程服务器用户名, 密码, ip, 端口
	auth := make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	var timeout time.Duration

	if os.Getenv(constant.SshNoTimeout) == "true" {
		timeout = 1
	} else {
		timeout = 5
	}
	clientConfig := &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: timeout * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr := host + ":" + port
	sshClient, err := ssh.Dial("tcp", addr, clientConfig) //连接ssh
	if err != nil {
		return nil, fmt.Errorf("连接ssh失败 %s", err)
	}

	if sftpClient, err = sftp.NewClient(sshClient); err != nil { //创建客户端
		return nil, fmt.Errorf("创建客户端失败 %s", err)
	}

	return sftpClient, nil
}

// RemoteRun 远程执行输出结果
func RemoteRun(b []byte, logger *logrus.Logger, cmd string) command.RunErr {

	results, err := GetResult(b, logger, cmd)
	if err != nil {
		return command.RunErr{Err: err}
	}
	var data [][]string

	for _, v := range results {
		if v.Err != nil {
			return command.RunErr{Err: v.Err, Msg: v.StdErrMsg}
		}
		data = append(data, []string{v.Host, v.Cmd, fmt.Sprintf("%d", v.Code), v.Status, v.StdOut, v.StdErrMsg})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IP ADDRESS", "cmd", "exit code", "result", "output", "exception"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	//table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return command.RunErr{}
}

// GetResult 远程执行，获取结果
func GetResult(b []byte, logger *logrus.Logger, cmd string) ([]ShellResult, error) {

	servers, err := ParseServerList(b, logger)
	if err != nil {
		return []ShellResult{}, err
	}

	// 组装执行器,执行命令
	executor := ExecutorInternal{Servers: servers, Script: cmd, Logger: logger}
	ch := executor.ParallelRun()

	// 打包执行结果
	var results []ShellResult

	for re := range ch {
		var result ShellResult
		_ = mapstructure.Decode(re, &result)
		results = append(results, result)
	}

	sort.Sort(ShellResultSlice(results))

	return results, nil
}
