package hook

import (
	"github.com/fangjie-luoxi/git-hook/config"
	"github.com/fangjie-luoxi/git-hook/log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GitHook hook 回调
type GitHook struct {
	ProjectName string // 项目名称
	HookName    string // hook名称
	Url         string // 项目地址
	Time        string // 提交代码时间
	UserName    string // 用户名称
	Message     string // 提交消息
	Branch      string // 分支
	CommitId    string // 推送后分支的 commit_id

	folderName    string // 构建项目文件夹
	folderPath    string // 构建项目文件夹路径
	imageName     string // 镜像名称 项目名称:分支名称
	containerName string // 容器名称 项目名称.分支名称

	dockerFile string   // dockerfile
	portList   []string // 暴露的端口
}

// Construct 构建项目
func (h *GitHook) Construct() {
	h.folderName = "temp/" + h.ProjectName + "/"
	h.folderPath = "./temp/" + h.ProjectName
	h.imageName = h.ProjectName + ":" + h.Branch
	h.containerName = h.ProjectName + "." + h.Branch

	if config.Config.Git.Branch != "" && h.Branch != config.Config.Git.Branch {
		log.Info("跳过分支")
		return
	}

	err := os.RemoveAll(h.folderPath)
	if err != nil {
		log.Error("删除文件夹失败:", err.Error())
	}

	log.Info("正在构建项目:", h.ProjectName)
	err = h.clone()
	if err != nil {
		return
	}

	// todo 判断是否有脚本 脚本目录 /script/hook.sh
	// 执行用户脚本
	// 执行默认命令
	h.dealWith()
	log.Info("项目构建完成:", h.ProjectName)
}

// 克隆代码
func (h *GitHook) clone() error {
	log.Info("正在拉取项目:", h.Url)
	err := exec.Command("git", "clone", "-b", h.Branch, h.Url, h.folderName).Run()
	if err != nil {
		log.Errorf("git clone fail:%v\n", err)
		return err
	}
	log.Info("拉取完成")
	return nil
}

// 默认处理
func (h *GitHook) dealWith() {
	// 检查是否有dockerfile
	if !h.checkDockerfile() {
		return
	}
	// 停止并删除正在运行的docker
	h.delDocker()
	// dockerBuild
	err := h.dockerBuild()
	if err != nil {
		log.Error("dockerBuild err:", err.Error())
		return
	}
	err = h.dockerRun()
	if err != nil {
		log.Error("dockerRun err:", err.Error())
		return
	}
	h.clearUp()
}

// clearUp 清理docker
func (h *GitHook) clearUp() {
	// docker rmi $(docker images -a | grep none | awk '{print $3}')
	// docker images | grep none | awk '{print $3}' | xargs docker rmi
	log.Info("正在清理none镜像")
	// cmd := exec.Command("docker", "rmi", "$(docker images -a | grep none | awk {print $3})")
	// cmd := exec.Command("docker", "images", "|", "grep", "none", "|", "awk", "{print $3}", "|", "xargs", "docker", "rmi")
	_ = exec.Command("docker", "image", "prune", "--force").Run()
	log.Info("完成清理none镜像")
}

func (h *GitHook) delDocker() {
	// docker stop container
	// docker rm container
	// docker rmi image
	log.Info("正在删除容器和镜像")
	_ = exec.Command("docker", "stop", h.containerName).Run()
	_ = exec.Command("docker", "rm", h.containerName).Run()
	_ = exec.Command("docker", "rmi", h.imageName).Run()
	log.Info("完成删除容器和镜像")
}

// 默认处理
func (h *GitHook) checkDockerfile() bool {
	// 检查是否有dockerfile
	dockerFile := "Dockerfile"
	fp := filepath.Join(h.folderPath, dockerFile)
	if checkNotExist(fp) {
		log.Error("Dockerfile 不存在")
		return false
	}
	// 根据分支判断使用哪个dockerfile
	if h.Branch != "master" {
		dockerFileBranch := "Dockerfile" + "-" + h.Branch
		fpBranch := filepath.Join(h.folderPath, dockerFileBranch)
		if !checkNotExist(fpBranch) {
			dockerFile = dockerFileBranch
			fp = fpBranch
		}
	}
	h.dockerFile = dockerFile
	bytes, err := os.ReadFile(fp)
	if err != nil {
		return false
	}
	h.getPortList(string(bytes))
	return true
}

func (h *GitHook) dockerBuild() error {
	// docker build -f ./folderName/Dockerfile-dev -t jxt-app-api .
	cmd := exec.Command("docker", "build", "-f", h.dockerFile, "-t", h.imageName, ".")
	cmd.Dir = "./" + h.folderName
	log.Info("正在执行 dockerBuild:", cmd.String())
	bytes, err := cmd.Output()
	if err != nil {
		log.Error("错误信息:", string(bytes))
		return err
	}
	log.Info("执行 dockerBuild 完成")
	return nil
}

func (h *GitHook) dockerRun() error {
	// docker run --restart=always -d --name jxt-store-api jxt-stock-api
	var cmd *exec.Cmd
	if len(h.portList) == 0 {
		cmd = exec.Command("docker", "run", "-d", "--restart=always", "--network=host", "--name", h.containerName, h.imageName)
	} else {
		cmd = exec.Command("docker", "run", "-d", "--restart=always")
		for _, port := range h.portList {
			cmd.Args = append(cmd.Args, "-p")
			cmd.Args = append(cmd.Args, port)
		}
		cmd.Args = append(cmd.Args, "--name")
		cmd.Args = append(cmd.Args, h.containerName)
		cmd.Args = append(cmd.Args, h.imageName)
	}
	log.Info("正在执行 dockerRun: ", cmd.String())
	cmd.Dir = h.folderPath
	bytes, err := cmd.Output()
	if err != nil {
		log.Error("错误信息:", string(bytes))
		return err
	}
	log.Info("执行 dockerRun 完成")
	return nil
}

// checkNotExist 检查文件是否存在
func checkNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

// 获取dockerfile的端口
func (h *GitHook) getPortList(file string) {
	compile, _ := regexp.Compile("EXPOSE (.*)")
	stringList := compile.FindAllString(file, -1)
	if len(stringList) == 0 {
		//EXPOSE 8080
		return
	}
	numRgp, _ := regexp.Compile("\\d+")
	// 获取暴露的端口
	var portList []string
	for _, exposeStr := range stringList {
		ports := strings.Split(exposeStr, " ")
		for _, port := range ports {
			portByte := numRgp.Find([]byte(port))
			if portByte == nil {
				continue
			}
			portList = append(portList, string(portByte)+":"+strings.TrimSpace(port))
		}
	}
	h.portList = portList
}
