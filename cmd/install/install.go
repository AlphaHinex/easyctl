package install

import (
	"github.com/spf13/cobra"
)

var offline bool
var offlineFilePath string
var serverListFile string

// ConfigFilePath 配置文件
var ConfigFilePath string

func init() {
	RootCmd.PersistentFlags().BoolVarP(&offline, "offline", "", false, "是否离线安装")
	RootCmd.PersistentFlags().StringVarP(&offlineFilePath, "offline-file", "", "", "离线文件")
	RootCmd.PersistentFlags().StringVarP(&serverListFile, "server-list", "", "", "服务器批量连接信息")
	RootCmd.PersistentFlags().StringVarP(&ConfigFilePath, "config", "c", "", "配置文件路径")

	RootCmd.AddCommand(keepaliveCmd)
	RootCmd.AddCommand(haproxyCmd)
	RootCmd.AddCommand(dockerCmd)
	RootCmd.AddCommand(dockerComposeCmd)
	RootCmd.AddCommand(harborCmd)
}

// RootCmd 安装根指令
var RootCmd = &cobra.Command{
	Use:   "install [OPTIONS] [flags]",
	Short: "安装指令集",
	Args:  cobra.MinimumNArgs(1),
}
