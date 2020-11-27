package apollo

import (
	"github.com/patrickmn/go-cache"
	"github.com/shima-park/agollo"
	"log"
	"os"
)

type Apollo struct {
	LocalCache *cache.Cache
}

func init() {
	// 通过默认根目录下的app.properties初始化agollo
	err := agollo.InitWithDefaultConfigFile(
		agollo.WithLogger(agollo.NewLogger(agollo.LoggerWriter(os.Stdout))), // 打印日志信息
		agollo.PreloadNamespaces(),                                          // 预先加载的namespace列表，如果是通过配置启动，会在app.properties配置的基础上追加
		agollo.AutoFetchOnCacheMiss(),                                       // 在配置未找到时，去apollo的带缓存的获取配置接口，获取配置
		agollo.FailTolerantOnBackupExists(),                                 // 在连接apollo失败时，如果在配置的目录下存在.agollo备份配置，会读取备份在服务器无法连接的情况下
	)
	if err != nil {
		log.Fatalf("in initApollo call agollo.InitWithDefaultConfigFile error: %v", err)
	}
}

func (apollo *Apollo) ReadApolloConfig(nameSpace string) {
	for _, val := range new(agollo.Configurations).Different(agollo.GetNameSpace(nameSpace)) {
		if val.Type != agollo.ChangeTypeDelete {
			apollo.LocalCache.SetDefault(val.Key, val.Value)
		}
	}
	var (
		errChan            = agollo.Start()
		watchNameSpaceChan = agollo.WatchNamespace("consul", make(chan bool))
	)

	for {
		select {
		case err := <- errChan:
			log.Fatalf("err: %v",err)
		case watchData := <- watchNameSpaceChan:
			for _, val := range watchData.OldValue.Different(watchData.NewValue) {
				if val.Type != agollo.ChangeTypeDelete {
					apollo.LocalCache.SetDefault(val.Key, val.Value)
				}
			}
		}
	}
}
