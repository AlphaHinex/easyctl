package install

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/weiliang-ms/easyctl/pkg/runner"
	"github.com/weiliang-ms/easyctl/pkg/util/command"
	"github.com/weiliang-ms/easyctl/pkg/util/constant"
	"github.com/weiliang-ms/easyctl/pkg/util/errors"
	"os"
	"sort"
	"testing"
)

// todo 合并测试用例

func TestRedisCluster(t *testing.T) {
	os.Setenv(constant.SshNoTimeout, "true")

	content := `
server:
  - host: 10.10.10.[1:3]
    username: root
    password: 123456
    port: 22
excludes:
  - 192.168.235.132
redis-cluster:
  password: "ddd"
  cluster-type: 0 # [0] 本地伪集群 ; [1] 三节点3分片2副本 ; [2] 6节点3分片2副本
  package: /root/redis-5.0.13.tar.gz
`
	assert.Nil(t, RedisCluster(command.OperationItem{B: []byte(content), Logger: logrus.New()}))
}

func TestParse(t *testing.T) {
	os.Setenv(constant.SshNoTimeout, "true")
	content := `
server:
  - host: 10.10.10.[1:3]
    username: root
    password: 123456
    port: 22
excludes:
  - 192.168.235.132
redis-cluster:
  password: "ddd"
  cluster-type: 0 # [0] 本地伪集群 ; [1] 三节点3分片2副本 ; [2] 6节点3分片2副本
  package: /root/redis-5.0.13.tar.gz
`
	var servers []runner.ServerInternal
	for i := 1; i < 4; i++ {
		servers = append(servers, runner.ServerInternal{
			Host:     fmt.Sprintf("10.10.10.%d", i),
			Port:     "22",
			Username: "root",
			Password: "123456",
		})
	}
	config := redisClusterConfig{
		Logger:        logrus.New(),
		ConfigContent: []byte(content),
	}

	actualServers := runner.InternelServersSlice{}
	err := config.Parse()
	actualServers = config.Servers
	sort.Sort(actualServers)

	assert.Nil(t, err)
	assert.Equal(t, "ddd", config.Password)
	assert.Equal(t, "/root/redis-5.0.13.tar.gz", config.Package)
	assert.Equal(t, servers, config.Servers)

	// test yaml.Unmarshal RedisClusterConfig err
	ddd := `
server:
  - host: 10.10.10.[1:3]
    username: root
    password: 123456
    port: 22
excludes:
  - 192.168.235.132
redis-cluster:
  password: "ddd"
  cluster-type: "0" # [0] 本地伪集群 ; [1] 三节点3分片2副本 ; [2] 6节点3分片2副本
  package: /root/redis-5.0.13.tar.gz
`
	ccc := redisClusterConfig{}
	ccc.ConfigContent = []byte(ddd)
	err = ccc.Parse()
	assert.Errorf(t, err, "Expected nil, but got: &yaml.TypeError{Errors:[]string{\"line 11: cannot unmarshal !!str `0` into install.RedisClusterType")

	// test yaml.Unmarshal RedisClusterConfig err
	aaa := `
server:
   host: 10.10.10.[1:3]
   username: root
   password: 123456
   port: 22
excludes:
 - 192.168.235.132
redis-cluster:
 password: "ddd"
 cluster-type: 0 # [0] 本地伪集群 ; [1] 三节点3分片2副本 ; [2] 6节点3分片2副本
 package: /root/redis-5.0.13.tar.gz
`
	bbb := redisClusterConfig{}
	bbb.ConfigContent = []byte(aaa)
	err = bbb.Parse()
	assert.Errorf(t, err, "[redis-cluster] 反序列化主机列表失败 -> yaml: unmarshal errors:\n  line 3: cannot unmarshal !!map into []runner.ServerExternal")
}

