package hook

import "testing"

func TestGitHook(t *testing.T) {
	hook := GitHook{
		ProjectName: "test-api",
		HookName:    "",
		Url:         "git@gitee.com:fangjie-luoxi/test-api.git",
		Time:        "",
		UserName:    "",
		Message:     "",
		folderName:  "",
		folderPath:  "",
		portList:    []string{"9090", "9090"},
	}
	hook.dockerRun()
}

func TestClearUp(t *testing.T) {
	h := GitHook{}
	h.clearUp()
}
