## git的回调处理

### 概览

本项目主要用来处理各种回调,目前已完成git的Webhook处理


### 功能
- [x] 监听git的Webhook, 自动把项目部署在docker

### 使用说明

#### 安装git 

- 项目依赖git的Webhook，必须先安装好git

#### 开始使用

编写配置文件，先通过 ./git-hook -c ./config.yaml 启动服务端