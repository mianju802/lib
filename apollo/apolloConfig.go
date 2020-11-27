package apollo

import (
	"github.com/patrickmn/go-cache"
	"github.com/shima-park/agollo"
	"log"
	"os"
)

type Apollo struct {
	localCache cache.Cache
}

func init() {
	err := agollo.InitWithDefaultConfigFile(
		agollo.WithLogger(agollo.NewLogger(agollo.LoggerWriter(os.Stdout))),
		agollo.PreloadNamespaces(),
		agollo.AutoFetchOnCacheMiss(),
		agollo.FailTolerantOnBackupExists(),
	)
	if err != nil {
		log.Fatalf("in initApollo call agollo.InitWithDefaultConfigFile error: %v", err)
	}
}

func (apollo *Apollo) ReadApolloConfig(nameSpace string) {
	var (
		errChan   = agollo.Start()
		watchChan = agollo.WatchNamespace(nameSpace, make(chan bool))
	)
	for _, val := range new(agollo.Configurations).Different(agollo.GetNameSpace(nameSpace)) {
		if val.Type != agollo.ChangeTypeDelete {
			apollo.localCache.SetDefault(val.Key, val.Value)
		}
	}
	for {
		select {
		case err := <-errChan:
			log.Fatalf("in initApollo agollo.Start error: %v", err)
		case rsp := <-watchChan:
			for _, val := range rsp.OldValue.Different(rsp.NewValue) {
				if val.Type != agollo.ChangeTypeDelete {
					apollo.localCache.SetDefault(val.Key, val.Value)
				}
			}
		}
	}
}