func TestDetect(t *testing.T) {
	os.Setenv(constant.SshNoTimeout, "true")

	var config redisClusterConfig
	var err error
	err = config.Detect()
	assert.Equal(t, errors.FileNotFoundErr(config.Package), err)

	f, _ := os.Create("1.tar.gz")
	config.Package = "1.tar.gz"
	config.Logger = logrus.New()

	config.CluterType = threeNodesThreeShards
	err = config.Detect()
	assert.Equal(t, errors.NumNotEqualErr("节点", 3, 0), err)

	config.CluterType = sixNodesThreeShards
	err = config.Detect()
	assert.Equal(t, errors.NumNotEqualErr("节点", 6, 0), err)

	// test local gcc exist
	config.CluterType = local
	config.Servers = []runner.ServerInternal{
		{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "123456",
		},
	}
	err = config.Detect()
	assert.NotNil(t, err)

	// test many node
	var servers []runner.ServerInternal
	for i := 1; i < 4; i++ {
		servers = append(servers, runner.ServerInternal{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "123456",
		})
	}
	config.Servers = servers
	config.CluterType = threeNodesThreeShards
	err = config.Detect()
	assert.EqualError(t, err, "10.10.10.1 依赖检测失败 -> runtime error: invalid memory address or nil pointer dereference")

	// ignore err return nil
	config.IgnoreErr = true
	assert.Nil(t, config.Detect())

	f.Close()
	os.Remove("1.tar.gz")
}

func TestPrune(t *testing.T) {
	os.Setenv(constant.SshNoTimeout, "true")

	// test local
	var config redisClusterConfig
	config.CluterType = local
	config.Servers = []runner.ServerInternal{
		{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "123456",
		},
	}
	config.Logger = logrus.New()
	assert.NotNil(t, config.Prune())

	// test many node
	var servers []runner.ServerInternal
	for i := 1; i < 4; i++ {
		servers = append(servers, runner.ServerInternal{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "123456",
		})
	}
	config.Servers = servers
	config.CluterType = threeNodesThreeShards
	err := config.Prune()
	assert.EqualError(t, err, "[10.10.10.1] 执行清理指令失败 runtime error: invalid memory address or nil pointer dereference")

	// ignore err return nil
	config.IgnoreErr = true
	assert.Nil(t, config.Prune())
}

func TestHandPackage(t *testing.T) {
	// test local
	var config redisClusterConfig
	config.CluterType = local
	config.Servers = []runner.ServerInternal{
		{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "123456",
		},
	}
	config.Logger = logrus.New()
	assert.Nil(t, config.HandPackage())

	os.Setenv(constant.SshNoTimeout, "true")

	// test many node
	var servers []runner.ServerInternal
	for i := 1; i < 4; i++ {
		servers = append(servers, runner.ServerInternal{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "123456",
		})
	}
	config.Servers = servers
	config.CluterType = threeNodesThreeShards
	err := config.HandPackage()
	assert.NotNil(t, err)

	// ignore err return nil
	config.IgnoreErr = true
	assert.Nil(t, config.HandPackage())
}

func TestRedisClusterConfig_Compile(t *testing.T) {
	// test local
	var config redisClusterConfig
	config.CluterType = local
	config.Servers = []runner.ServerInternal{
		{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "123456",
		},
	}
	config.Package = "1.tar.gz"
	assert.NotNil(t, config.Compile())
}

func TestRedisClusterConfig_Config(t *testing.T) {
	// test local
	var config redisClusterConfig
	config.Package = "1.tar.gz"
	config.CluterType = local
	assert.NotNil(t, config.Config())

	config.CluterType = threeNodesThreeShards
	assert.Nil(t, config.Config())

	config.CluterType = sixNodesThreeShards
	assert.Nil(t, config.Config())
}

func TestRedisClusterConfig_SetUpRuntime(t *testing.T) {

	// test local
	var config redisClusterConfig
	config.CluterType = local
	config.Servers = []runner.ServerInternal{
		{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "123456",
		},
	}
	config.Logger = logrus.New()
	err := config.SetUpRuntime()
	assert.NotNil(t, err)

	// test multi nodes
	config.Servers = []runner.ServerInternal{}
	config.CluterType = threeNodesThreeShards
	assert.Nil(t, config.SetUpRuntime())
}

