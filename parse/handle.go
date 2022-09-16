package parse

import (
	"github.com/fangjie-luoxi/git-hook/hook"
	"github.com/fangjie-luoxi/git-hook/log"
	"io/ioutil"
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	// 测试
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("read body failed, err:%v\n", err)
	}

	// 解析
	// 1.gitee :Header User-Agent: git-oschina-hook
	userAgents := r.Header["User-Agent"]
	if len(userAgents) > 0 && userAgents[0] == "git-oschina-hook" && len(body) > 0 {
		hook.GiteeHandle(w, body)
	}
}
