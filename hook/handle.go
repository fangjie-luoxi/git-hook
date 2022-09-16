package hook

import (
	"github.com/fangjie-luoxi/git-hook/log"
	"github.com/tidwall/gjson"
	"net/http"
	"path"
	"sync"
)

var lock sync.Mutex

// GiteeHandle gitee回调
func GiteeHandle(w http.ResponseWriter, body []byte) {
	hook := GitHook{
		ProjectName: gjson.Get(string(body), "repository.path").String(),
		HookName:    gjson.Get(string(body), "hook_name").String(),
		Url:         gjson.Get(string(body), "repository.git_ssh_url").String(),
		Time:        gjson.Get(string(body), "head_commit.author.time").String(),
		UserName:    gjson.Get(string(body), "head_commit.author.name").String(),
		Message:     gjson.Get(string(body), "head_commit.message").String(),
		CommitId:    gjson.Get(string(body), "after").String(),
		Branch:      path.Base(gjson.Get(string(body), "ref").String()),
	}
	if hook.Url == "" {
		log.Error("GiteeHook Url is null")
		return
	}
	go func() {
		log.Info("等待锁")
		lock.Lock()
		defer lock.Unlock()
		log.Info("获得锁")
		hook.Construct()
	}()
}