func TestRedisClusterConfig_Boot(t *testing.T) {
	// test three noeds
	var config redisClusterConfig
	config.CluterType = threeNodesThreeShards
	config.Servers = []runner.ServerInternal{
		{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "222",
		},
		{
			Host:     "10.10.10.2",
			Port:     "22",
			Username: "root",
			Password: "222",
		},
		{
			Host:     "10.10.10.3",
			Port:     "22",
			Username: "root",
			Password: "222",
		},
	}
	config.Logger = logrus.New()
	err := config.Boot()
	assert.NotNil(t, err)
}

func TestRedisClusterConfig_CloseFirewall(t *testing.T) {
	var config redisClusterConfig
	config.CluterType = local
	config.Servers = []runner.ServerInternal{
		{
			Host:     "10.10.10.1",
			Port:     "22",
			Username: "root",
			Password: "123456",
		},
	}
	config.Logger = logrus.New()
	err := config.CloseFirewall()
	assert.NotEqual(t, nil, err)

	config.CluterType = threeNodesThreeShards
	config.Servers = []runner.ServerInternal{}
	config.Logger = logrus.New()
	err = config.CloseFirewall()
	assert.Equal(t, nil, err)

	config.CluterType = sixNodesThreeShards
	config.Servers = []runner.ServerInternal{}
	config.Logger = logrus.New()
	err = config.CloseFirewall()
	assert.Equal(t, nil, err)

	config.CluterType = 77
	config.Logger = logrus.New()
	err = config.CloseFirewall()
	assert.Equal(t, nil, err)
}

func TestRedisClusterConfig_SetValue(t *testing.T) {
	var config redisClusterConfig
	config.CluterType = local
	config.Logger = logrus.New()

	ports := []int{
		26379,
		36379,
		26380,
		36380,
		26381,
		36381,
		26382,
		36382,
		26383,
		36383,
		26384,
		36384,
	}
	// return 26379-26384 36379-36384
	err := config.SetValue()
	assert.Nil(t, err)
	assert.Equal(t, ports[:], config.PortsNeedOpen)

	// return 26379-26380 36379-36380
	config.CluterType = threeNodesThreeShards
	err = config.SetValue()
	assert.Nil(t, err)
	assert.Equal(t, ports[:4], config.PortsNeedOpen)

	// return 26379 36379
	config.CluterType = sixNodesThreeShards
	err = config.SetValue()
	assert.Nil(t, err)
	assert.Equal(t, ports[:2], config.PortsNeedOpen)

	// return 0
	config.CluterType = 77
	err = config.SetValue()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(config.PortsNeedOpen))
}

func TestRedisClusterConfig_Init(t *testing.T) {
	var config redisClusterConfig
	config.CluterType = threeNodesThreeShards
	config.Servers = []runner.ServerInternal{
		{Host: "10.10.10.10"},
	}
	config.Logger = logrus.New()
	config.EndpointList = []string{}
	err := config.Init()
	assert.NotNil(t, err)
}

func TestRedisClusterConfig_SetService(t *testing.T) {
	var config redisClusterConfig
	config.CluterType = threeNodesThreeShards
	config.Logger = logrus.New()
	assert.Equal(t, nil, config.SetService())
}

func TestRedisClusterConfig_Print(t *testing.T) {
	var config redisClusterConfig
	config.CluterType = threeNodesThreeShards
	config.Logger = logrus.New()
	config.Password = "redis@sss"
	config.EndpointList = []string{
		"10.10.10.1:26379",
		"10.10.10.1:26380",
		"10.10.10.2:26379",
		"10.10.10.2:26380",
		"10.10.10.3:26379",
		"10.10.10.3:26380",
	}

	assert.Equal(t, nil, config.Print())
}
